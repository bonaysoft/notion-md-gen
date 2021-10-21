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

func filterFromConfig(config notion_blog.BlogConfig) *notionapi.CompoundFilter {
	if config.FilterProp == "" || len(config.FilterValue) == 0 {
		return nil
	}

	properties := make([]notionapi.PropertyFilter, len(config.FilterValue))

	for i, val := range config.FilterValue {
		properties[i] = notionapi.PropertyFilter{
			Property: config.FilterProp,
			Select: &notionapi.SelectFilterCondition{
				Equals: val,
			},
		}
	}

	return &notionapi.CompoundFilter{
		notionapi.FilterOperatorOR: properties,
	}
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

// chageStatus changes the Notion article status to the published value if set.
// It returns true if status changed.
func changeStatus(client *notionapi.Client, p notionapi.Page, config notion_blog.BlogConfig) bool {
	// No published value or filter prop to change
	if config.FilterProp == "" || config.PublishedValue == "" {
		return false
	}

	if v, ok := p.Properties[config.FilterProp]; ok {
		if status, ok := v.(*notionapi.SelectProperty); ok {
			// Already has published value
			if status.Select.Name == config.PublishedValue {
				return false
			}
		} else { // Filter prop is not a select property
			return false
		}
	} else { // No filter prop in page, can't change it
		return false
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

	return err == nil
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
			CompoundFilter: filterFromConfig(config),
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

	// number of article status changed
	changed := 0

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
