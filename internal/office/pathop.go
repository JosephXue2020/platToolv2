package office

import (
	"os"
	"path"
	"path/filepath"
)

// GetPathInfo function collects all the files in root directory
func GetPathInfo(direc string) ([][]string, error) {
	var fInfo [][]string

	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var inSli []string
		if info.IsDir() {
			return nil
		} else {
			_, fname := filepath.Split(p)
			ext := path.Ext(fname)
			inSli = []string{fname, ext, p}
		}
		fInfo = append(fInfo, inSli)
		return nil
	}

	err := filepath.Walk(direc, walkFunc)
	return fInfo, err
}
