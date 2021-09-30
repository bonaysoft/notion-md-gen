package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

const (
	DATABASE = "f887dfac795547ff97a81bb669b1052f"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := notionapi.NewClient(notionapi.Token(os.Getenv("NOTION_SECRET")))
	q, _ := client.Database.Query(context.Background(), DATABASE, &notionapi.DatabaseQueryRequest{
		PropertyFilter: &notionapi.PropertyFilter{
			Property: "Status",
			Select: &notionapi.SelectFilterCondition{
				Equals: "Finished ✅",
			},
		},
		Sorts: []notionapi.SortObject{
			{
				Timestamp: notionapi.TimestampCreated,
				Direction: notionapi.SortOrderDESC,
			},
		},
		PageSize: 100,
	})

	for _, res := range q.Results {
		blocks, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(res.ID), &notionapi.Pagination{
			PageSize: 100,
		})
		if err != nil {
			log.Println("err:", err)
			continue
		}

		f, _ := os.Create("test.md")

		// for _, block := range blocks.Results {
		// 	log.Println(block.GetType())
		// }
		GenerateHeader(f, res)
		Generate(f, blocks.Results)

		f.Close()
		break
	}
}
