package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
	"github.com/gvcgo/proxy-collector/pkgs/utils"
)

const (
	GithubVersionFileNamePattern string = "%s.version.json"
)

/*
parse release list from github.

https://github.com/neovim/neovim
https://github.com/sharkdp/fd
https://github.com/BurntSushi/ripgrep
https://github.com/JetBrains/kotlin
https://github.com/gerardog/gsudo
https://github.com/zigtools/zls
https://github.com/typst/typst
https://github.com/nvarner/typst-lsp
https://github.com/vlang/v
https://github.com/v-analyzer/v-analyzer
https://github.com/oven-sh/bun
*/
type GithubRepo struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions map[string]Versions
}

func NewGithubRepo(cnf *confs.CollectorConf) (g *GithubRepo) {
	g = &GithubRepo{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: make(map[string]Versions),
	}
	return
}

type Assets struct {
	Name string `json:"name"`
	Url  string `json:"browser_download_url"`
}

type ReleaseItem struct {
	Assets  []*Assets `json:"assets"`
	TagName string    `json:"tag_name"`
}

func filterByUrl(dUrl string) bool {
	excludeList := []string{
		".sha256sum",
		".appimage",
		".zsync",
		"archive/refs",
	}
	for _, s := range excludeList {
		if strings.Contains(dUrl, s) {
			return false
		}
	}
	return true
}

func (g *GithubRepo) fetchRepo(repo string) {
	nList := strings.Split(repo, "/")
	name := nList[len(nList)-1]
	versions := Versions{}

	content := g.uploader.GetGithubReleaseList(repo)

	if len(content) > 0 {
		itemList := []*ReleaseItem{}
		if err := json.Unmarshal(content, &itemList); err == nil {
			for _, item := range itemList {
				versions[item.TagName] = []*VFile{}
			ASSET:
				for _, asset := range item.Assets {
					ver := &VFile{}
					ver.Url = asset.Url
					if !filterByUrl(ver.Url) {
						continue ASSET
					}
					ver.Arch = utils.ParseArch(asset.Name)
					ver.Os = utils.ParsePlatform(asset.Name)
					versions[item.TagName] = append(versions[item.TagName], ver)
					// fmt.Println(ver.Arch, ver.Os, ver.Url)
				}
			}
		} else {
			fmt.Println(err)
		}
	}
	// os.WriteFile("test.txt", content, os.ModePerm)
	g.versions[name] = versions
}

func (g *GithubRepo) FetchAll() {
	repoList := g.cnf.ReadGithubRepos()
	for _, repo := range repoList {
		fmt.Printf("fetching %s ...\n", repo)
		g.fetchRepo(repo)
		break
	}
}

func (g *GithubRepo) Upload() {
	for name, ver := range g.versions {
		if len(ver) > 0 {
			fileName := fmt.Sprintf(GithubVersionFileNamePattern, name)
			fPath := filepath.Join(g.cnf.DirPath(), fileName)
			if content, err := json.MarshalIndent(g.versions, "", "  "); err == nil && content != nil {
				os.WriteFile(fPath, content, os.ModePerm)
				g.uploader.Upload(fPath)
			}
		}
	}
}
