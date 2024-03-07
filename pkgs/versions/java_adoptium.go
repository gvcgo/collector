package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
)

const (
	AdoptiumVersionListUrl   string = "https://api.adoptium.net/v3/info/available_releases"
	AdoptiumGithubRepPattern string = `adoptium/temurin%d-binaries`
)

var AdoptiumRegExp = regexp.MustCompile(`github\.com/(.+)/release`)

/*
https://api.adoptium.net/v3/info/available_releases

https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.10%2B7/OpenJDK17U-jdk_x64_linux_hotspot_17.0.10_7.tar.gz

@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
Find realease list for each version with prereleases excluded.
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
Example:
https://github.com/adoptium/temurin17-binaries/releases
*/
type AdoptiumJDK struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	jdk      *JDK
}

func NewAdoptiumJDK(cnf *confs.CollectorConf) (a *AdoptiumJDK) {
	a = &AdoptiumJDK{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: make(Versions),
		jdk:      NewJDK(cnf),
	}
	return
}

func (a *AdoptiumJDK) GetRepoList() (r []string) {
	a.jdk.FetchAll()
	for _, v := range a.jdk.versions {
		if len(v) > 0 {
			s := AdoptiumRegExp.FindStringSubmatch(v[0].Url)
			if len(s) > 1 {
				r = append(r, s[1])
			}
		}
	}
	return
}

func filterJDKByUrl(dUrl string) bool {
	if !strings.Contains(dUrl, "_hotspot_") {
		return false
	}
	toBunList := []string{
		".sig",
		".pkg",
		".exe",
		".smi",
		"debugimage",
		"alpine",
		"static",
		"jre",
	}
	for _, b := range toBunList {
		if strings.Contains(dUrl, b) {
			return false
		}
	}
	return true
}

func (a *AdoptiumJDK) fetchRepo(repo string) {
	content := a.uploader.GetGithubReleaseList(repo)

	if len(content) > 0 {
		itemList := []*ReleaseItem{}
		if err := json.Unmarshal(content, &itemList); err == nil {
		OUTTER:
			for _, item := range itemList {
				if gconv.Bool(item.PreRelease) {
					continue OUTTER
				}
			INNER:
				for _, asset := range item.Assets {
					if !strings.Contains(asset.Url, "_hotspot_") {
						continue INNER
					}
					if !filterJDKByUrl(asset.Url) {
						continue INNER
					}
					ver := &VFile{}
					ver.Url = asset.Url
					if filterGithubByUrl(asset.Url) {
						ver.Arch = utils.ParseArch(asset.Url)
						ver.Os = utils.ParsePlatform(asset.Url)
						if ver.Os == "linux" && ver.Arch == "" {
							continue INNER
						}
						if vlist, ok := a.versions[item.TagName]; !ok || vlist == nil {
							a.versions[item.TagName] = []*VFile{}
						}
						a.versions[item.TagName] = append(a.versions[item.TagName], ver)
					}
					// fmt.Println(ver.Arch, ver.Os, ver.Url)
				}
			}
		} else {
			fmt.Println(err)
		}
	}
	// os.WriteFile("test.txt", content, os.ModePerm)
}

func (a *AdoptiumJDK) FetchAll() {
	repoList := a.GetRepoList()
	for _, repo := range repoList {
		fmt.Printf("fetching %s...\n", repo)
		a.fetchRepo(repo)
	}
}

func (a *AdoptiumJDK) Upload() {
	if len(a.versions) > 0 {
		fPath := filepath.Join(a.cnf.DirPath(), JavaVersionFileName)
		if content, err := json.MarshalIndent(a.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			a.uploader.Upload(fPath)
		}
	}
}
