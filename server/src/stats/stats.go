package stats

import (
	"encoding/json"
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func log_error(err error, context string) {
	if err != nil {
		log.Println(context, err)
	}
}

var (
	old_tracker *old_file_info
)

const (
	logFileDates = iota
	importStatistics
	checkImportStatistics
)

type player_statistics struct {
	Stats       map[string]map[string]int `json:"stats"`
	DataVersion int                       `json:"DataVersion"`
}

type old_file_info struct {
	size map[string]int64
	date map[string]time.Time
}

func collectPlayerStat(file_location, uuid, world string, date time.Time) player_statistics {

	println("Collecting stats for player " + uuid)

	file, err := os.ReadFile(file_location)
	log_error(err, "Error while reading player statistic file:")

	var player_stat_data player_statistics
	err = json.Unmarshal(file, &player_stat_data)
	log_error(err, "Error while unmarshaling player statistic file:")

	var new_stats []*db.Stat_item

	for category, items := range player_stat_data.Stats {
		for statistic, value := range items {
			player_stat := &db.Stat_item{
				Uuid:     uuid,
				Date:     date,
				Category: category,
				Item:     statistic,
				Value:    value,
				World:    world,
			}
			new_stats = append(new_stats, player_stat)
		}
	}

	err = db.UpdatePlayerStat(&db.Update_data{Statistics: new_stats})
	log_error(err, "An error has occured while updating stats: ")

	return player_stat_data

}

func loopStats(stats_directory, world_name string, directory_members []fs.DirEntry, operation int) {
player:
	for _, player_json := range directory_members {
		file_name := player_json.Name()
		extension := filepath.Ext(file_name)
		file_path := stats_directory + "/" + file_name

		if extension == ".json" {

			player_uuid := strings.Trim(file_name, extension)

			for _, uuid := range config.Config_file.Scan.Blacklist.List {
				if uuid == player_uuid && !config.Config_file.Scan.Whitelist {
					continue player
				} else if uuid != player_uuid && config.Config_file.Scan.Whitelist {
					continue player
				}
			}

			file_info, err := os.Stat(file_path)
			log_error(err, "Error while checking file information.")

			switch operation {
			case logFileDates:
				old_tracker.size[file_path] = file_info.Size()
				old_tracker.date[file_path] = file_info.ModTime()
			case importStatistics:
				old_tracker.size[file_path] = file_info.Size()
				old_tracker.date[file_path] = file_info.ModTime()
				date := file_info.ModTime().UTC().Round(time.Second)
				collectPlayerStat(file_path, player_uuid, world_name, date)
			case checkImportStatistics:
				if file_info.Size() != old_tracker.size[file_path] || file_info.ModTime() != old_tracker.date[file_path] {
					old_tracker.size[file_path] = file_info.Size()
					old_tracker.date[file_path] = file_info.ModTime()
					date := file_info.ModTime().UTC().Round(time.Second)
					collectPlayerStat(file_path, player_uuid, world_name, date)
				}
			}
		}
	}
}

func CollectAllStats(get_stats bool) {
	old_tracker = &old_file_info{size: map[string]int64{}, date: map[string]time.Time{}}
	for _, v := range config.Config_file.ServerList {
		var get_stats_for_world int
		if get_stats {
			get_stats_for_world = importStatistics
		} else {
			get_stats_for_world = logFileDates
		}

		println("Collecting stats for " + v.WorldName)

		if !get_stats && !db.GetWorld(v.WorldName) {
			get_stats_for_world = importStatistics
		}

		server_location, world_name := v.ServerPath, v.WorldName
		stats_directory := server_location + "/" + world_name + "/stats"
		player_stats, err := os.ReadDir(stats_directory)
		log_error(err, "Error while reading statistics directory:")

		loopStats(stats_directory, world_name, player_stats, get_stats_for_world)

		Poll_official.Monitor(stats_directory)
	}
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
