package buildinfo

import (
	srv_utils "github.com/tombenke/go-12f-common/utils"
)

// Version is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var Version string = "0.0.1"

// BuildTime is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var BuildTime string = "2021-09-01T12:00:00Z"

// AppName is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var AppName string = "scheduler_otel"

type BuildInfoApp struct{}

func (b *BuildInfoApp) Version() string {
	return Version
}

func (b *BuildInfoApp) BuildTime() string {
	return BuildTime
}

func (b *BuildInfoApp) AppName() string {
	if AppName == "" {
		return "UNKNOWN"
	}
	return AppName
}

func (b *BuildInfoApp) ModulePath() string {
	return srv_utils.ModulePath(b.ModulePath)
	//return modulePath(b.ModulePath)
}

// func modulePath(fn any) string {
// 	value := reflect.ValueOf(fn)
// 	ptr := value.Pointer()
// 	ffp := runtime.FuncForPC(ptr)
// 	modulePath := path.Dir(path.Dir(ffp.Name()))

// 	return modulePath
// }

var BuildInfo = &BuildInfoApp{}
