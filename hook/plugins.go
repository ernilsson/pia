package hook

import (
	"io/fs"
	"path/filepath"
	"plugin"
)

func list(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".so" {
			files = append(files, s)
		}
		return nil
	})
	return files, err
}

func load(dir string) ([]*plugin.Plugin, error) {
	files, err := list(dir)
	if err != nil {
		return nil, err
	}
	plugs := make([]*plugin.Plugin, 0)
	for _, file := range files {
		p, err := plugin.Open(file)
		if err != nil {
			return nil, err
		}
		plugs = append(plugs, p)
	}
	return plugs, nil
}
