package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/adrg/xdg"
)

type File struct {
	Config
	Path string
}

func (file File) Save() error {
	f, err := os.OpenFile(file.Path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	err = enc.Encode(file.Config)
	return errors.Join(err, f.Close())
}

func DefaultPath(appname string) (string, error) {
	return xdg.ConfigFile(appname + "/config.json")
}

func getPath(appname, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return DefaultPath(appname)
}

func Load(appname, path string) (file File, err error) {
	file.Path, err = getPath(appname, path)
	if err != nil {
		return
	}
	flags := os.O_RDONLY
	if path == "" {
		flags |= os.O_CREATE
	}
	f, err := os.OpenFile(file.Path, flags, 0644)
	if err != nil {
		return
	}

	err = json.NewDecoder(f).Decode(&file.Config)
	return file, errors.Join(err, f.Close())
}
