package notion_blog

type BlogConfig struct {
	DatabaseID string `usage:"ID of the Notion database of the blog."`

	ImagesLink    string `usage:"Directory in which the hugo content will be generated."`
	ImagesFolder  string `usage:"Directory in which the static images will be stored. E.g.: ./web/static/images"`
	ContentFolder string `usage:"URL beggining to link the static images. E.g.: /images"`
	ArchetypeFile string `usage:"Route to the archetype file to generate the header."`

	FilterProp  string `usage:"Property of the filter to apply to a select value of the articles."`
	FilterValue string `usage:"Value of the filter to apply to the Notion articles database."`
}
