package versions

/*
collect version info for apps.
*/
import (
	"os"
	"regexp"

	"github.com/gogf/gf/v2/util/gconv"
)

const (
	UseCNSourceEnv = "PC_USE_CN_SOURCE"
)

var (
	VersionPattern = regexp.MustCompile(`(\d+\.\d+\.\d+)`)
)

func FindVersion(s string) (vName string) {
	vName = VersionPattern.FindString(s)
	if vName != "" {
		return
	}
	return s
}

// Uses resource from china or not.
func UseCNSource() bool {
	return gconv.Bool(os.Getenv(UseCNSourceEnv))
}

type VFile struct {
	Url     string `json,koanf:"url"`
	Arch    string `json,koanf:"arch"`
	Os      string `json,koanf:"os"`
	Sum     string `json,koanf:"sum"`
	SumType string `json,koanf:"sum_type"`
	Extra   string `json,koanf:"extra"`
}

type VFileList []*VFile

type Versions map[string]VFileList

type IFetcher interface {
}
