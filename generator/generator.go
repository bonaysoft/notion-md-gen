package generator

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bonaysoft/notion-md-gen/pkg/tomarkdown"

	"github.com/dstotijn/go-notion"
)

func Run(config Config) error {
	if err := os.MkdirAll(config.Markdown.PostSavePath, 0755); err != nil {
		return fmt.Errorf("couldn't create content folder: %s", err)
	}

	// find database page
	client := notion.NewClient(os.Getenv("NOTION_SECRET"))
	q, err := queryDatabase(client, config.Notion)
	if err != nil {
		return fmt.Errorf("❌ Querying Notion database: %s", err)
	}
	fmt.Println("✔ Querying Notion database: Completed")

	// fetch page children
	changed := 0 // number of article status changed
	for i, page := range q.Results {
		pageName := tomarkdown.ConvertRichText(page.Properties.(notion.DatabasePageProperties)["Name"].Title)
		fmt.Printf("-- Article [%d/%d] %s --\n", i+1, len(q.Results), pageName)

		// Get page blocks tree
		blocks, err := queryBlockChildren(client, page.ID)
		if err != nil {
			log.Println("❌ Getting blocks tree:", err)
			continue
		}
		fmt.Println("✔ Getting blocks tree: Completed")

		// Generate content to file
		if err := generate(page, blocks, config.Markdown); err != nil {
			fmt.Println("❌ Generating blog post:", err)
			continue
		}
		fmt.Println("✔ Generating blog post: Completed")

		// Change status of blog post if desired
		if changeStatus(client, page, config.Notion) {
			changed++
		}
	}

	// Set GITHUB_ACTIONS info variables
	// https://docs.github.com/en/actions/learn-github-actions/workflow-commands-for-github-actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Printf("::set-output name=articles_published::%d\n", changed)
	}

	return nil
}

func generate(page notion.Page, blocks []notion.Block, config Markdown) error {
	// Create file
	pageName := config.PageNamePrefix + tomarkdown.ConvertRichText(page.Properties.(notion.DatabasePageProperties)["Name"].Title)
	f, err := os.Create(filepath.Join(config.PostSavePath, generateArticleFilename(pageName, page.CreatedTime, config)))
	if err != nil {
		return fmt.Errorf("error create file: %s", err)
	}

	// Generate markdown content to the file
	tm := tomarkdown.New()
	tm.ImgSavePath = filepath.Join(config.ImageSavePath, pageName)
	tm.ImgVisitPath = filepath.Join(config.ImagePublicLink, url.PathEscape(pageName))
	tm.ContentTemplate = config.Template
	tm.WithFrontMatter(page)
	if config.ShortcodeSyntax != "" {
		tm.EnableExtendedSyntax(config.ShortcodeSyntax)
	}

	return tm.GenerateTo(blocks, f)
}

func generateArticleFilename(title string, date time.Time, config Markdown) string {
	escapedTitle := strings.ReplaceAll(
		strings.ToValidUTF8(
			strings.ToLower(title),
			"",
		),
		" ", "-",
	)
	escapedFilename := escapedTitle + ".md"

	if config.GroupByMonth {
		return filepath.Join(date.Format("2006-01-02"), escapedFilename)
	}

	return escapedFilename
}
