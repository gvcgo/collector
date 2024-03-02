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
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
)

const (
	PythonVersionFileName string = "python.version.json"
)

/*
	Available python versions from miniconda.

https://anaconda.org/conda-forge/python/files
*/
type Python struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewPython(cnf *confs.CollectorConf) (p *Python) {
	p = &Python{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://anaconda.org/conda-forge/python/files",
	}
	if confs.EnableProxyOrNot() {
		pxy := p.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		p.fetcher.Proxy = pxy
	}
	return
}

func (p *Python) getDoc() {
	p.fetcher.SetUrl(p.homepage)
	p.fetcher.Timeout = 30 * time.Second
	if resp, sCode := p.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
		var err error
		p.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if p.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", p.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Println(sCode)
	}
}

func (p *Python) FetchAll() {
	p.getDoc()

	if p.doc != nil {
		p.doc.Find("ul#Version").Find("a").Each(func(i int, s *goquery.Selection) {
			vName := strings.TrimSpace(s.Text())
			// fmt.Println(vName)
			if strings.ToLower(vName) == "all" {
				return
			}
			if _, ok := p.versions[vName]; !ok {
				p.versions[vName] = []*VFile{
					{
						Extra: "please use conda to install.",
					},
				}
			}
		})
	}
}

func (p *Python) Upload() {
	if len(p.versions) > 0 {
		fPath := filepath.Join(p.cnf.DirPath(), PythonVersionFileName)
		if content, err := json.MarshalIndent(p.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			p.uploader.Upload(fPath)
		}
	}
}
