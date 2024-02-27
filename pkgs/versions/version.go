package versions

/*
collect version info for apps.
*/
import (
	"os"

	"github.com/gogf/gf/v2/util/gconv"
)

const (
	UseCNSourceEnv = "PC_USE_CN_Source"
)

// Uses resource from china or not.
func UseCNSource() bool {
	return gconv.Bool(os.Getenv(UseCNSourceEnv))
}

type Version struct {
	Tag     string `json,koanf:"tag"`
	Url     string `json,koanf:"url"`
	Arch    string `json,koanf:"arch"`
	Os      string `json,koanf:"os"`
	Sum     string `json,koanf:"sum"`
	SumType string `json,koanf:"sum_type"`
}

type IFetcher interface {
}
