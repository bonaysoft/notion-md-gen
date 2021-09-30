package main

import (
	"flag"
	"log"

	"notion-blog/internal"
	notion_blog "notion-blog/pkg"

	"github.com/itzg/go-flagsfiller"
	"github.com/joho/godotenv"
)

var config notion_blog.BlogConfig

func parseFlagsConfig() {
	// create a FlagSetFiller
	filler := flagsfiller.New()
	// fill and map struct fields to flags
	err := filler.Fill(flag.CommandLine, &config)
	if err != nil {
		log.Fatal(err)
	}

	// parse command-line like usual
	flag.Parse()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file provided")
	}

	parseFlagsConfig()

	internal.ParseAndGenerate(config)
}
