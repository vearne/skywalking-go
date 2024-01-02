package prometheus

import (
	"embed"
	"strings"

	"github.com/apache/skywalking-go/plugins/core/instrument"
)

//go:embed *
var fs embed.FS

//skywalking:nocopy
type Instrument struct {
}

func NewInstrument() *Instrument {
	return &Instrument{}
}

func (i *Instrument) Name() string {
	return "prometheus"
}

func (i *Instrument) BasePackage() string {
	return "github.com/prometheus/client_golang"
}

func (i *Instrument) VersionChecker(version string) bool {
	return strings.HasPrefix(version, "v1")
}

func (i *Instrument) Points() []*instrument.Point {
	return []*instrument.Point{
		{
			PackagePath: "prometheus",
			At: instrument.NewStaticMethodEnhance("NewRegistry",
				instrument.WithResultCount(1),
				instrument.WithResultType(0, "*Registry")),
			Interceptor: "RegistryInterceptor",
		},
	}
}

func (i *Instrument) FS() *embed.FS {
	return &fs
}
