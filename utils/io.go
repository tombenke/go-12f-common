package utils

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"

	yaml "github.com/goccy/go-yaml"
)

func SaveServerConfig(serverConfig any, dirName string, fileName string) (string, error) {
	configFile := filepath.Join(dirName, fileName)
	file, err := os.Create(configFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(serverConfig); err != nil {
		return "", err
	}

	return configFile, nil
}

func RenderServerConfig(serverConfig any) (string, error) {
	yamlData, err := yaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

func ModulePath(fn any) string {
	value := reflect.ValueOf(fn)
	ptr := value.Pointer()
	ffp := runtime.FuncForPC(ptr)
	modulePath := path.Dir(path.Dir(ffp.Name()))

	return modulePath
}
