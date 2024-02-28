package versions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
	"github.com/gvcgo/proxy-collector/pkgs/utils"
)

const (
	NodeVersionFileName string = "nodejs.version.json"
	NodeDownloadUrl     string = "https://nodejs.org/download/release"
	NodeSumUrlPattern   string = "https://nodejs.org/download/release/%s/SHASUMS256.txt"
)

type Item struct {
	Version string `json:"version"`
	LTS     any    `json:"lts"`
	Date    string `json:"date"`
}

/*
https://nodejs.org/dist/index.json
https://nodejs.org/download/release
*/
type Nodejs struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	itemList []*Item
}

func NewNodejs(cnf *confs.CollectorConf) (n *Nodejs) {
	n = &Nodejs{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://nodejs.org/dist/index.json",
		itemList: []*Item{},
	}
	return
}

// get versions released after 2018.
// And all LTS versions.
func filterVersion(vItem *Item) bool {
	if gconv.Bool(vItem.LTS) {
		return true
	}
	r := false
	dList := strings.Split(vItem.Date, "-")
	if len(dList) > 1 {
		year, _ := strconv.Atoi(dList[0])
		if year > 2017 {
			r = true
		}
	}
	return r
}

func (n *Nodejs) getVersion(vItem *Item) {
	n.fetcher.SetUrl(fmt.Sprintf(NodeSumUrlPattern, vItem.Version))
	n.fetcher.Timeout = 30 * time.Second
	content, _ := n.fetcher.GetString()
	// os.WriteFile("test.txt", []byte(content), os.ModePerm)

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, ".tar.gz") || strings.Contains(line, ".zip") {
			uList := strings.Split(strings.TrimSpace(line), " ")
			if len(uList) >= 2 {
				fName := strings.TrimSpace(uList[len(uList)-1])
				archStr := utils.ParseArch(fName)
				osStr := utils.ParsePlatform(fName)
				if archStr != "" && osStr != "" {
					ver := &VFile{}
					ver.Sum = strings.TrimSpace(uList[0])
					ver.SumType = "sha256"
					ver.Url = fmt.Sprintf(
						"%s/%s/%s",
						NodeDownloadUrl,
						vItem.Version,
						fName,
					)
					ver.Arch = archStr
					ver.Os = osStr
					vName := strings.TrimPrefix(vItem.Version, "v")
					if vlist, ok := n.versions[vName]; !ok || vlist == nil {
						n.versions[vName] = []*VFile{}
					}
					if gconv.Bool(vItem.LTS) {
						ver.Extra = "LTS"
					}
					n.versions[vName] = append(n.versions[vName], ver)
				}
			}
		}
	}
}

func (n *Nodejs) GetVersions() {
	n.fetcher.SetUrl(n.homepage)
	n.fetcher.Timeout = 180 * time.Second
	if resp := n.fetcher.Get(); resp != nil {
		content, _ := io.ReadAll(resp.RawBody())
		if err := json.Unmarshal(content, &n.itemList); err != nil {
			gprint.PrintError(fmt.Sprintf("Parse content from %s failed.", n.fetcher.Url))
			return
		}
	}
	for _, item := range n.itemList {
		if !filterVersion(item) {
			continue
		}
		n.getVersion(item)
	}
}

func (n *Nodejs) FetchAll() {
	n.GetVersions()
}

func (n *Nodejs) Upload() {
	if len(n.versions) > 0 {
		fPath := filepath.Join(n.cnf.DirPath(), NodeVersionFileName)
		if content, err := json.MarshalIndent(n.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			n.uploader.Upload(fPath)
		}
	}
}
