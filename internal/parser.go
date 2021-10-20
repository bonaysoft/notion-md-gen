package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	notion_blog "notion-blog/pkg"

	"github.com/janeczku/go-spinner"
	"github.com/jomei/notionapi"
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

func recursiveGetChildren(client *notionapi.Client, blockID notionapi.BlockID) (blocks []notionapi.Block, err error) {
	res, err := client.Block.GetChildren(context.Background(), blockID, &notionapi.Pagination{
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	blocks = res.Results
	if len(blocks) == 0 {
		return
	}

	for _, block := range blocks {
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			b.Paragraph.Children, err = recursiveGetChildren(client, b.ID)
		case *notionapi.CalloutBlock:
			b.Callout.Children, err = recursiveGetChildren(client, b.ID)
		case *notionapi.QuoteBlock:
			b.Quote.Children, err = recursiveGetChildren(client, b.ID)
		case *notionapi.BulletedListItemBlock:
			b.BulletedListItem.Children, err = recursiveGetChildren(client, b.ID)
		case *notionapi.NumberedListItemBlock:
			b.NumberedListItem.Children, err = recursiveGetChildren(client, b.ID)
		}

		if err != nil {
			return
		}
	}

	return
}

func ParseAndGenerate(config notion_blog.BlogConfig) error {
	client := notionapi.NewClient(notionapi.Token(os.Getenv("NOTION_SECRET")))

	spin := spinner.StartNew("Querying Notion database")
	q, err := client.Database.Query(context.Background(), notionapi.DatabaseID(config.DatabaseID),
		&notionapi.DatabaseQueryRequest{
			PropertyFilter: filterFromConfig(config),
			PageSize:       100,
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

	for i, res := range q.Results {
		title := notion_blog.ConvertRichText(res.Properties["Name"].(*notionapi.TitleProperty).Title)

		fmt.Printf("-- Article [%d/%d] --\n", i+1, len(q.Results))
		spin = spinner.StartNew("Getting blocks tree")
		// Get page blocks tree
		blocks, err := recursiveGetChildren(client, notionapi.BlockID(res.ID))
		spin.Stop()
		if err != nil {
			log.Println("❌ Getting blocks tree:", err)
			continue
		}
		fmt.Println("✔ Getting blocks tree: Completed")

		// Create file
		f, _ := os.Create(filepath.Join(
			config.ContentFolder,
			generateArticleName(title, res.CreatedTime),
		))

		// Generate and dump content to file
		if err := notion_blog.Generate(f, res, blocks, config); err != nil {
			fmt.Println("❌ Generating blog post:", err)
			f.Close()
			continue
		}
		fmt.Println("✔ Generating blog post: Completed")

		// Change status of blog post if desired
		changeStatus(client, res, config)

		f.Close()
	}

	return nil
}
