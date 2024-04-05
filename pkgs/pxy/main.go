package main

import (
	"os"

	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/versions"
)

func main() {
	// sites.TestEDomains()
	// sites.TestEDCollector()
	// sites.TestTDomains()
	os.Setenv(confs.ToEnableProxyEnvName, "true")
	// app := NewApp()
	// app.Run()

	cfg := confs.NewCollectorConf()

	gr := versions.NewGroovy(cfg)
	gr.FetchAll()
	gr.Upload()

	// pypy := versions.NewPyPy(cfg)
	// pypy.FetchAll()
	// pypy.Upload()

	// aj := versions.NewAdoptiumJDK(cfg)
	// aj.FetchAll()
	// aj.Upload()

	// gl := versions.NewGolang(cfg)
	// gl.FetchAll()
	// gl.Upload()

	// jdk := versions.NewJDK(cfg)
	// jdk.FetchAll()
	// jdk.Upload()

	// gra := versions.NewGradle(cfg)
	// gra.FetchAll()
	// gra.Upload()

	// maven := versions.NewMaven(cfg)
	// maven.FetchAll()
	// maven.Upload()

	// node := versions.NewNodejs(cfg)
	// node.FetchAll()
	// node.Upload()

	// zig := versions.NewZig(cfg)
	// zig.FetchAll()
	// zig.Upload()

	// julia := versions.NewJulia(cfg)
	// julia.FetchAll()
	// julia.Upload()

	// flutter := versions.NewFlutter(cfg)
	// flutter.FetchAll()
	// flutter.Upload()

	// py := versions.NewPython(cfg)
	// py.FetchAll()
	// py.Upload()

	// php := versions.NewPhP(cfg)
	// php.FetchAll()
	// php.Upload()

	// ins := versions.NewInstaller(cfg)
	// ins.GetAndroidSDKManager()
	// ins.GetVSCode()
	// ins.GetMiniconda()
	// ins.FetchAll()
	// ins.Upload()

	// gh := versions.NewGithubRepo(cfg)
	// gh.FetchAll()
	// gh.Upload()

	// scala := versions.NewScala(cfg)
	// scala.FetchAll()
	// scala.Upload()

	// dn := versions.NewDotNet(cfg)
	// dn.FetchAll()
	// dn.Upload()

	// kctl := versions.NewKubectl(cfg)
	// kctl.FetchAll()
	// kctl.Upload()
}
