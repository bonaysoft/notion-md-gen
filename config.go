package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ImagesLink    string
	ImagesFolder  string
	ContentFolder string
}

var config Config

func parseConfig() {
	content, err := os.ReadFile("notionblog.config.json")
	if err != nil {
		log.Fatal("error reading config file: ", err)
	}
	json.Unmarshal(content, &config)
}
