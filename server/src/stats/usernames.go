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
			if obj.Uuid == orig_obj.Uuid && obj.Name == orig_obj.Name {
				continue loop
			}
		}
		target = append(target, obj)
	}
	return target
}

// func purgeList(target []config.Username, source []config.Username) []config.Username {
// 	for i, orig_obj := range target {
// 		for _, obj := range source {
// 			if obj == orig_obj {
// 				target = append(target[:i], target[i+1:]...)
// 			}
// 		}
// 	}
// 	return target
// }

func compareDb(database []config.Username, local []config.Username) []string {
	var missingUsernames []string
nextName:
	for _, databasePlayer := range database {
		for _, localPlayer := range local {
			if databasePlayer.Uuid == localPlayer.Uuid {
				continue nextName
			}
		}
		missingUsernames = append(missingUsernames, databasePlayer.Uuid)
	}
	return missingUsernames
}

func CollectUsernames() ([]config.Username, error) {
	var names []config.Username

	for _, serverDir := range config.Config_file.ServerList {
		path := serverDir.ServerPath
		filesToUpdate, err := checkDates(path)
		if err != nil {
			return nil, err
		}

		for _, filePath := range filesToUpdate {
			var data []config.Username
			if config.Config_file.Scan.Blacklist.ExOps && filePath == path+"/ops.json" {
				config.UpdateConfigBlacklist(data)
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
			bans, err := os.ReadFile(serverDir.ServerPath + "/banned-players.json")
			if err != nil {
				return nil, err
			}
			var banned_data []config.Username
			err = json.Unmarshal(bans, &banned_data)
			if err != nil {
				return nil, err
			}
			config.UpdateConfigBlacklist(banned_data)
		}
	}
	return names, nil
}

func pollUsernames() {
	names, err := CollectUsernames()
	if err != nil {
		log.Println(err)
	}
	names = FetchMissing(names)
	err = db.InsertUsernames(names)
	if err != nil {
		log.Println(err)
	}
}

func FetchMissing(names []config.Username) []config.Username {
	if config.Config_file.Scan.MojangFetch {
		missingNames := compareDb(db.GetUuidsFromStats(), names)
		databaseNames := compareDb(db.GetUuidsFromUsernames(), names)
		if len(missingNames) != len(databaseNames) {
			for _, uuid := range missingNames {
				playerInfo, err := fetch.FetchUsernameFromUUID(uuid)
				if err != nil {
					log.Println(err)
				}
				names = append(names, config.Username{Uuid: playerInfo.Id, Name: playerInfo.Name})
			}
		}
	}
	return names
}

func checkDates(path string) ([]string, error) {
	var files []fs.FileInfo
	var filesToUpdate []string

	userCache, err := os.Stat(path + "/usercache.json")
	if err != nil {
		return nil, err
	}
	whitelist, err := os.Stat(path + "/whitelist.json")
	if err != nil {
		return nil, err
	}
	ops, err := os.Stat(path + "/ops.json")
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
