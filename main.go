package main

import (
	"encoding/json"
	"log"
	"os"

	"notion-md-gen/internal"
	notion_blog "notion-md-gen/pkg"

	"github.com/joho/godotenv"
)

var config notion_blog.BlogConfig

func parseJSONConfig() {
	content, err := os.ReadFile("notionblog.config.json")
	if err != nil {
		log.Fatal("error reading config file: ", err)
	}
	json.Unmarshal(content, &config)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file provided")
	}

	parseJSONConfig()
	if err := internal.ParseAndGenerate(config); err != nil {
		log.Println(err)
	}
}
