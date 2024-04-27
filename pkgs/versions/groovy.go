package versions

import (
	"encoding/json"
	"fmt"
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
	GroovyVersionFileName string = "groovy.version.json"
	GroovyUrlPattern      string = "https://archive.apache.org/dist/groovy/%s/distribution/"
)

/*
Groovy versions.

https://groovy.apache.org/download.html
https://archive.apache.org/dist/groovy/
*/
type Groovy struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewGroovy(cnf *confs.CollectorConf) (g *Groovy) {
	g = &Groovy{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://archive.apache.org/dist/groovy/",
	}
	if confs.EnableProxyOrNot() {
		pxy := g.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		g.fetcher.Proxy = pxy
	}
	return
}

func (g *Groovy) getDoc() {
	g.fetcher.SetUrl(g.homepage)
	g.fetcher.Timeout = 180 * time.Second
	if resp, sCode := g.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
		var err error
		g.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if g.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", g.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Printf("Failed: %s, code: %d", g.homepage, sCode)
	}
}

func (g *Groovy) getSha256(sdkUrl string) (resp string) {
	g.fetcher.SetUrl(sdkUrl + ".sha256")
	g.fetcher.Timeout = 180 * time.Second
	var code int
	resp, code = g.fetcher.GetString()
	if code != 200 {
		g.fetcher.SetUrl(sdkUrl + ".md5")
		resp, _ = g.fetcher.GetString()
	}
	return
}

func (g *Groovy) FetchOne(href string) {
	vName := strings.Trim(href, "/")
	var (
		sdkUrl string
		sha256 string
	)
	distUrl := fmt.Sprintf(GroovyUrlPattern, vName)
	g.fetcher.SetUrl(distUrl)
	g.fetcher.Timeout = 180 * time.Second
	if resp, sCode := g.fetcher.GetString(); resp != "" && sCode == 200 {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(resp))
		if doc != nil {
			doc.Find("a").Each(func(_ int, s *goquery.Selection) {
				h := s.AttrOr("href", "")
				if !strings.Contains(h, "-sdk-") {
					return
				}
				if strings.HasSuffix(h, ".zip") {
					sdkUrl = fmt.Sprintf("%s/%s", strings.Trim(distUrl, "/"), strings.Trim(h, "/"))
					sha256 = g.getSha256(sdkUrl)
				}
			})
			if sdkUrl != "" && sha256 != "" {
				ver := &VFile{}
				ver.Url = sdkUrl
				ver.Arch = "any"
				ver.Os = "any"
				ver.Sum = strings.TrimSpace(sha256)
				ver.SumType = "sha256"
				if strings.Contains(ver.Sum, "apache-groovy-sdk") {
					sList := strings.Split(ver.Sum, "apache-groovy-sdk")
					ver.Sum = strings.TrimSpace(sList[0])
					ver.SumType = "md5"
				}
				if vlist, ok := g.versions[vName]; !ok || vlist == nil {
					g.versions[vName] = []*VFile{}
				}
				g.versions[vName] = append(g.versions[vName], ver)
			}
		}
	}
}

func (g *Groovy) FetchAll() {
	g.getDoc()
	if g.doc == nil {
		return
	}
	g.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if versionRegexp.FindString(href) != "" {
			g.FetchOne(href)
		}
	})
}

func (g *Groovy) Upload() {
	if len(g.versions) > 0 {
		fPath := filepath.Join(g.cnf.DirPath(), GroovyVersionFileName)
		if content, err := json.MarshalIndent(g.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			g.uploader.Upload(fPath)
		}
	}
}
