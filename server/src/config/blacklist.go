package config

type Username struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

func UpdateConfigBlacklist(list []Username) {
	for _, list := range list {
		Config_file.Scan.Blacklist.List = append(Config_file.Scan.Blacklist.List, list.Uuid)
	}
}

func CheckBlacklist(uuid string) bool {
	for _, blacklistedUuid := range Config_file.Scan.Blacklist.List {
		if blacklistedUuid == uuid {
			return true
		} else {
			return false
		}
	}
	return false
}
