package notion_blog

type BlogConfig struct {
	DatabaseID string `usage:"ID of the Notion database of the blog."`

	ImagesLink    string `usage:"Directory in which the hugo content will be generated."`
	ImagesFolder  string `usage:"Directory in which the static images will be stored. E.g.: ./web/static/images"`
	ContentFolder string `usage:"URL beggining to link the static images. E.g.: /images"`
	ArchetypeFile string `usage:"Route to the archetype file to generate the header."`

	// Optional:

	PropertyDescription string `usage:"Description property name in Notion."`
	PropertyTags        string `usage:"Tags multi-select porperty name in Notion."`
	PropertyCategories  string `usage:"Categories multi-select porperty name in Notion."`

	FilterProp     string   `usage:"Property of the filter to apply to a select value of the articles."`
	FilterValue    []string `usage:"Value of the filter to apply to the Notion articles database."`
	PublishedValue string   `usage:"Value to which the filter property will be set after generating the content."`

	UseShortcodes bool `usage:"True if you want to generate shortcodes for unimplemented markdown blocks, such as callout or quote."`
}
