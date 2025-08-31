package buildinfo

import (
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
)

// Version is set by the linker.
var version string = ""

// AppName is set by the linker.
var appName string = ""

var BuildInfo *debug.BuildInfo

func init() {
	BuildInfo, _ = debug.ReadBuildInfo()
}

func Version() string {
	return version
}

func AppName() string {
	if appName == "" {
		return os.Args[0]
	}
	return appName
}

func ModulePath(fn any) string {
	value := reflect.ValueOf(fn)
	ptr := value.Pointer()
	ffp := runtime.FuncForPC(ptr)
	modulePath := path.Dir(path.Dir(ffp.Name()))

	return modulePath
}

func MainPath() string {
	return BuildInfo.Main.Path
}
