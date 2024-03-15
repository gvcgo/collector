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
	ScalaVersionFileName string = "scala.version.json"
)

/*
Scala versions.

https://www.scala-lang.org/download/all.html
*/
type Scala struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewScala(cnf *confs.CollectorConf) (s *Scala) {
	s = &Scala{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://www.scala-lang.org/download/all.html",
	}
	if confs.EnableProxyOrNot() {
		pxy := s.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		s.fetcher.Proxy = pxy
	}
	return
}

func (s *Scala) getDoc() {
	s.fetcher.SetUrl(s.homepage)
	s.fetcher.Timeout = 180 * time.Second
	if resp, sCode := s.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
		var err error
		s.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if s.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", s.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Printf("Failed: %s, code: %d", s.homepage, sCode)
	}
}

func (s *Scala) FetchAll() {
	s.getDoc()
	if s.doc != nil {
		s.doc.Find("div.download-elem").Find("a").Each(func(_ int, ss *goquery.Selection) {
			vName := strings.ReplaceAll(ss.Text(), "Scala ", "")
			vName = strings.ReplaceAll(vName, " ", "-")
			if _, ok := s.versions[vName]; !ok {
				s.versions[vName] = []*VFile{
					{
						Extra: "please use coursier to install.",
					},
				}
			}
		})
	}
}

func (s *Scala) Upload() {
	if len(s.versions) > 0 {
		fPath := filepath.Join(s.cnf.DirPath(), ScalaVersionFileName)
		if content, err := json.MarshalIndent(s.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			s.uploader.Upload(fPath)
		}
	}
}
