package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	notion_blog "notion-md-gen/pkg"
	"notion-md-gen/pkg/tomarkdown"

	"github.com/dstotijn/go-notion"
	"github.com/janeczku/go-spinner"
)

func filterFromConfig(config notion_blog.BlogConfig) *notion.DatabaseQueryFilter {
	if config.FilterProp == "" || len(config.FilterValue) == 0 {
		return nil
	}

	properties := make([]notion.DatabaseQueryFilter, len(config.FilterValue))

	for i, val := range config.FilterValue {
		properties[i] = notion.DatabaseQueryFilter{
			Property: config.FilterProp,
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: val,
			},
		}
	}

	return &notion.DatabaseQueryFilter{
		Or: properties,
	}
}

func generateArticleName(title string, date time.Time, config notion_blog.BlogConfig) string {
	escapedTitle := strings.ReplaceAll(
		strings.ToValidUTF8(
			strings.ToLower(title),
			"",
		),
		" ", "_",
	)
	escapedFilename := escapedTitle + ".md"

	if config.UseDateForFilename {
		// Add date to the name to allow repeated titles
		return date.Format("2006-01-02") + escapedFilename
	}
	return escapedFilename
}

// chageStatus changes the Notion article status to the published value if set.
// It returns true if status changed.
func changeStatus(client *notion.Client, p notion.Page, config notion_blog.BlogConfig) bool {
	// No published value or filter prop to change
	if config.FilterProp == "" || config.PublishedValue == "" {
		return false
	}

	if v, ok := p.Properties.(notion.DatabasePageProperties)[config.FilterProp]; ok {
		if v.Select.Name == config.PublishedValue {
			return false
		}
	} else { // No filter prop in page, can't change it
		return false
	}

	updatedProps := make(notion.DatabasePageProperties)
	updatedProps[config.FilterProp] = notion.DatabasePageProperty{
		Select: &notion.SelectOptions{
			Name: config.PublishedValue,
		},
	}

	_, err := client.UpdatePage(context.Background(), p.ID,
		notion.UpdatePageParams{
			DatabasePageProperties: &updatedProps,
		},
	)
	if err != nil {
		log.Println("error changing status:", err)
	}

	return err == nil
}

func recursiveGetChildren(client *notion.Client, blockID string) (blocks []notion.Block, err error) {
	res, err := client.FindBlockChildrenByID(context.Background(), blockID, &notion.PaginationQuery{
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	blocks = res.Results
	if len(blocks) == 0 {
		return
	}

	for _, block := range res.Results {
		if !block.HasChildren {
			continue
		}

		switch block.Type {
		case notion.BlockTypeParagraph:
			block.Paragraph.Children, err = recursiveGetChildren(client, block.ID)
		case notion.BlockTypeCallout:
			block.Callout.Children, err = recursiveGetChildren(client, block.ID)
		case notion.BlockTypeQuote:
			block.Quote.Children, err = recursiveGetChildren(client, block.ID)
		case notion.BlockTypeBulletedListItem:
			block.BulletedListItem.Children, err = recursiveGetChildren(client, block.ID)
		case notion.BlockTypeNumberedListItem:
			block.NumberedListItem.Children, err = recursiveGetChildren(client, block.ID)
		case notion.BlockTypeTable:
			block.Table.Children, err = recursiveGetChildren(client, block.ID)
		}

		if err != nil {
			return
		}
	}

	return blocks, nil
}

func ParseAndGenerate(config notion_blog.BlogConfig) error {
	// client := notionapi.NewClient(notionapi.Token(os.Getenv("NOTION_SECRET")))
	client := notion.NewClient(os.Getenv("NOTION_SECRET"))

	spin := spinner.StartNew("Querying Notion database")
	q, err := client.QueryDatabase(context.Background(), config.DatabaseID,
		&notion.DatabaseQuery{
			Filter:   filterFromConfig(config),
			PageSize: 100,
		})
	spin.Stop()
	if err != nil {
		return fmt.Errorf("❌ Querying Notion database: %s", err)
	}
	fmt.Println("✔ Querying Notion database: Completed")

	err = os.MkdirAll(config.ContentFolder, 0777)
	if err != nil {
		return fmt.Errorf("couldn't create content folder: %s", err)
	}

	// number of article status changed
	changed := 0

	for i, res := range q.Results {
		title := tomarkdown.ConvertRichText(res.Properties.(notion.DatabasePageProperties)["Name"].Title)

		fmt.Printf("-- Article [%d/%d] --\n", i+1, len(q.Results))
		spin = spinner.StartNew("Getting blocks tree")
		// Get page blocks tree
		blocks, err := recursiveGetChildren(client, res.ID)
		spin.Stop()
		if err != nil {
			log.Println("❌ Getting blocks tree:", err)
			continue
		}
		fmt.Println("✔ Getting blocks tree: Completed")

		// Create file
		f, _ := os.Create(filepath.Join(
			config.ContentFolder,
			generateArticleName(title, res.CreatedTime, config),
		))

		// Generate and dump content to file
		if err := notion_blog.Generate(f, res, blocks, config); err != nil {
			fmt.Println("❌ Generating blog post:", err)
			f.Close()
			continue
		}
		fmt.Println("✔ Generating blog post: Completed")

		// Change status of blog post if desired
		if changeStatus(client, res, config) {
			changed++
		}

		f.Close()
	}

	// Set GITHUB_ACTIONS info variables
	// https://docs.github.com/en/actions/learn-github-actions/workflow-commands-for-github-actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Printf("::set-output name=articles_published::%d\n", changed)
	}

	return nil
}
