package versions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
)

const (
	GithubVersionFileNamePattern string = "%s.version.json"
)

type Assets struct {
	Name string `json:"name"`
	Url  string `json:"browser_download_url"`
}

type ReleaseItem struct {
	Assets     []*Assets `json:"assets"`
	TagName    string    `json:"tag_name"`
	PreRelease any       `json:"prerelease"`
}

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
https://github.com/gvcgo/asciinema
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

func filterGithubByUrl(dUrl string) bool {
	if strings.Contains(dUrl, "git-for-windows/git") && !strings.Contains(dUrl, "PortableGit") {
		return false
	}
	excludeList := []string{
		".sha256sum",
		".sha256",
		".appimage",
		".zsync",
		"archive/refs",
		".msi",
		".deb",
		".json",
		".png",
		".md",
		".vsix",
		".toml",
		".txt",
		".d.ts",
		"src.tar.gz",
		"-baseline.zip", // for bun
		"-profile.zip",  // for bun
		"denort-",       // for deno
		// "-unknown-linux-musl.tar.gz",          // for fd.
		"fd$pc-windows-msvc.zip",              // for fd.
		"linux-gnueabihf",                     // for fd
		"linux-musleabihf",                    // for fd
		"kotlin-compiler-",                    // for kotlin
		"unknown-linux-gnueabihf.",            // for ripgrep
		"unknown-linux-musleabi.",             // for ripgrep
		"unknown-linux-musleabihf.",           // for ripgrep
		"ripgrep$pc-windows-msvc.zip",         // for ripgrep
		"arm-unknown-linux-gnueabihf",         // for typst-lsp
		"typst-lsp-x86_64-unknown-linux-musl", // for typst-lsp
		"-unknown-linux-musleabi.",            // for typst
		"wasm32-wasi.",                        // for zls
		"debugging",                           // for PHP
		"-PM4.",                               // for PHP
		"tree-sitter.js",                      // for tree-sitter
		"tree-sitter.wasm",                    // for tree-sitter
		".rpm",                                // for glow, vhs
		".gz.sbom",                            // for glow, vhs
		".zip.sbom",                           // for glow, vhs
		".apk",                                // for glow
		".sha512",                             // for gleam
		"cs-x86_64-apple-darwin.gz",           // for coursier
		"cs-x86_64-pc-linux-container.gz",     // for coursier
		"cs-x86_64-pc-linux-mostly-static.gz", // for soursier
		"cs-x86_64-pc-linux-static.gz",        // for coursier
		"cs-x86_64-pc-linux.gz",               // for coursier
		"cs-x86_64-pc-win32.zip",              // for coursier
		"coursier.jar",                        // for coursier
		"typst-preview$index.html",
		"typst-preview$.debug",
		"typst-preview$.dwarf",
		"typst-preview$.pdb",
		"typst-preview$alpine",
		"typst-preview$armhf",
		"agg$-gnueabihf",
		"agg$-musl",
		"upx$src",
		"upx$armeb",
		"upx$a.zip",
		"upx$d.zip",
		"upx$w.zip",
		"upx$wce.zip",
		"upx$mips",
		"upx$powerpc",
		"upx$dos",
		"upx$arm_linux",
	}

	for _, s := range excludeList {
		if strings.Contains(s, "$") {
			ss := strings.Split(s, "$")
			if strings.Contains(dUrl, ss[0]) && strings.Contains(dUrl, ss[1]) {
				return false
			}
		} else if strings.Contains(dUrl, s) {
			return false
		}
	}
	return true
}

var ToFindVersionList = []string{
	"bun",
}

func (g *GithubRepo) fetchRepo(repo string) {
	nList := strings.Split(repo, "/")
	name := nList[len(nList)-1]
	if strings.HasPrefix(name, "PHP-") {
		name = "php"
	}
	if name == "Odin" {
		name = "odin"
	}

	versions := Versions{}

	content := g.uploader.GetGithubReleaseList(repo)

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
						// if strings.Contains(asset.Url, "typst-") && strings.Contains(asset.Url, "windows") {
						// 	fmt.Println(asset.Url)
						// }
						ver.Arch = utils.ParseArch(asset.Url)
						ver.Os = utils.ParsePlatform(asset.Url)
						for _, n := range ToFindVersionList {
							if name == n {
								item.TagName = FindVersion(item.TagName)
								break
							}
						}
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
	g.versions[name] = versions
}

func (g *GithubRepo) FetchAll() {
	repoList := g.cnf.ReadGithubRepos()
	for _, repo := range repoList {
		// if repo != "oven-sh/bun" {
		// 	continue
		// }
		rp := repo
		fmt.Printf("fetching %s ...\n", rp)
		g.fetchRepo(rp)
	}
}

func (g *GithubRepo) Upload() {
	for name, ver := range g.versions {
		if len(ver) > 0 {
			fileName := fmt.Sprintf(GithubVersionFileNamePattern, name)
			fPath := filepath.Join(g.cnf.DirPath(), fileName)
			if content, err := json.MarshalIndent(ver, "", "  "); err == nil && content != nil {
				os.WriteFile(fPath, content, os.ModePerm)
				g.uploader.Upload(fPath)
			}
		}
	}
}
