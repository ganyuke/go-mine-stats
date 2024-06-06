package stats

import (
	"io/fs"
	"log"
	"os"
	"strings"
	"time"
)

var (
	old_tracker *old_file_info
)

type old_file_info struct {
	size map[string]int64
	date map[string]time.Time
}

func (metadata *old_file_info) updateRecord(filePath string, fileInfo fs.FileInfo) time.Time {
	metadata.size[filePath] = fileInfo.Size()
	metadata.date[filePath] = fileInfo.ModTime()
	date := fileInfo.ModTime().UTC().Round(time.Second)
	return date
}

func InitFileTracking() {
	old_tracker = &old_file_info{size: map[string]int64{}, date: map[string]time.Time{}}
}

func pollDir(file_path string) {
	directory_members, err := os.ReadDir(file_path)
	concat := strings.Split(file_path, "/")
	world_name := concat[len(concat)-2]

	if err != nil {
		Poll_official.Remove(file_path)
		log.Println("Directory " + file_path + " no longer found. Removing from monitoring.")
		return
	}

	loopStats(file_path, world_name, directory_members, checkImportStatistics)
}
