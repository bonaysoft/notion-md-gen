package notion_blog

import (
	"time"

	"github.com/jomei/notionapi"
)

type ArchetypeFields struct {
	Title        string
	Description  string
	Banner       string
	CreationDate time.Time
	LastModified time.Time
	Author       string
	Tags         []notionapi.Option
	Categories   []notionapi.Option
}
