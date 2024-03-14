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
	"github.com/gvcgo/collector/pkgs/utils"
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
		fmt.Printf("Failed: %s, code: %d\n", d.homepage, sCode)
	}
}

func (d *DotNet) fetchVersion(vUrl, vStr string) {
	d.fetcher.SetUrl(vUrl)
	d.fetcher.Timeout = 180 * time.Second
	var doc *goquery.Document
	if resp, sCode := d.fetcher.GetString(); resp != "" && sCode == 200 {
		var err error
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", d.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Printf("Failed: %s, code: %d\n", d.homepage, sCode)
		return
	}

	link := doc.Find("a#directLink").AttrOr("href", "")
	sha512Str := doc.Find("input#checksum").AttrOr("value", "")
	sumType := "SHA512"
	if sha512Str == "" {
		sumType = sha512Str
	}
	if _, ok := d.versions[vStr]; !ok {
		d.versions[vStr] = []*VFile{}
	}
	d.versions[vStr] = append(d.versions[vStr], &VFile{
		Url:     link,
		Sum:     sha512Str,
		SumType: sumType,
		Arch:    utils.ParseArch(vUrl),
		Os:      utils.ParsePlatform(vUrl),
	})
}

func filterDotNetSDKByUrl(vUrl string) bool {
	if vUrl == "" {
		return false
	}
	excludeList := []string{
		"alpine",
		"installer",
		"winget",
		"arm32",
		"install",
		"scripts",
	}
	for _, ee := range excludeList {
		if strings.Contains(vUrl, ee) {
			return false
		}
	}
	return true
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
		d.doc.Find("div.download-panel").Find("div").Find("table").Each(func(i int, ss *goquery.Selection) {
			vInfo := ss.Find("caption").AttrOr("id", "")
			vList := strings.Split(vInfo, "-sdk-")
			if len(vList) < 2 {
				return
			}
			vName := vList[len(vList)-1]
			ss.Find("a").Each(func(i int, sa *goquery.Selection) {
				uu := sa.AttrOr("href", "")
				if filterDotNetSDKByUrl(uu) {
					// fmt.Println(uu)
					if !strings.Contains(uu, d.host) {
						uu, _ = url.JoinPath(d.host, uu)
					}
					d.fetchVersion(uu, vName)
				}
			})
		})

		// vInfo := d.doc.Find("div.download-panel").Find("div").Find("table").Eq(0).Find("caption").AttrOr("id", "")
		// vList := strings.Split(vInfo, "-sdk-")
		// vStr := vList[len(vList)-1]
		// // //div[@class="download-panel"]//div//table[1]
		// d.doc.Find("div.download-panel").Find("div").Find("table").Eq(0).Find("a").Each(func(_ int, s *goquery.Selection) {
		// 	uu := s.AttrOr("href", "")
		// 	if filterDotNetSDKByUrl(uu) {
		// 		if !strings.Contains(uu, d.host) {
		// 			uu, _ = url.JoinPath(d.host, uu)
		// 		}
		// 		d.fetchVersion(uu, vStr)
		// 	}
		// })
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
