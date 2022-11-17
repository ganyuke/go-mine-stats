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

func collectPlayerStat(file_location, uuid, world string, date int64) player_statistics {

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

	db.UpdatePlayerStat(&db.Update_data{Statistics: new_stats})

	return player_stat_data

}

func CollectAllStats(get_stats bool) {

	log.Println("Collecting all stats...")

	for _, v := range config.Config_file.ServerList {

		println("Collecting stats for " + v.WorldName)

		server_location, world_name := v.ServerPath, v.WorldName
		stats_directory := server_location + "/" + world_name + "/stats"
		Poll_official.Monitor(stats_directory)
		player_stats, err := os.ReadDir(stats_directory)
		log_error(err, "Error while reading statistics directory:")
		old_tracker = &old_file_info{size: map[string]int64{}, date: map[string]time.Time{}}

		for _, player_json := range player_stats {
			file_name := player_json.Name()
			extension := filepath.Ext(file_name)
			file_path := stats_directory + "/" + file_name

			if extension == ".json" {

				for _, v := range config.Config_file.Scan.Blacklist {
					if v == file_path {
						return
					}
				}

				file_info, err := os.Stat(file_path)
				log_error(err, "Error while checking file information.")
				old_tracker.size[file_path] = file_info.Size()
				old_tracker.date[file_path] = file_info.ModTime()
				date := file_info.ModTime().Unix()

				if get_stats {
					println("Collecting stats for " + file_name)
					collectPlayerStat(file_path, strings.Trim(file_name, extension), world_name, date)
				}

			}

		}

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

	for _, v := range directory_members {
		player_stat_file := file_path + "/" + v.Name()
		extension := filepath.Ext(v.Name())
		if extension == ".json" {
			file_info, err := os.Stat(player_stat_file)
			log_error(err, "Error occured while comparing player stats JSON:")
			if file_info.Size() != old_tracker.size[player_stat_file] || file_info.ModTime() != old_tracker.date[player_stat_file] {
				old_tracker.size[player_stat_file] = file_info.Size()
				old_tracker.date[player_stat_file] = file_info.ModTime()
				date := file_info.ModTime().Unix()
				collectPlayerStat(player_stat_file, strings.Trim(v.Name(), ".json"), world_name, date)
				return
			}
		}
	}

}
