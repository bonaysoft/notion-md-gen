package notion_blog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"notion-md-gen/pkg/tomarkdown"

	"github.com/dstotijn/go-notion"
)

type ArchetypeFields struct {
	Title        string
	Description  string
	Banner       string
	CreationDate time.Time
	LastModified time.Time
	Author       string
	Tags         []notion.SelectOptions
	Categories   []notion.SelectOptions
	Content      string
}

func MakeArchetypeFields(p notion.Page, config BlogConfig) ArchetypeFields {
	// Initialize first default Notion page fields
	pageProps := p.Properties.(notion.DatabasePageProperties)
	a := ArchetypeFields{
		Title:        tomarkdown.ConvertRichText(pageProps["Name"].Title),
		Author:       pageProps["Created By"].CreatedBy.Name,
		CreationDate: p.CreatedTime,
		LastModified: p.LastEditedTime,
	}

	a.Banner = ""
	// if p.Cover != nil && p.Cover.GetURL() != "" {
	// 	coverSrc, _ := getImage(p.Cover.GetURL(), config)
	// 	a.Banner = coverSrc
	// }

	// Custom fields
	propExtractBind(pageProps, config.PropertyTitle, func(target interface{}) { a.Title = target.(string) })
	propExtractBind(pageProps, config.PropertyDescription, func(target interface{}) { a.Description = target.(string) })
	propExtractBind(pageProps, config.PropertyCategories, func(target interface{}) { a.Categories = target.([]notion.SelectOptions) })
	propExtractBind(pageProps, config.PropertyTags, func(target interface{}) { a.Tags = target.([]notion.SelectOptions) })
	return a
}

// propExtractBind extract some prop
// todo refactor when the go v1.18 release
func propExtractBind(props notion.DatabasePageProperties, key string, bindTo func(target interface{})) {
	v, ok := props[key]
	if !ok {
		log.Printf("warning: given property %s is not exist\n", key)
		return
	}

	switch v.Type {
	case notion.DBPropTypeRichText:
		bindTo(tomarkdown.ConvertRichText(v.RichText))
	case notion.DBPropTypeMultiSelect:
		bindTo(v.MultiSelect)
	default:
		log.Printf("warning: given property %s is not supported type: %T\n", key, v)
	}
}

func Generate(w io.Writer, page notion.Page, blocks []notion.Block, config BlogConfig) error {
	// Parse template file
	t := template.New(path.Base(config.ArchetypeFile)).Delims("[[", "]]")
	t.Funcs(template.FuncMap{
		"add":    func(a, b int) int { return a + b },
		"sub":    func(a, b int) int { return a - b },
		"mul":    func(a, b int) int { return a * b },
		"div":    func(a, b int) int { return a / b },
		"repeat": func(s string, n int) string { return strings.Repeat(s, n) },
		"rich":   tomarkdown.ConvertRichText,
	})

	t, err := t.ParseFiles(config.ArchetypeFile)
	if err != nil {
		return fmt.Errorf("error parsing archetype file: %s", err)
	}

	// Dump markdown content into output according to archetype file
	fileArchetype := MakeArchetypeFields(page, config)
	config.ImagesFolder = filepath.Join(config.ImagesFolder, fileArchetype.Title)
	config.ImagesLink = filepath.Join(config.ImagesLink, url.PathEscape(fileArchetype.Title))

	// Generate markdown content
	buffer := &bytes.Buffer{}
	// GenerateContent(buffer, blocks, config)
	if err := tomarkdown.New(buffer).Gen(nil, blocks); err != nil {
		return err
	}

	fileArchetype.Content = buffer.String()
	err = t.Execute(w, fileArchetype)
	if err != nil {
		return fmt.Errorf("error filling archetype file: %s", err)
	}

	return nil
}