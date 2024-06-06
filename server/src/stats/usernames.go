package stats

import (
	"encoding/json"
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"go-mine-stats/src/stats/fetch"
	"io/fs"
	"log"
	"os"
)

func uniqueList(target, source []config.Username) []config.Username {
	if len(target) < 1 {
		return source
	}
loop:
	for _, obj := range source {
		for _, orig_obj := range target {
			if obj.Uuid == orig_obj.Uuid && (obj.Name == orig_obj.Name ||
				(obj.Name == "unknown" && orig_obj.Name != "unknown")) {
				continue loop
			}
		}
		target = append(target, obj)
	}
	return target
}

func updateUuidToUsernameList(serverDirectory string, namelessUuids []string) {
	var allPlayers []config.Username
	var missingUsernames []config.Username
	knownPlayers := db.GetPlayers()
	collectedPlayers, err := collectUsernames(serverDirectory)
	if err != nil {
		log.Println(err)
	}

nextPlayer:
	for _, uuid := range namelessUuids {
		for _, player := range collectedPlayers {
			if player.Uuid == uuid {
				continue nextPlayer
			}
		}
		missingUsernames = append(missingUsernames, config.Username{Uuid: uuid})
	}

	for _, knownPlayer := range knownPlayers {
		for i, missingPlayer := range missingUsernames {
			if missingPlayer.Uuid == knownPlayer.Uuid && knownPlayer.Name != "" {
				missingUsernames = append(missingUsernames[:i], missingUsernames[i+1:]...)
			}
		}
	}

	retrievedPlayers, failedPlayers := fetchMissing(missingUsernames)

	allPlayers = append(allPlayers, collectedPlayers...)
	allPlayers = append(allPlayers, retrievedPlayers...)
	allPlayers = append(allPlayers, failedPlayers...)

	db.InsertUsernames(allPlayers)
}

func collectUsernames(serverPath string) ([]config.Username, error) {
	var names []config.Username

	filesToUpdate, err := checkDates(serverPath)
	if err != nil {
		return nil, err
	}

	for _, filePath := range filesToUpdate {
		var data []config.Username
		if config.Config_file.Scan.Blacklist.ExOps && filePath == serverPath+"ops.json" {
			config.Config_file.UpdateConfigBlacklist(data)
		}
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &data)
		if err != nil {
			return nil, err
		}
		names = uniqueList(names, data)
	}

	if config.Config_file.Scan.Blacklist.ExBan {
		bans, err := os.ReadFile(serverPath + "banned-players.json")
		if err != nil {
			return nil, err
		}
		var banned_data []config.Username
		err = json.Unmarshal(bans, &banned_data)
		if err != nil {
			return nil, err
		}
		config.Config_file.UpdateConfigBlacklist(banned_data)
	}
	return names, nil
}

func fetchMissing(missingNames []config.Username) ([]config.Username, []config.Username) {
	var acquiredNames []config.Username
	var unknownNames []config.Username

	if config.Config_file.Scan.MojangFetch {
		for _, player := range missingNames {
			log.Println("Could not find username for " + player.Uuid + ". Fetching... ")
			playerInfo, err := fetch.FetchUsernameFromUUID(player.Uuid)
			if err != nil {
				log.Println("Error: recieved response " + err.Error() + " while fetching UUID " + player.Uuid)
				unknownNames = append(unknownNames, config.Username{Uuid: player.Uuid, Name: ""})
			}
			acquiredNames = append(acquiredNames, config.Username{Uuid: player.Uuid, Name: playerInfo.Name})
			log.Print("Username found: " + playerInfo.Name)
		}
	}
	return acquiredNames, unknownNames
}

func checkDates(path string) ([]string, error) {
	var files []fs.FileInfo
	var filesToUpdate []string

	userCache, err := os.Stat(path + "usercache.json")
	if err != nil {
		return nil, err
	}
	whitelist, err := os.Stat(path + "whitelist.json")
	if err != nil {
		return nil, err
	}
	ops, err := os.Stat(path + "ops.json")
	if err != nil {
		return nil, err
	}
	files = append(files, userCache, whitelist, ops)

	for _, fileInfo := range files {
		file_path := path + "/" + fileInfo.Name()
		if fileInfo.Size() != old_tracker.size[file_path] || fileInfo.ModTime() != old_tracker.date[file_path] {
			old_tracker.size[file_path] = fileInfo.Size()
			old_tracker.date[file_path] = fileInfo.ModTime()
			filesToUpdate = append(filesToUpdate, file_path)
		}
	}
	return filesToUpdate, nil
}
