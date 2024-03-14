package versions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
)

const (
	DotNetVersionFileName string = "dotnet.version.json"
)

/*
.Net versions.

https://dotnet.microsoft.com/en-us/download/dotnet
*/
type DotNet struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	host     string
	doc      *goquery.Document
}

func NewDotNet(cnf *confs.CollectorConf) (d *DotNet) {
	d = &DotNet{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://dotnet.microsoft.com/en-us/download/dotnet",
		host:     "https://dotnet.microsoft.com",
	}
	if confs.EnableProxyOrNot() {
		pxy := d.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		d.fetcher.Proxy = pxy
	}
	return d
}

func (d *DotNet) getDoc() {
	d.fetcher.SetUrl(d.homepage)
	d.fetcher.Timeout = 180 * time.Second
	if resp, sCode := d.fetcher.GetString(); resp != "" && sCode == 200 {
		var err error
		d.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if d.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", d.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Printf("Failed: %s, code: %d", d.homepage, sCode)
	}
}

func (d *DotNet) FetchAll() {
	d.getDoc()
	if d.doc == nil {
		return
	}
	supportedVersionUrls := []string{}
	d.doc.Find("div#supported-versions-table").Find("table").Find("a").Each(func(_ int, s *goquery.Selection) {
		u := s.AttrOr("href", "")
		if u != "" && !strings.Contains(u, d.host) {
			u, _ = url.JoinPath(d.host, u)
		}
		supportedVersionUrls = append(supportedVersionUrls, u)
	})
	for _, u := range supportedVersionUrls {
		d.doc = nil
		d.homepage = u
		d.getDoc()
		if d.doc == nil {
			continue
		}
		// //div[@class="download-panel"]//div//table[1]
		d.doc.Find("div.download-panel").Find("div").Find("table").Eq(0).Find("a").Each(func(_ int, s *goquery.Selection) {
			uu := s.AttrOr("href", "")
			fmt.Println(uu)
		})
	}
}

func (d *DotNet) Upload() {
	if len(d.versions) > 0 {
		fPath := filepath.Join(d.cnf.DirPath(), DotNetVersionFileName)
		if content, err := json.MarshalIndent(d.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			d.uploader.Upload(fPath)
		}
	}
}
