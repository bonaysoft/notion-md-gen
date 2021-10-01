package notion_blog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jomei/notionapi"
)

func emphFormat(a *notionapi.Annotations) (s string) {
	s = "%s"
	if a == nil {
		return
	}

	if a.Code {
		return "`%s`"
	}

	switch {
	case a.Bold && a.Italic:
		s = "***%s***"
	case a.Bold:
		s = "**%s**"
	case a.Italic:
		s = "*%s*"
	}

	if a.Underline {
		s = "__" + s + "__"
	} else if a.Strikethrough {
		s = "~~" + s + "~~"
	}

	// TODO: color

	return s
}

func ConvertRich(t notionapi.RichText) string {
	switch t.Type {
	case notionapi.ObjectTypeText:
		if t.Text.Link != "" {
			return fmt.Sprintf(
				emphFormat(t.Annotations),
				fmt.Sprintf("[%s](%s)", t.Text.Content, t.Text.Link),
			)
		}
		return fmt.Sprintf(emphFormat(t.Annotations), t.Text.Content)
	case notionapi.ObjectTypeList:
	}
	return ""
}

func ConvertRichText(t []notionapi.RichText) string {
	buf := &bytes.Buffer{}
	for _, word := range t {
		buf.WriteString(ConvertRich(word))
	}

	return buf.String()
}

func getImage(url string, config BlogConfig) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("couldn't download image: %s", err)
	}
	defer resp.Body.Close()

	name := url[strings.LastIndex(url, "/")+1 : strings.Index(url, "?")]

	err = os.MkdirAll(config.ImagesFolder, 0777)
	if err != nil {
		return "", fmt.Errorf("couldn't create images folder: %s", err)
	}

	// Create the file
	out, err := os.Create(filepath.Join(config.ImagesFolder, name))
	if err != nil {
		return name, fmt.Errorf("couldn't create image file: %s", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return filepath.Join(config.ImagesLink, name), err
}

func GenerateHeader(w io.Writer, p notionapi.Page, config BlogConfig) {
	t, err := template.ParseFiles(config.ArchetypeFile)
	if err != nil {
		log.Fatalf("error parsing archetype file: %s", err)
	}

	err = t.Execute(w, ArchetypeFields{
		Title: ConvertRichText(p.Properties["Name"].(*notionapi.TitleProperty).Title),
		// Description:  ConvertRichText(p.Properties["Description"].(*notionapi.TextProperty).Text),
		Description:  "",
		Banner:       "",
		CreationDate: p.CreatedTime,
		LastModified: p.LastEditedTime,
		Author:       p.Properties["Created By"].(*notionapi.CreatedByProperty).CreatedBy.Name,
		Tags:         p.Properties["Tags"].(*notionapi.MultiSelectProperty).MultiSelect,
		// Categories:   p.Properties["Categories"].(*notionapi.MultiSelectProperty).MultiSelect,
		Categories: []notionapi.Option{},
	})
	if err != nil {
		log.Fatalf("error filling archetype file: %s", err)
	}
}

func Generate(w io.Writer, blocks []notionapi.Block, config BlogConfig) {
	if len(blocks) == 0 {
		return
	}

	numberedList := false
	bulletedList := false

	for _, block := range blocks {
		// Add line break after list is finished
		if bulletedList && block.GetType() != notionapi.BlockTypeBulletedListItem {
			bulletedList = false
			fmt.Fprintln(w)
		}
		if numberedList && block.GetType() != notionapi.BlockTypeNumberedListItem {
			numberedList = false
			fmt.Fprintln(w)
		}

		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			fmt.Fprintln(w, ConvertRichText(b.Paragraph.Text))
			fmt.Fprintln(w)
			Generate(w, b.Paragraph.Children, config)
		case *notionapi.Heading1Block:
			fmt.Fprintf(w, "# %s\n", ConvertRichText(b.Heading1.Text))
		case *notionapi.Heading2Block:
			fmt.Fprintf(w, "## %s\n", ConvertRichText(b.Heading2.Text))
		case *notionapi.Heading3Block:
			fmt.Fprintf(w, "### %s\n", ConvertRichText(b.Heading3.Text))
		case *notionapi.BulletedListItemBlock:
			bulletedList = true
			fmt.Fprintf(w, "- %s\n", ConvertRichText(b.BulletedListItem.Text))
			Generate(w, b.BulletedListItem.Children, config)
		case *notionapi.NumberedListItemBlock:
			numberedList = true
			fmt.Fprintf(w, "1. %s\n", ConvertRichText(b.NumberedListItem.Text))
			Generate(w, b.NumberedListItem.Children, config)
		case *notionapi.ImageBlock:
			src, err := getImage(b.Image.File.URL, config)
			if err != nil {
				log.Println("error getting image:", err)
			}
			fmt.Fprintf(w, "![%s](%s)\n\n", ConvertRichText(b.Image.Caption), src)
		case *notionapi.CodeBlock:
			fmt.Fprintf(w, "```%s\n", b.Code.Language)
			fmt.Fprintln(w, ConvertRichText(b.Code.Text))
			fmt.Fprintln(w, "```")
		case *notionapi.UnsupportedBlock:
			if b.GetType() != "unsupported" {
				log.Println("unimplemented", block.GetType())
			} else {
				log.Println("unsupported block type")
			}
		default:
			log.Println("unimplemented", block.GetType())
		}
	}
}
