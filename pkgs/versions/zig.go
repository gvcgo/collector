package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
	"github.com/gvcgo/proxy-collector/pkgs/utils"
)

const (
	ZigVersionFileName string = "zig.version.json"
)

/*
https://ziglang.org/download/
*/
type Zig struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
	l        *sync.Mutex
}

func NewZig(cnf *confs.CollectorConf) (z *Zig) {
	z = &Zig{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://ziglang.org/download/",
		l:        &sync.Mutex{},
	}
	return
}

func (z *Zig) getDoc() {
	z.fetcher.SetUrl(z.homepage)
	z.fetcher.Timeout = 60 * time.Second
	if resp := z.fetcher.Get(); resp != nil {
		var err error
		z.doc, err = goquery.NewDocumentFromReader(resp.RawBody())
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if z.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", z.fetcher.Url))
			os.Exit(1)
		}
	}
}

func (z *Zig) GetVersions() {
	if z.doc == nil {
		z.getDoc()
	}
	if z.doc != nil {
		z.doc.Find("h2").Each(func(i int, s *goquery.Selection) {
			vName := strings.TrimSpace(s.Text())
			if vName == "master" {
				return
			}
			fmt.Println(vName)
			z.doc.Find("table").Eq(i - 1).Find("tr").Each(func(_ int, ss *goquery.Selection) {
				thStr := strings.ToLower(strings.TrimSpace(ss.Find("th").Text()))
				tdList := ss.Find("td").Nodes
				if thStr == "os" || len(tdList) < 4 {
					return
				}
				ver := &VFile{}
				archStr := utils.ParseArch(ss.Find("td").Eq(0).Text())
				if archStr == "" {
					return
				}
				ver.Arch = archStr
				ver.Url = ss.Find("td").Eq(1).Find("a").AttrOr("href", "")
				ver.Os = utils.ParsePlatform(ver.Url)
				// fmt.Println(ver.Os, ver.Arch, ver.Url)
				z.l.Lock()
				if vlist, ok := z.versions[vName]; !ok || vlist == nil {
					z.versions[vName] = []*VFile{}
				}
				z.versions[vName] = append(z.versions[vName], ver)
				z.l.Unlock()
			})

		})
	}
}

func (z *Zig) FetchAll() {
	z.GetVersions()
}

func (z *Zig) Upload() {
	if len(z.versions) > 0 {
		fPath := filepath.Join(z.cnf.DirPath(), ZigVersionFileName)
		if content, err := json.MarshalIndent(z.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			z.uploader.Upload(fPath)
		}
	}
}
