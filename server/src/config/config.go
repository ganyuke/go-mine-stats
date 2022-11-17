package config

import (
	"encoding/json"
	"log"
	"os"
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
}

type scan_config_struct struct {
	PollSpeed int      `json:"polling_speed"`
	Blacklist []string `json:"blacklist"`
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

func LoadConfig() *config_struct {
	config_json, err := os.ReadFile("../../config.json")
	log_error(err, "Error while reading config file:")

	var config config_struct
	json.Unmarshal(config_json, &config)
	log_error(err, "Error while unmarshaling config file:")

	return &config
}

func SanityCheck(config_file *config_struct) bool {
	var world_name_list []string
	for _, server_object := range config_file.ServerList {
		for _, world_name_exist := range world_name_list {
			if server_object.WorldName == world_name_exist {
				log.Fatal("Error: Duplicate 'world_names' detected! Please change the name of your world.")
			}
		}
		world_name_list = append(world_name_list, server_object.WorldName)
	}

	return true
}
