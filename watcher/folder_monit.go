package watcher

import (
	"ezpz_uploader/utube"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func Folder_watcher(dir string) {

	youtubeService, err := utube.YouTubeSetup()
	if err != nil {
		fmt.Println(err)
		return

	}

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
						folder_slices := strings.Split(event.Name, "\\")
						folder_name := folder_slices[len(folder_slices)-1]
						fmt.Println(folder_name)
						utube.CreatePlaylist(youtubeService, folder_name, folder_name)

						addDirToWatcher(event.Name, watcher)
					} else {
						folder_slices := strings.Split(event.Name, "\\")
						folder_name := folder_slices[len(folder_slices)-1]
						playlist_id, err := utube.CreatePlaylist(youtubeService, folder_name, folder_name)
						if err != nil {
							fmt.Println(err)
							return
						}
						video_id := utube.UploadVideo(youtubeService, getPathWithFile(event.Name), getTitleFromFileName(event.Name, "."), "", []string{"a"})
						fmt.Println(playlist_id)
						err = utube.AddVideoToPlaylist(youtubeService, playlist_id, video_id)
						if err != nil {
							fmt.Println(err)
							return
						}
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

	root := dir
	err = addDirToWatcher(root, watcher)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func getFolderName(path string) string {
	// Use filepath.Dir to get the directory part
	dir := filepath.Dir(path)

	// Use filepath.Base to get the last element of the path, which is the folder name
	folderName := filepath.Base(dir)

	// If on Windows and there are backslashes in the path, clean up to remove them
	folderName = cleanFolderName(folderName)
	return folderName
}

func cleanFolderName(folderName string) string {
	// If on Windows, replace backslashes with slashes and extract last part
	folderName = strings.ReplaceAll(folderName, "\\", "/")
	parts := strings.Split(folderName, "/")
	return parts[len(parts)-1]
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

func getPathWithFile(path string) string {
	// Use filepath.Dir to get the directory part
	dir := filepath.Dir(path)

	// Use filepath.Base to get the last element of the path, which is the file name
	fileName := filepath.Base(path)

	// Concatenate dir and fileName to get the path with the file name
	pathWithFile := filepath.Join(dir, fileName)

	return pathWithFile
}

func getTitleFromFileName(fileName string, delimiter string) string {
	// Split the fileName by delimiter
	parts := strings.Split(fileName, delimiter)

	// Assume the title is the first part (adjust as per your naming convention)
	title := parts[0]

	return title
}
