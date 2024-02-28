package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

const (
	MavenVersionFilename string = "maven.version.json"
	MavenBinUrlPattern   string = "%s%s/binaries/apache-maven-%s-bin.tar.gz"
	MavenSumUrlPattern   string = "%s%s/binaries/apache-maven-%s-bin.tar.gz.sha512"
	MavenSumType         string = "sha512"
)

/*
maven-1, maven-2;
https://dlcdn.apache.org/maven/
https://dlcdn.apache.org/maven/maven-3/
https://dlcdn.apache.org/maven/maven-4/
*/

type Maven struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewMaven(cnf *confs.CollectorConf) (m *Maven) {
	m = &Maven{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
	}
	if confs.EnableProxyOrNot() {
		pxy := m.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		m.fetcher.Proxy = pxy
	}
	return
}

func (m *Maven) getDoc() {
	m.fetcher.Url = m.homepage
	if resp := m.fetcher.Get(); resp != nil {
		m.doc, _ = goquery.NewDocumentFromReader(resp.RawBody())
	}
}

func (m *Maven) getSum(sumUrl string) string {
	m.fetcher.SetUrl(sumUrl)
	r, _ := m.fetcher.GetString()
	return r
}

func (m *Maven) GetVersions() {
	uList := map[string]string{
		"4.": "https://dlcdn.apache.org/maven/maven-4/",
		"3.": "https://dlcdn.apache.org/maven/maven-3/",
	}

	for k, u := range uList {
		m.homepage = u
		m.getDoc()
		if m.doc != nil {
			m.doc.Find("a").Each(func(i int, s *goquery.Selection) {
				link := s.AttrOr("href", "")
				if strings.HasPrefix(link, k) {
					ver := &VFile{}
					vName := strings.ReplaceAll(link, "/", "")
					ver.Url = fmt.Sprintf(
						MavenBinUrlPattern,
						u,
						vName,
						vName,
					)
					ver.Sum = m.getSum(fmt.Sprintf(
						MavenSumUrlPattern,
						u,
						vName,
						vName,
					))
					if ver.Sum != "" {
						ver.SumType = MavenSumType
					}
					if vlist, ok := m.versions[vName]; !ok || vlist == nil {
						m.versions[vName] = []*VFile{}
					}
					ver.Arch = "any"
					ver.Os = "any"
					m.versions[vName] = append(m.versions[vName], ver)
				}
			})
		}
	}
}

func (m *Maven) FetchAll() {
	if m.doc == nil {
		m.getDoc()
	}
	m.GetVersions()
}

func (m *Maven) Upload() {
	if len(m.versions) > 0 {
		fPath := filepath.Join(m.cnf.DirPath(), MavenVersionFilename)
		if content, err := json.MarshalIndent(m.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			m.uploader.Upload(fPath)
		}
	}
}
