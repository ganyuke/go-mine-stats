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
	DefaultRespLimit int `json:"default_response_limit"`
	MaxRespLimit     int `json:"max_response_limit"`
}

type config_struct struct {
	ServerPath string            `json:"server_path"`
	WorldName  string            `json:"world_name"`
	PollSpeed  int               `json:"polling_speed"`
	API        api_config_struct `json:"api"`
}

func Load_config() *config_struct {
	config_json, err := os.ReadFile("../../config.json")
	log_error(err, "Error while reading config file:")

	var config config_struct
	json.Unmarshal(config_json, &config)
	log_error(err, "Error while unmarshaling config file:")

	return &config
}
