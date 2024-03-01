package versions

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

const (
	InstallerVersionFileName string = "installers.version.json"
)

/*
Only the latest version for:

1. android sdkmanager
https://developer.android.com/tools/sdkmanager?hl=zh-cn

	download:
	https://developer.android.google.cn/studio?hl=zh-cn
	https://developer.android.com/studio?hl=en

2. cygwin installer
https://cygwin.com/install.html

	download:
	https://cygwin.com/setup-x86_64.exe

3. msys2 installer
https://www.msys2.org/#installation
https://github.com/msys2/msys2-installer/releases

	download:
	https://github.com/msys2/msys2-installer/releases/download/2024-01-13/msys2-x86_64-20240113.exe

4. rust installer
https://forge.rust-lang.org/infra/other-installation-methods.html

	download:
	https://static.rust-lang.org/rustup/rustup-init.sh
	https://static.rust-lang.org/rustup/dist/i686-pc-windows-gnu/rustup-init.exe

5. VSCode
https://code.visualstudio.com/sha?build=stable

6. miniconda
https://anaconda.org.cn/anaconda/install/silent-mode/
https://repo.anaconda.com/miniconda/
*/
type Installer struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
	doc      *goquery.Document
}

func NewInstaller(cnf *confs.CollectorConf) (i *Installer) {
	i = &Installer{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
	}
	if confs.EnableProxyOrNot() {
		pxy := i.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		i.fetcher.Proxy = pxy
	}
	return
}

func (i *Installer) getDoc() {
	i.fetcher.SetUrl(i.homepage)
	i.fetcher.Timeout = 30 * time.Second
	if resp, sCode := i.fetcher.GetString(); resp != "" && sCode == 200 {
		// fmt.Println(resp)
		var err error
		i.doc, err = goquery.NewDocumentFromReader(strings.NewReader(resp))
		if err != nil {
			gprint.PrintError(fmt.Sprintf("Parse page errored: %+v", err))
		}
		if i.doc == nil {
			gprint.PrintError(fmt.Sprintf("Cannot parse html for %s", i.fetcher.Url))
			os.Exit(1)
		}
	} else {
		fmt.Println(sCode)
	}
}

func (i *Installer) GetAndroidSDKManager() {
	i.homepage = "https://developer.android.com/studio?hl=en"
	i.getDoc()
}

func (i *Installer) GetCygwinInstaller() {

}

func (i *Installer) GetMsys2Installer() {

}

func (i *Installer) GetRustInstaller() {

}

func (i *Installer) GetVSCode() {

}

func (i *Installer) GetMiniconda() {

}

func (i *Installer) FetchAll() {

}

func (i *Installer) Upload() {

}
