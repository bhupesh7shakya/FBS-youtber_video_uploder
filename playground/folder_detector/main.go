package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// also write not needted
				// fmt.Println("event:", event)
				// if event.Op&fsnotify.Write == fsnotify.Write {
				// 	fmt.Println("modified file:", event.Name)
				// 	checkFileOrFolder(event.Name)
				// }
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("created file:", event.Name)
					checkFileOrFolder(event.Name)
					// Add the new directory to the watcher if it's a directory
					if isDir(event.Name) {
						// write code for creating youtube playlist
						addDirToWatcher(event.Name, watcher)
					}
				}
				// remove lisenteer not need
				// if event.Op&fsnotify.Remove == fsnotify.Remove {
				// 	fmt.Println("deleted file:", event.Name)
				// 	// Optionally remove the directory from the watcher if needed
				// }
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	root := "E:\\go lang\\go_youtube\\folders"
	err = addDirToWatcher(root, watcher)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func checkFileOrFolder(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	if fileInfo.IsDir() {
		fmt.Println("Yes! It's a DIR..")
		return true
	}
	fmt.Println("No! It's a File")
	return false
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func addDirToWatcher(path string, watcher *fsnotify.Watcher) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Println("Error adding directory to watcher:", err)
				return err
			}
			fmt.Println("Watching directory:", path)
		}
		return nil
	})
}
