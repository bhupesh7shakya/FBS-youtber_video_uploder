package main

import (
	"ezpz_uploader/watcher"
	"os"
	"path/filepath"
)

func main() {
	dir := "chokidar_re"
	os.Mkdir(dir, 0755)
	filepath, err := filepath.Abs(dir)
	if err != nil {
		panic("something went wrong")
	}
	watcher.Folder_watcher(filepath)
}
