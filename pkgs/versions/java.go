package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
	"github.com/gvcgo/proxy-collector/pkgs/utils"
)

/*
https://injdk.cn/

Latest versions only:
https://www.oracle.com/cn/java/technologies/downloads/
https://www.oracle.com/java/technologies/downloads/
*/

const (
	JavaVersionFileName string = "jdk.version.json"
	OfficialHomepage    string = "https://www.oracle.com/java/technologies/downloads/"
)

type JDK struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewJDK(cnf *confs.CollectorConf) (j *JDK) {
	j = &JDK{
		cnf:      cnf,
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://injdk.cn/",
		uploader: upload.NewUploader(cnf),
	}
	if confs.EnableProxyOrNot() {
		pxy := j.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		j.fetcher.Proxy = pxy
	}
	return
}

func (j *JDK) GetFileSuffix(fName string) string {
	var allowedSuffixes = []string{
		".zip",
		".tar.gz",
		".tar.bz2",
		".tar.xz",
	}
	for _, k := range allowedSuffixes {
		if strings.HasSuffix(fName, k) {
			return k
		}
	}
	return ""
}

func (j *JDK) GetDoc(homepage string) {
	j.fetcher.Url = homepage
	if resp := j.fetcher.Get(); resp != nil {
		j.doc, _ = goquery.NewDocumentFromReader(resp.RawBody())
	}
	if j.doc == nil {
		gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", j.fetcher.Url))
		os.Exit(1)
	}
}

func formatVersionName(vName string) string {
	vName = strings.ToLower(vName)
	vName = strings.ReplaceAll(vName, "\n", "")
	vName = strings.ReplaceAll(vName, "\r", "")
	vName = strings.ReplaceAll(vName, " ", "")
	vName = strings.ReplaceAll(vName, "jdk", "")
	return vName
}

func (j *JDK) GetVersionsFromInjdk() {
	j.homepage = "https://injdk.cn/"
	j.GetDoc(j.homepage)

	j.doc.Find("div#oracle-jdk").Find("div.col-sm-3").Each(func(i int, s *goquery.Selection) {
		vName := formatVersionName(strings.ToLower(s.Find("span").Text()))
		var extraStr string
		if strings.Contains(vName, "(lts)") {
			vName = strings.ReplaceAll(vName, "(lts)", "")
			extraStr = "LTS"
		}
		extraStr = strings.Join([]string{extraStr, "injdk.cn"}, ",")

		s.Find("li").Each(func(i int, ss *goquery.Selection) {
			if strings.Contains(vName, "jdk8") {
				return
			}

			fileName := strings.ReplaceAll(strings.ToLower(ss.Find("a").Text()), " ", "")
			ver := &VFile{}
			ver.Arch = utils.ParseArch(fileName)
			ver.Os = utils.ParsePlatform(fileName)
			if ver.Arch == "" || ver.Os == "" {
				return
			}
			if suffix := j.GetFileSuffix(fileName); suffix == "" {
				return
			}
			ver.Url = strings.ReplaceAll(ss.Find("a").AttrOr("href", ""), " ", "")
			if ver.Url == "" {
				return
			}

			if vlist, ok := j.versions[vName]; !ok || vlist == nil {
				j.versions[vName] = []*VFile{}
			}

			ver.Extra = extraStr
			j.versions[vName] = append(j.versions[vName], ver)
		})
	})

	j.doc.Find("#Kona").Find("div.col-sm-3").Each(func(i int, s *goquery.Selection) {
		vName := formatVersionName(strings.ToLower(s.Find("span").Text()))
		if strings.Contains(vName, "(lts)") {
			vName = strings.ReplaceAll(vName, "(lts)", "")
		}
		extraStr := "LTS$injdk.cn"

		s.Find("li").Each(func(i int, ss *goquery.Selection) {
			if !strings.Contains(vName, "jdk8") {
				return
			}

			fileName := strings.ReplaceAll(strings.ToLower(ss.Find("a").Text()), " ", "")
			ver := &VFile{}
			ver.Arch = utils.ParseArch(fileName)
			ver.Os = utils.ParsePlatform(fileName)
			if ver.Arch == "" || ver.Os == "" {
				return
			}
			if suffix := j.GetFileSuffix(fileName); suffix == "" {
				return
			}
			ver.Url = strings.ReplaceAll(ss.Find("a").AttrOr("href", ""), " ", "")
			if ver.Url == "" {
				return
			}

			if vlist, ok := j.versions[vName]; !ok || vlist == nil {
				j.versions[vName] = []*VFile{}
			}
			ver.Extra = extraStr
			j.versions[vName] = append(j.versions[vName], ver)
		})
	})
}

func (j *JDK) GetSha(sUrl string) (res string) {
	j.fetcher.Url = sUrl
	res, _ = j.fetcher.GetString()
	return
}

func (j *JDK) GetVersionsFromOfficial() {
	j.homepage = OfficialHomepage
	j.GetDoc(j.homepage)

	j.doc.Find("ul.rw-inpagetabs").First().Find("li").Each(func(i int, s *goquery.Selection) {
		v, _ := s.Find("a").Attr("href")
		sList := strings.Split(v, "java")
		vn := sList[len(sList)-1]
		j.doc.Find(fmt.Sprintf("div#java%s", vn)).After("nav").Find("table").Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}
			tArchive := strings.ToLower(s.Find("td").Eq(0).Text())
			tArchive = strings.ReplaceAll(tArchive, " ", "")
			tUrl, _ := s.Find("td").Eq(2).Find("a").Eq(0).Attr("href")
			// tSha, _ := s.Find("td").Eq(2).Find("a").Eq(1).Attr("href")
			if !strings.Contains(tArchive, "archive") {
				return
			}
			ver := &VFile{}
			ver.Arch = utils.ParseArch(tUrl)
			ver.Os = utils.ParsePlatform(tUrl)
			if ver.Arch == "" || ver.Os == "" || tUrl == "" {
				return
			}

			ver.Url = tUrl
			if suffix := j.GetFileSuffix(ver.Url); suffix == "" {
				return
			}
			ver.Sum = j.GetSha(fmt.Sprintf("%s.%s", ver.Url, "sha256"))
			if ver.Sum != "" {
				ver.SumType = "sha256"
			}
			if vlist, ok := j.versions[vn]; !ok || vlist == nil {
				j.versions[vn] = []*VFile{}
			}
			ver.Extra = "oracle.com"
			j.versions[vn] = append(j.versions[vn], ver)
		})
	})
}

func (j *JDK) FetchAll() {
	j.GetVersionsFromInjdk()
	j.GetVersionsFromOfficial()
}

func (j *JDK) Upload() {
	if len(j.versions) > 0 {
		fPath := filepath.Join(j.cnf.DirPath(), JavaVersionFileName)
		if content, err := json.MarshalIndent(j.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			j.uploader.Upload(fPath)
		}
	}
}
