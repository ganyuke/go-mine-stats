package stats

import (
	"encoding/json"
	"go-mine-stats/src/config"
	"os"
)

type username struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

func uniqueList(target, source []username) []username {
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

func CollectUsernames() (error, []username) {
	var names []username

	for _, serverDir := range config.Config_file.ServerList {
		userCache, err := os.ReadFile(serverDir.ServerPath + "/usercache.json")
		if err != nil {
			return err, nil
		}
		whitelist, err := os.ReadFile(serverDir.ServerPath + "/whitelist.json")
		if err != nil {
			return err, nil
		}
		ops, err := os.ReadFile(serverDir.ServerPath + "/ops.json")
		if err != nil {
			return err, nil
		}

		var user_cache_data []username
		err = json.Unmarshal(userCache, &user_cache_data)
		if err != nil {
			return err, nil
		}
		var whitelist_data []username
		err = json.Unmarshal(whitelist, &whitelist_data)
		if err != nil {
			return err, nil
		}
		var ops_data []username
		err = json.Unmarshal(ops, &ops_data)
		if err != nil {
			return err, nil
		}

		names = user_cache_data

		uniqueList(names, whitelist_data)
		uniqueList(names, ops_data)

		for _, obj := range names {
			println(obj.Name)
			println(obj.Uuid)
		}
	}

	return nil, names
}
