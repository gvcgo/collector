package versions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
)

/*
https://injdk.cn/

Latest versions only:
https://www.oracle.com/cn/java/technologies/downloads/
https://www.oracle.com/java/technologies/downloads/

temurin:
https://api.adoptium.net/v3/info/available_releases
https://api.adoptium.net/v3/assets/latest/8/hotspot
*/

const (
	JavaVersionFileName string = "jdk.version.json"
	AdoptiumURL         string = "https://api.adoptium.net/v3/info/available_releases"
	AdoptiumAssetsURL   string = "https://api.adoptium.net/v3/assets/latest/%d/hotspot"
)

/*
	{
	    "available_lts_releases": [
	        8,
	        11,
	        17,
	        21
	    ],
	    "available_releases": [
	        8,
	        11,
	        16,
	        17,
	        18,
	        19,
	        20,
	        21
	    ],
	    "most_recent_feature_release": 21,
	    "most_recent_feature_version": 23,
	    "most_recent_lts": 21,
	    "tip_version": 23
	}
*/
type JdkAvailableVersions struct {
	Releases []int `json:"available_releases"`
}

/*
	{
		"binary": {
			"architecture": "x64",
			"download_count": 590242,
			"heap_size": "normal",
			"image_type": "jdk",
			"jvm_impl": "hotspot",
			"os": "linux",
			"package": {
				"checksum": "fcfd08abe39f18e719e391f2fc37b8ac1053075426d10efac4cbf8969e7aa55e",
				"checksum_link": "https://github.com/adoptium/temurin8-binaries/releases/download/jdk8u402-b06/OpenJDK8U-jdk_x64_linux_hotspot_8u402b06.tar.gz.sha256.txt",
				"download_count": 590242,
				"link": "https://github.com/adoptium/temurin8-binaries/releases/download/jdk8u402-b06/OpenJDK8U-jdk_x64_linux_hotspot_8u402b06.tar.gz",
				"metadata_link": "https://github.com/adoptium/temurin8-binaries/releases/download/jdk8u402-b06/OpenJDK8U-jdk_x64_linux_hotspot_8u402b06.tar.gz.json",
				"name": "OpenJDK8U-jdk_x64_linux_hotspot_8u402b06.tar.gz",
				"signature_link": "https://github.com/adoptium/temurin8-binaries/releases/download/jdk8u402-b06/OpenJDK8U-jdk_x64_linux_hotspot_8u402b06.tar.gz.sig",
				"size": 103003119
			},
			"project": "jdk",
			"scm_ref": "jdk8u402-b06_adopt",
			"updated_at": "2024-01-19T15:54:40Z"
		},
		"release_link": "https://github.com/adoptium/temurin8-binaries/releases/tag/jdk8u402-b06",
		"release_name": "jdk8u402-b06",
		"vendor": "eclipse",
		"version": {
			"build": 6,
			"major": 8,
			"minor": 0,
			"openjdk_version": "1.8.0_402-b06",
			"security": 402,
			"semver": "8.0.402+6"
		}
	}
*/
type JPackage struct {
	Checksum string `json:"checksum"`
	Url      string `json:"link"`
}

type JBinary struct {
	Arch      string    `json:"architecture"`
	Os        string    `json:"os"`
	ImageType string    `json:"image_type"`
	Package   *JPackage `json:"package"`
	CLib      string    `json:"c_lib"`
}

type JdkItem struct {
	Binary      *JBinary `json:"binary"`
	ReleaseName string   `json:"release_name"`
	Vendor      string   `json:"vendor"`
}

type JDK struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
}

func NewJDK(cnf *confs.CollectorConf) (j *JDK) {
	j = &JDK{
		cnf:      cnf,
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: AdoptiumURL,
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

func (j *JDK) FetchAll() {
	j.fetcher.Timeout = time.Second * 60
	j.fetcher.SetUrl(j.homepage)
	versionList := JdkAvailableVersions{
		Releases: []int{},
	}
	if resp := j.fetcher.Get(); resp != nil {
		content, _ := io.ReadAll(resp.RawBody())
		if err := json.Unmarshal(content, &versionList); err != nil {
			gprint.PrintError(fmt.Sprintf("Parse content from %s failed.", j.fetcher.Url))
			return
		}
	}

OUTTER:
	for _, vInt := range versionList.Releases {
		vName := fmt.Sprintf("%d", vInt)
		u := fmt.Sprintf(AdoptiumAssetsURL, vInt)
		j.fetcher.SetUrl(u)

		vList := []*JdkItem{}
		if resp := j.fetcher.Get(); resp != nil {
			content, _ := io.ReadAll(resp.RawBody())
			if err := json.Unmarshal(content, &vList); err != nil {
				gprint.PrintError(fmt.Sprintf("Parse content from %s failed.", j.fetcher.Url))
				continue OUTTER
			}
		}
	INNER:
		for _, item := range vList {
			if item.Binary == nil || item.Binary.ImageType != "jdk" || item.Binary.Package == nil {
				continue INNER
			}
			if item.Binary.Os == "alpine-linux" || item.Binary.CLib == "musl" {
				continue INNER
			}
			ver := &VFile{}
			ver.Url = item.Binary.Package.Url
			ver.Sum = item.Binary.Package.Checksum
			if ver.Sum != "" {
				ver.SumType = "sha256"
			}
			ver.Arch = utils.ParseArch(item.Binary.Arch)
			ver.Os = utils.ParsePlatform(item.Binary.Os)
			if ver.Arch == "" || ver.Os == "" {
				continue INNER
			}
			ver.Extra = fmt.Sprintf("%s$%s", item.ReleaseName, item.Vendor)
			if vl, ok := j.versions[vName]; !ok || vl == nil {
				j.versions[vName] = []*VFile{}
			}
			j.versions[vName] = append(j.versions[vName], ver)
		}
	}
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
