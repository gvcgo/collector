package versions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

const (
	PhpVersionFileName string = "php.version.json"
)

var (
	VersionPattern = regexp.MustCompile(`(\d+\.\d+\.\d+)`)
)

/*
Windows: https://windows.php.net/download/
https://windows.php.net/downloads/releases/archives/

Unix: https://www.php.net/manual/zh/install.unix.php
https://www.php.net/downloads
https://www.php.net/releases/
*/
type PhP struct {
	cnf       *confs.CollectorConf
	uploader  *upload.Uploader
	versions  Versions
	fetcher   *request.Fetcher
	homepage  string
	doc       *goquery.Document
	urlFilter map[string]struct{}
}

func NewPhP(cnf *confs.CollectorConf) (p *PhP) {
	p = &PhP{
		cnf:       cnf,
		uploader:  upload.NewUploader(cnf),
		versions:  Versions{},
		fetcher:   request.NewFetcher(),
		urlFilter: map[string]struct{}{},
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

func (p *PhP) getDoc() {
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

func filterPhPVersion(vName string) bool {
	r := false
	vList := strings.Split(vName, ".")
	if len(vList) < 2 {
		return r
	}
	if gconv.Int(vList[0]) >= 7 && gconv.Int(vList[1]) >= 2 {
		r = true
	}
	return r
}

func (p *PhP) GetWindowsVersions() {
	vcPattern := regexp.MustCompile(`(v[a-z]\d+)`)
	baseUrl := "https://windows.php.net"
	p.homepage = "https://windows.php.net/downloads/releases/archives/"
	p.doc = nil
	p.getDoc()
	if p.doc != nil {
		p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			fName := strings.ToLower(s.Text())
			if !strings.HasPrefix(fName, "php") || !strings.Contains(fName, ".zip") {
				return
			}
			if strings.Contains(fName, "-debug-") || strings.Contains(fName, "-devel-") {
				return
			}
			if strings.Contains(fName, "test-pack") || strings.Contains(fName, "-src") {
				return
			}
			vName := VersionPattern.FindString(fName)
			if !filterPhPVersion(vName) {
				return
			}
			ver := &VFile{}
			u := s.AttrOr("href", "")
			if !strings.HasPrefix(u, "https://") {
				u, _ = url.JoinPath(baseUrl, u)
			}
			ver.Url = u
			ver.Extra = vcPattern.FindString(fName)
			if strings.Contains(fName, "-nts-") {
				ver.Extra += "$NTS"
			}

			ver.Arch = "386"
			if strings.Contains(fName, "-x64") {
				ver.Arch = "amd64"
			}
			ver.Os = "windows"
			if vlist, ok := p.versions[vName]; !ok || vlist == nil {
				p.versions[vName] = []*VFile{}
			}
			if _, ok := p.urlFilter[ver.Url]; !ok {
				p.versions[vName] = append(p.versions[vName], ver)
				p.urlFilter[ver.Url] = struct{}{}
			}
			// fmt.Println(ver.Arch, ver.Os, ver.Extra, ver.Url)
		})
	}
}

func (p *PhP) parseULTag(s *goquery.Selection) {
	baseUrl := "https://www.php.net"
	s.Find("a").Each(func(_ int, s *goquery.Selection) {
		u := s.AttrOr("href", "")
		if u == "" {
			return
		}
		fName := strings.ToLower(s.Text())
		vName := VersionPattern.FindString(fName)

		if vName == "" || !filterPhPVersion(vName) {
			return
		}
		// fmt.Println(fName, vName, u)
		if strings.Contains(fName, "tar.gz") {
			ver := &VFile{}
			if !strings.HasPrefix(u, "http") {
				u, _ = url.JoinPath(baseUrl, u)
			}
			ver.Url = u
			ver.Os = "linux"
			ver.Arch = "all"
			ver.Extra = "src"
			if vlist, ok := p.versions[vName]; !ok || vlist == nil {
				p.versions[vName] = []*VFile{}
			}
			if _, ok := p.urlFilter[ver.Url]; !ok {
				p.versions[vName] = append(p.versions[vName], ver)
				p.urlFilter[ver.Url] = struct{}{}
			}
			// fmt.Println(ver.Arch, ver.Os, ver.Extra, ver.Url)
		}
	})
}

func (p *PhP) GetUnixVersions() {

	p.doc = nil
	p.homepage = "https://www.php.net/downloads"
	p.getDoc()
	if p.doc != nil {
		p.doc.Find("section#layout-content").Find("ul").Each(func(_ int, s *goquery.Selection) {
			p.parseULTag(s)
		})
	}

	p.homepage = "https://www.php.net/releases/"
	p.doc = nil
	p.getDoc()
	if p.doc != nil {
		p.doc.Find("ul").Each(func(_ int, s *goquery.Selection) {
			p.parseULTag(s)
		})
	}
}

func (p *PhP) FetchAll() {
	p.GetWindowsVersions()
	p.GetUnixVersions()
}

func (p *PhP) Upload() {
	if len(p.versions) > 0 {
		fPath := filepath.Join(p.cnf.DirPath(), PhpVersionFileName)
		if content, err := json.MarshalIndent(p.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			p.uploader.Upload(fPath)
		}
	}
}
