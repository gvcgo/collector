package versions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
)

const (
	DlangVersionFileName string = "dlang.version.json"
	DlangUrl             string = "https://downloads.dlang.org/releases/2.x/"
)

/*
Dlang versions.

https://downloads.dlang.org/releases/
*/
type Dlang struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewDlang(cnf *confs.CollectorConf) (d *Dlang) {
	d = &Dlang{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://downloads.dlang.org",
	}
	if confs.EnableProxyOrNot() {
		pxy := d.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		d.fetcher.Proxy = pxy
	}
	return
}

func (d *Dlang) getDoc() {
	d.fetcher.SetUrl(DlangUrl)
	d.fetcher.Timeout = 180 * time.Second
	if resp, sCode := d.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
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

func (d *Dlang) FetchOne(vName, vHref string) {
	u, err := url.JoinPath(d.homepage, vHref)
	if err != nil {
		return
	}
	d.fetcher.SetUrl(u)
	d.fetcher.Timeout = 180 * time.Second
	fmt.Println("** ", u)
	if resp, sCode := d.fetcher.GetString(); resp != "" && sCode == 200 {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil || doc == nil {
			return
		}
		doc.Find("div#content").Find("li").Find("a").Each(func(_ int, s *goquery.Selection) {
			href := s.AttrOr("href", "")
			ver := &VFile{}
			if strings.HasSuffix(href, ".windows.zip") {
				ver.Url, _ = url.JoinPath(d.homepage, href)
				ver.Arch = "amd64"
				ver.Os = "windows"
			} else if strings.HasSuffix(href, ".osx.zip") {
				ver.Url, _ = url.JoinPath(d.homepage, href)
				ver.Arch = "amd64"
				ver.Os = "darwin"
			} else if strings.HasSuffix(href, ".linux.zip") {
				ver.Url, _ = url.JoinPath(d.homepage, href)
				ver.Arch = "amd64"
				ver.Os = "linux"
			}
			if ver.Url != "" {
				if vlist, ok := d.versions[vName]; !ok || vlist == nil {
					d.versions[vName] = []*VFile{}
				}
				d.versions[vName] = append(d.versions[vName], ver)
			}
		})
	}
}

func (d *Dlang) FetchAll() {
	if d.doc == nil {
		d.getDoc()
	}

	if d.doc != nil {
		// //div[@id="content"]//li/a/@href
		d.doc.Find("div#content").Find("li").Find("a").Each(func(_ int, s *goquery.Selection) {
			vName := s.Text()
			vHref := s.AttrOr("href", "")
			if vHref == "" {
				return
			}
			if strings.Count(vName, ".") < 2 {
				return
			}

			// higher than 2.065.0
			vList := strings.Split(vName, ".")
			vMinor, _ := strconv.Atoi(vList[1])
			if vMinor < 65 {
				return
			}
			d.FetchOne(vName, vHref)
		})
	}
}

func (d *Dlang) Upload() {
	if len(d.versions) > 0 {
		fPath := filepath.Join(d.cnf.DirPath(), DlangVersionFileName)
		if content, err := json.MarshalIndent(d.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			d.uploader.Upload(fPath)
		}
	}
}
