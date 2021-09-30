package main

import (
	"encoding/json"
	"log"
	"os"

	"notion-blog/internal"
	notion_blog "notion-blog/pkg"

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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	parseJSONConfig()

	internal.ParseAndGenerate(config)
}
