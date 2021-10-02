package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jomei/notionapi"
	"notion-blog/pkg"
)

func filterFromConfig(config notion_blog.BlogConfig) *notionapi.PropertyFilter {
	if config.FilterProp != "" {
		if config.FilterValue == "" {
			log.Println("error: a value is needed to use a filter property")
			return nil
		}

		return &notionapi.PropertyFilter{
			Property: config.FilterProp,
			Select: &notionapi.SelectFilterCondition{
				Equals: config.FilterValue,
			},
		}
	}

	return nil
}

func generateArticleName(title string, date time.Time) string {
	return fmt.Sprintf(
		"%s_%s.md",
		date.Format("2006-01-02"),
		strings.ReplaceAll(
			strings.ToValidUTF8(
				strings.ToLower(title),
				"",
			),
			" ", "_",
		),
	)
}

func changeStatus(client *notionapi.Client, p notionapi.Page, config notion_blog.BlogConfig) {
	if config.FilterProp == "" || config.PublishedValue == "" {
		return
	}

	updatedProps := make(notionapi.Properties)
	updatedProps[config.FilterProp] = notionapi.SelectProperty{
		Select: notionapi.Option{
			Name: config.PublishedValue,
		},
	}

	_, err := client.Page.Update(context.Background(), notionapi.PageID(p.ID),
		&notionapi.PageUpdateRequest{
			Properties: updatedProps,
		},
	)
	if err != nil {
		log.Println("error changing status:", err)
	}
}

func ParseAndGenerate(config notion_blog.BlogConfig) {
	client := notionapi.NewClient(notionapi.Token(os.Getenv("NOTION_SECRET")))
	q, err := client.Database.Query(context.Background(), notionapi.DatabaseID(config.DatabaseID),
		&notionapi.DatabaseQueryRequest{
			PropertyFilter: filterFromConfig(config),
			PageSize:       100,
		})
	if err != nil {
		log.Fatalf("couldn't query articles database: %s", err)
	}

	err = os.MkdirAll(config.ContentFolder, 0777)
	if err != nil {
		log.Fatalf("couldn't create content folder: %s", err)
	}

	for _, res := range q.Results {
		title := notion_blog.ConvertRichText(res.Properties["Name"].(*notionapi.TitleProperty).Title)

		blocks, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(res.ID), &notionapi.Pagination{
			PageSize: 100,
		})
		if err != nil {
			log.Println("err:", err)
			continue
		}

		f, _ := os.Create(filepath.Join(
			config.ContentFolder,
			generateArticleName(title, res.CreatedTime),
		))

		notion_blog.GenerateHeader(f, res, config)
		notion_blog.Generate(f, blocks.Results, config)
		changeStatus(client, res, config)

		f.Close()
	}
}
