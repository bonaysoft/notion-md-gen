package notion_blog

import (
	"log"
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
	Content      string
	Properties   notionapi.Properties
}

func MakeArchetypeFields(p notionapi.Page, config BlogConfig) ArchetypeFields {
	// Initialize first default Notion page fields
	a := ArchetypeFields{
		Title:        ConvertRichText(p.Properties["Name"].(*notionapi.TitleProperty).Title),
		CreationDate: p.CreatedTime,
		LastModified: p.LastEditedTime,
		Author:       p.Properties["Created By"].(*notionapi.CreatedByProperty).CreatedBy.Name,
	}

	a.Banner = ""
	if p.Cover != nil && p.Cover.GetURL() != "" {
		coverSrc, _ := getImage(p.Cover.GetURL(), config)
		a.Banner = coverSrc
	}

	// Custom fields
	if v, ok := p.Properties[config.PropertyDescription]; ok {
		text, ok := v.(*notionapi.RichTextProperty)
		if ok {
			a.Description = ConvertRichText(text.RichText)
		} else {
			log.Println("warning: given property description is not a text property")
		}
	}

	if v, ok := p.Properties[config.PropertyCategories]; ok {
		multiSelect, ok := v.(*notionapi.MultiSelectProperty)
		if ok {
			a.Categories = multiSelect.MultiSelect
		} else {
			log.Println("warning: given property categories is not a multi-select property")
		}
	}

	if v, ok := p.Properties[config.PropertyTags]; ok {
		multiSelect, ok := v.(*notionapi.MultiSelectProperty)
		if ok {
			a.Tags = multiSelect.MultiSelect
		} else {
			log.Println("warning: given property tags is not a multi-select property")
		}
	}

	a.Properties = p.Properties

	return a
}
