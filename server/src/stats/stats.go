package stats

import (
	"encoding/json"
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func log_error(err error, context string) {
	if err != nil {
		log.Fatal(context, err)
	}
}

var (
	old_tracker *old_file_info
)

type player_statistics struct {
	Stats       map[string]map[string]int `json:"stats"`
	DataVersion int                       `json:"DataVersion"`
}

type old_file_info struct {
	size map[string]int64
	date map[string]time.Time
}

func collect_player_stat(file_location string, uuid string) player_statistics {

	file, err := os.ReadFile(file_location)
	log_error(err, "Error while reading player statistic file:")

	var player_stat_data player_statistics
	err = json.Unmarshal(file, &player_stat_data)
	log_error(err, "Error while unmarshaling player statistic file:")

	for category, items := range player_stat_data.Stats {
		for statistic, value := range items {
			db.Update_player_stat(uuid, category, statistic, value)
		}
	}

	return player_stat_data

}

func Collect_all_stats(get_stats bool) {

	log.Println("Collecting all stats...")

	server_location, world_name := config.Config_file.ServerPath, config.Config_file.WorldName
	stats_directory := server_location + "/" + world_name + "/stats"
	player_stats, err := os.ReadDir(stats_directory)
	log_error(err, "Error while reading statistics directory:")

	old_tracker = &old_file_info{size: map[string]int64{}, date: map[string]time.Time{}}

	for _, player_json := range player_stats {
		file_name := player_json.Name()
		extension := filepath.Ext(file_name)
		file_path := stats_directory + "/" + file_name
		if extension == ".json" {
			file_info, err := os.Stat(file_path)
			log_error(err, "Error while checking file information.")
			old_tracker.size[file_path] = file_info.Size()
			old_tracker.date[file_path] = file_info.ModTime()
			Poll_official.Monitor(file_path)
			if get_stats {
				collect_player_stat(file_path, strings.Trim(file_name, extension))
			}
		}
	}
}

func Poll_json(file_path string) {

	file_info, err := os.Stat(file_path)

	if err != nil {
		Poll_official.Remove(file_path)
		log.Println("File " + file_path + " no longer found. Removing from monitoring.")
		return
	}

	server_location, world_name := config.Config_file.ServerPath, config.Config_file.WorldName
	stats_directory := server_location + "/" + world_name + "/stats"
	extension := filepath.Ext(file_path)

	if file_info.Size() != old_tracker.size[file_path] || file_info.ModTime() != old_tracker.date[file_path] {
		old_tracker.size[file_path] = file_info.Size()
		old_tracker.date[file_path] = file_info.ModTime()
		collect_player_stat(file_path, strings.Trim(strings.Trim(file_path, stats_directory), extension))
		return
	}

}
