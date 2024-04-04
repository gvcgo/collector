package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
)

const (
	PyPyVersionFileName string = "pypy.version.json"
)

/*
Available PyPy versions from miniconda.

https://anaconda.org/conda-forge/pypy/files
*/

type PyPy struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewPyPy(cnf *confs.CollectorConf) (p *PyPy) {
	p = &PyPy{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://anaconda.org/conda-forge/pypy/files",
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

func (p *PyPy) getDoc() {
	p.fetcher.SetUrl(p.homepage)
	p.fetcher.Timeout = 180 * time.Second
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
		fmt.Printf("Failed: %s, code: %d", p.homepage, sCode)
	}
}

func (p *PyPy) formatVersionName(vName string) string {
	majorStr := vName[0:1]
	return strings.Replace(vName, majorStr, majorStr+".", 1)
}

var pypyRegExp = regexp.MustCompile(`pypy(\d+)\.`)

func (p *PyPy) FetchAll() {
	p.getDoc()
	if p.doc != nil {
		p.doc.Find("td").Find("a").Each(func(i int, s *goquery.Selection) {
			vNameList := pypyRegExp.FindStringSubmatch(s.AttrOr("href", ""))
			if len(vNameList) < 2 {
				return
			}
			vName := p.formatVersionName(vNameList[1])
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

func (p *PyPy) Upload() {
	if len(p.versions) > 0 {
		fPath := filepath.Join(p.cnf.DirPath(), PyPyVersionFileName)
		if content, err := json.MarshalIndent(p.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			p.uploader.Upload(fPath)
		}
	}
}
