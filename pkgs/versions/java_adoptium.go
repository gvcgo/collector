package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	versions map[string]Versions
	jdk      *JDK
}

func NewAdoptiumJDK(cnf *confs.CollectorConf) (a *AdoptiumJDK) {
	a = &AdoptiumJDK{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: make(map[string]Versions),
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

func (a *AdoptiumJDK) fetchRepo(repo string) {
	// TODO: for adoptium releases.
	nList := strings.Split(repo, "/")
	name := nList[len(nList)-1]
	versions := Versions{}

	content := a.uploader.GetGithubReleaseList(repo)

	if len(content) > 0 {
		itemList := []*ReleaseItem{}
		if err := json.Unmarshal(content, &itemList); err == nil {
			for _, item := range itemList {
				for _, asset := range item.Assets {
					// if strings.Contains(item.TagName, "1.0.30") {
					// 	fmt.Println(asset.Url, filterByUrl(asset.Url))
					// }
					ver := &VFile{}
					ver.Url = asset.Url
					if filterGithubByUrl(asset.Url) {
						ver.Arch = utils.ParseArch(asset.Url)
						ver.Os = utils.ParsePlatform(asset.Url)
						versions[item.TagName] = append(versions[item.TagName], ver)
					}
					// fmt.Println(ver.Arch, ver.Os, ver.Url)
				}
			}
		} else {
			fmt.Println(err)
		}
	}
	// os.WriteFile("test.txt", content, os.ModePerm)
	a.versions[name] = versions
}

func (a *AdoptiumJDK) FetchAll() {
	repoList := a.GetRepoList()
	for _, repo := range repoList {
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
