package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

const (
	GradleFileName string = "gradle.version.json"
	GradleSumUrl   string = "https://gradle.org/release-checksums/"
)

/*
https://gradle.org/releases/
https://gradle.org/release-checksums/
*/
type Gradle struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
	sha      map[string]string
}

func NewGradle(cnf *confs.CollectorConf) (g *Gradle) {
	g = &Gradle{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://gradle.org/releases/",
		sha:      map[string]string{},
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

func (g *Gradle) getDoc() {
	// get sum info.
	g.fetcher.SetUrl(GradleSumUrl)
	g.fetcher.Timeout = 30 * time.Second
	if resp := g.fetcher.Get(); resp != nil {
		g.doc, _ = goquery.NewDocumentFromReader(resp.RawBody())
	}
	if g.doc != nil {
		g.doc.Find("h3.u-text-with-icon").Each(func(i int, s *goquery.Selection) {
			version := s.Find("a").AttrOr("id", "")
			if version == "" {
				return
			}
			shaCode := s.Next().Find("li").Eq(0).Find("code").Text()
			if shaCode != "" {
				g.sha[version] = shaCode
			}
		})
	}

	// get version list info.
	g.fetcher.SetUrl(g.homepage)
	g.fetcher.Timeout = 30 * time.Second
	if resp := g.fetcher.Get(); resp != nil {
		var err error
		g.doc, err = goquery.NewDocumentFromReader(resp.RawBody())
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if g.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", g.fetcher.Url))
			os.Exit(1)
		}
	}
}

func (g *Gradle) getSum(version string) (code string) {
	if len(g.sha) == 0 {
		g.getDoc()
	}
	for k, v := range g.sha {
		if strings.ReplaceAll(k, "v", "") == version {
			return v
		}
	}
	return
}

func (g *Gradle) GetVersions() {
	g.getDoc()

	g.doc.Find("div.indent").Each(func(i int, s *goquery.Selection) {
		aLabel := s.Find("li").Eq(0).Find("a").Eq(1)
		ver := &VFile{}
		ver.Url = aLabel.AttrOr("href", "")
		vName := aLabel.AttrOr("data-version", "")
		if ver.Url == "" || vName == "" {
			return
		}
		ver.Sum = strings.TrimSpace(g.getSum(vName))
		if ver.Sum != "" {
			ver.SumType = " SHA256"
		}

		ver.Arch = "any"
		ver.Os = "any"

		if vlist, ok := g.versions[vName]; !ok || vlist == nil {
			g.versions[vName] = []*VFile{}
		}
		g.versions[vName] = append(g.versions[vName], ver)
	})
}

func (g *Gradle) FetchAll() {
	g.GetVersions()
}

func (g *Gradle) Upload() {
	if len(g.versions) > 0 {
		fPath := filepath.Join(g.cnf.DirPath(), GradleFileName)
		if content, err := json.MarshalIndent(g.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			g.uploader.Upload(fPath)
		}
	}
}
