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
		Author:       p.Properties["Created By"].(*notionapi.CreatedByProperty).CreatedBy.Name,
		CreationDate: p.CreatedTime,
		LastModified: p.LastEditedTime,
	}

	a.Banner = ""
	if p.Cover != nil && p.Cover.GetURL() != "" {
		coverSrc, _ := getImage(p.Cover.GetURL(), config)
		a.Banner = coverSrc
	}

	// Custom fields
	propExtractBind(p.Properties, config.PropertyTitle, func(target interface{}) { a.Title = target.(string) })
	propExtractBind(p.Properties, config.PropertyDescription, func(target interface{}) { a.Description = target.(string) })
	propExtractBind(p.Properties, config.PropertyCategories, func(target interface{}) { a.Categories = target.([]notionapi.Option) })
	propExtractBind(p.Properties, config.PropertyTags, func(target interface{}) { a.Tags = target.([]notionapi.Option) })

	a.Properties = p.Properties

	return a
}

// propExtractBind extract some prop
// todo refactor when the go v1.18 release
func propExtractBind(props notionapi.Properties, key string, bindTo func(target interface{})) {
	v, ok := props[key]
	if !ok {
		return
	}

	switch vv := v.(type) {
	case *notionapi.RichTextProperty:
		bindTo(ConvertRichText(vv.RichText))
	case *notionapi.MultiSelectProperty:
		bindTo(vv.MultiSelect)
	default:
		log.Println("warning: given property %s is not supported type: %T", key, v)
	}
}
