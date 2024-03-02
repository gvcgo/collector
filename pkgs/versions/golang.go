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
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
)

const (
	GoVersionFileName string = "golang.version.json"
)

/*
https://golang.google.cn/dl/
https://go.dev/dl/
*/
type Golang struct {
	cnf       *confs.CollectorConf
	uploader  *upload.Uploader
	versions  Versions
	fetcher   *request.Fetcher
	homepage  string
	doc       *goquery.Document
	parsedUrl *url.URL
}

func NewGolang(cnf *confs.CollectorConf) (g *Golang) {
	g = &Golang{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://go.dev/dl/",
	}
	if UseCNSource() {
		g.homepage = "https://golang.google.cn/dl/"
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

func (g *Golang) getDoc() {
	g.parsedUrl, _ = url.Parse(g.homepage)
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

func (g *Golang) hasUnstableVersions() bool {
	if g.doc == nil {
		g.getDoc()
	}
	label := g.doc.Find("#unstable")
	if label == nil {
		return false
	}
	return g.doc.Find("#unstable").Length() > 0
}

func (g *Golang) findPackages(table *goquery.Selection, vTagName string) {
	sType := strings.TrimSuffix(table.Find("thead").Find("th").Last().Text(), " Checksum")
	table.Find("tr").Not(".first").Each(func(j int, tr *goquery.Selection) {
		td := tr.Find("td")
		href := td.Eq(0).Find("a").AttrOr("href", "")
		if strings.HasPrefix(href, "/") {
			// relative paths
			href = fmt.Sprintf("%s://%s%s", g.parsedUrl.Scheme, g.parsedUrl.Host, href)
		}
		if k := strings.ToLower(strings.ToLower(td.Eq(1).Text())); k != "archive" {
			return
		}

		ver := &VFile{
			Url:     href,
			Arch:    utils.MapArchAndOS(td.Eq(3).Text()),
			Os:      utils.MapArchAndOS(td.Eq(2).Text()),
			Sum:     td.Eq(5).Text(),
			SumType: sType,
		}
		if ver.Arch == "bootstrap" && vTagName == "1" {
			ver.Os = ver.Arch
		}
		if vfiles, ok := g.versions[vTagName]; !ok || vfiles == nil {
			g.versions[vTagName] = []*VFile{}
		}
		g.versions[vTagName] = append(g.versions[vTagName], ver)
	})
}

func (g *Golang) GetUnstableVersions() {
	g.doc.Find("#unstable").NextUntil("#archive").Each(func(i int, div *goquery.Selection) {
		vname, ok := div.Attr("id")
		if !ok {
			return
		}
		vname = strings.TrimPrefix(vname, "go")
		g.findPackages(div.Find("table").First(), vname)
	})
}

func (g *Golang) GetStableVersions() {
	var divs *goquery.Selection
	if g.hasUnstableVersions() {
		divs = g.doc.Find("#stable").NextUntil("#unstable")
	} else {
		divs = g.doc.Find("#stable").NextUntil("#archive")
	}
	divs.Each(func(i int, div *goquery.Selection) {
		vname, ok := div.Attr("id")
		if !ok {
			return
		}
		vname = strings.TrimPrefix(vname, "go")
		g.findPackages(div.Find("table").First(), vname)
	})
}

func (g *Golang) GetArchivedVersions() {
	g.doc.Find("#archive").Find("div.toggle").Each(func(i int, div *goquery.Selection) {
		vname, ok := div.Attr("id")
		if !ok {
			return
		}
		vname = strings.TrimPrefix(vname, "go")
		g.findPackages(div.Find("table").First(), vname)
	})
}

func (g *Golang) FetchAll() {
	if g.doc == nil {
		g.getDoc()
	}
	g.GetStableVersions()
	g.GetArchivedVersions()
	g.GetUnstableVersions()
}

func (g *Golang) Upload() {
	if len(g.versions) > 0 {
		fPath := filepath.Join(g.cnf.DirPath(), GoVersionFileName)
		if content, err := json.MarshalIndent(g.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			g.uploader.Upload(fPath)
		}
	}
}
