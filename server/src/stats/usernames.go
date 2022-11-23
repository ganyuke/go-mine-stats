package stats

import (
	"encoding/json"
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"os"
)

func updateConfigBlacklist(list []db.Username) {
	for _, list := range list {
		config.Config_file.Scan.Blacklist.List = append(config.Config_file.Scan.Blacklist.List, list.Uuid)
	}
}

func uniqueList(target, source []db.Username) []db.Username {
	for _, orig_obj := range target {
		for _, obj := range source {
			if obj == orig_obj {
				break
			}
			target = append(target, obj)
		}
	}
	return target
}

func purgeList(target []db.Username, source []db.Username) []db.Username {
	for i, orig_obj := range target {
		for _, obj := range source {
			if obj == orig_obj {
				target = append(target[:i], target[i+1:]...)
			}
		}
	}
	return target
}

func CollectUsernames() ([]db.Username, error) {
	var names []db.Username

	for _, serverDir := range config.Config_file.ServerList {
		userCache, err := os.ReadFile(serverDir.ServerPath + "/usercache.json")
		if err != nil {
			return nil, err
		}
		whitelist, err := os.ReadFile(serverDir.ServerPath + "/whitelist.json")
		if err != nil {
			return nil, err
		}
		ops, err := os.ReadFile(serverDir.ServerPath + "/ops.json")
		if err != nil {
			return nil, err
		}

		var user_cache_data []db.Username
		err = json.Unmarshal(userCache, &user_cache_data)
		if err != nil {
			return nil, err
		}
		var whitelist_data []db.Username
		err = json.Unmarshal(whitelist, &whitelist_data)
		if err != nil {
			return nil, err
		}
		var ops_data []db.Username
		err = json.Unmarshal(ops, &ops_data)
		if err != nil {
			return nil, err
		}

		names = user_cache_data

		uniqueList(names, whitelist_data)
		if config.Config_file.Scan.Blacklist.ExOps {
			purgeList(names, ops_data)
			updateConfigBlacklist(ops_data)
		} else {
			uniqueList(names, ops_data)
		}

		if config.Config_file.Scan.Blacklist.ExBan {
			bans, err := os.ReadFile(serverDir.ServerPath + "/banned-players.json")
			if err != nil {
				return nil, err
			}
			var banned_data []db.Username
			err = json.Unmarshal(bans, &banned_data)
			if err != nil {
				return nil, err
			}
			purgeList(names, banned_data)
			updateConfigBlacklist(banned_data)
		}
	}
	return names, nil
}
