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
	KubectlVersionFileName string = "kubectl.version.json"
	KubectlURL             string = "https://kubernetes.io/releases/patch-releases/"
	KubectlLatestURL       string = "https://dl.k8s.io/release/stable.txt"
)

/*
https://kubernetes.io/zh-cn/docs/tasks/tools/

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"
curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl.sha256"

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl.sha256"

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl.sha256"

curl.exe -LO "https://dl.k8s.io/release/v1.29.2/bin/windows/amd64/kubectl.exe"
curl.exe -LO "https://dl.k8s.io/v1.29.2/bin/windows/amd64/kubectl.exe.sha256"
*/

const (
	KubectlDownloadUrlPattern  string = `https://dl.k8s.io/release/v%s/bin/%s/%s/kubectl`
	KubectlSha256UrlPattern    string = `https://dl.k8s.io/release/v%s/bin/%s/%s/kubectl.sha256`
	KubectlExeSha256UrlPattern string = `https://dl.k8s.io/v%s/bin/%s/%s/kubectl.exe.sha256`
)

var versionRegexp = regexp.MustCompile(`\d+(.\d+){2}`)

type Kubectl struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	doc      *goquery.Document
}

func NewKubectl(cnf *confs.CollectorConf) (k *Kubectl) {
	k = &Kubectl{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: make(Versions),
		fetcher:  request.NewFetcher(),
	}
	k.fetcher.Timeout = 5 * time.Second
	return
}

func (k *Kubectl) GetVersions() (r []string) {
	k.fetcher.SetUrl(KubectlURL)
	if resp := k.fetcher.Get(); resp != nil {
		var err error
		k.doc, err = goquery.NewDocumentFromReader(resp.RawBody())
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if k.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", k.fetcher.Url))
			os.Exit(1)
		}

		k.doc.Find("tr").Find("td").Each(func(_ int, s *goquery.Selection) {
			ss := versionRegexp.FindString(strings.TrimSpace(s.Text()))
			if ss != "" && strings.Contains(ss, ".") {
				r = append(r, ss)
			}
		})
	}

	k.fetcher.SetUrl(KubectlLatestURL)
	s, _ := k.fetcher.GetString()
	latestVersion := versionRegexp.FindString(s)
	if latestVersion != "" {
		r = append(r, latestVersion)
	}
	return
}

func (k *Kubectl) fetchOne(vStr, archStr, osStr string) {
	sha256Url := fmt.Sprintf(KubectlSha256UrlPattern, vStr, osStr, archStr)
	if osStr == "windows" {
		sha256Url = fmt.Sprintf(KubectlExeSha256UrlPattern, vStr, osStr, archStr)
	}
	k.fetcher.SetUrl(sha256Url)
	sha256, _ := k.fetcher.GetString()
	if strings.Contains(sha256, "NoSuchKey") {
		return
	}
	sha256 = strings.TrimSpace(sha256)
	fmt.Println(vStr, archStr, osStr, sha256)

	u := fmt.Sprintf(KubectlDownloadUrlPattern, vStr, osStr, archStr)
	if osStr == "windows" {
		u += ".exe"
	}
	ver := &VFile{
		Url:     u,
		Arch:    archStr,
		Os:      osStr,
		Sum:     sha256,
		SumType: "sha256",
	}
	if vfiles, ok := k.versions[vStr]; !ok || vfiles == nil {
		k.versions[vStr] = []*VFile{}
	}
	k.versions[vStr] = append(k.versions[vStr], ver)
}

func (k *Kubectl) FetchAll() {
	archOsList := []string{
		"darwin/amd64",
		"darwin/arm64",
		"linux/amd64",
		"linux/arm64",
		"windows/amd64",
	}
	for _, vStr := range k.GetVersions() {
		for _, archOs := range archOsList {
			sList := strings.Split(archOs, "/")
			k.fetchOne(vStr, sList[1], sList[0])
		}
	}
}

func (k *Kubectl) Upload() {
	if len(k.versions) > 0 {
		fPath := filepath.Join(k.cnf.DirPath(), KubectlVersionFileName)
		if content, err := json.MarshalIndent(k.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			k.uploader.Upload(fPath)
		}
	}
}
