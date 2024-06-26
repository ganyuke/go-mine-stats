package config

import (
	"encoding/json"
	"log"
	"os"
	"slices"
)

var (
	Config_file *config_struct
)

func log_error(err error, context string) {
	if err != nil {
		log.Fatal(context, err)
	}
}

type api_config_struct struct {
	DefaultRespLimit int    `json:"default_response_limit"`
	MaxRespLimit     int    `json:"max_response_limit"`
	DefaultWorld     string `json:"default_world"`
	Port             string `json:"port"`
}

type blacklist_struct struct {
	ExOps bool     `json:"operators"`
	ExBan bool     `json:"banned"`
	List  []string `json:"list"`
}

type scan_config_struct struct {
	Enabled     bool             `json:"enabled"`
	PollSpeed   int              `json:"polling_speed"`
	Whitelist   bool             `json:"invert_blacklist"`
	Blacklist   blacklist_struct `json:"blacklist"`
	MojangFetch bool             `json:"fetch_mojang_usernames"`
}

type server_config_struct struct {
	ServerPath string `json:"server_path"`
	WorldName  string `json:"world_name"`
}

type config_struct struct {
	ServerList []server_config_struct `json:"server_list"`
	API        api_config_struct      `json:"api"`
	Scan       scan_config_struct     `json:"polling"`
}

type ConfigFile interface {
	CheckSanity() bool
	UpdateConfigBlacklist()
	CheckBlacklist() bool
}

func LoadConfig(path string) *config_struct {
	config_json, err := os.ReadFile(path)
	log_error(err, "Error while reading config file:")

	var config config_struct
	json.Unmarshal(config_json, &config)
	log_error(err, "Error while unmarshaling config file:")

	return &config
}

func (c *config_struct) CheckSanity() bool {
	var world_name_list []string
	for _, server_object := range c.ServerList {
		for _, world_name_exist := range world_name_list {
			if server_object.WorldName == world_name_exist {
				log.Fatal("Error: Duplicate 'world_names' detected! Please change the name of your world.")
			}
		}
		world_name_list = append(world_name_list, server_object.WorldName)
	}

	return true
}

type Username struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

func (c *config_struct) UpdateConfigBlacklist(list []Username) {
	for _, list := range list {
		c.Scan.Blacklist.List = append(Config_file.Scan.Blacklist.List, list.Uuid)
	}
}

func (c *config_struct) CheckBlacklist(uuid string) bool {
	return slices.Contains(c.Scan.Blacklist.List, uuid)
}
