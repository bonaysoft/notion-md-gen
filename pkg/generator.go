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

func GenerateHeader(w io.Writer, p notionapi.Page) {
	title := p.Properties["Name"].(*notionapi.TitleProperty).Title
	createdBy := p.Properties["Created By"].(*notionapi.CreatedByProperty).CreatedBy.Name
	categories := p.Properties["Tags"].(*notionapi.MultiSelectProperty).MultiSelect
	categoriesStr := make([]string, len(categories))
	for i, cat := range categories {
		categoriesStr[i] = cat.Name
	}

	fmt.Fprintln(w, "+++")
	fmt.Fprintf(w, "title = %q\n", ConvertRichText(title))
	fmt.Fprintf(w, "date = %s\n", p.CreatedTime.Format("2006-01-02"))
	fmt.Fprintf(w, "lastmod = %s\n", p.LastEditedTime.Format("2006-01-02T15:04:05+07:00"))
	fmt.Fprintf(w, "categories = %q\n", categoriesStr)
	fmt.Fprintln(w, "draft = false")
	fmt.Fprintf(w, "author = %q\n", createdBy)
	fmt.Fprintln(w, "type = \"post\"")
	// fmt.Fprintf(w, "description = %v", categoriesStr)
	fmt.Fprintln(w, "+++")
	fmt.Fprintln(w)
	// +++
	// title = "Local item data"
	// date = 2019-09-11
	// lastmod = 2020-01-31T23:40:58+00:00
	// categories = ["version", "0.7.6"]
	// draft = false
	// description = "Item info stored locally. No need to update the whole game to modify items."
	// author = "Zebra"
	// type = "post"
	// +++
}

func Generate(w io.Writer, blocks []notionapi.Block, config BlogConfig) {
	if len(blocks) == 0 {
		return
	}

	for _, block := range blocks {
		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			log.Println("paragraph")
			fmt.Fprintln(w, ConvertRichText(b.Paragraph.Text))
			fmt.Fprintln(w)
			Generate(w, b.Paragraph.Children, config)
		case *notionapi.Heading1Block:
			log.Println("heading")
			fmt.Fprintf(w, "# %s\n", ConvertRichText(b.Heading1.Text))
		case *notionapi.Heading2Block:
			log.Println("heading")
			fmt.Fprintf(w, "## %s\n", ConvertRichText(b.Heading2.Text))
		case *notionapi.Heading3Block:
			log.Println("heading")
			fmt.Fprintf(w, "### %s\n", ConvertRichText(b.Heading3.Text))
		case *notionapi.BulletedListItemBlock:
			fmt.Fprintf(w, "- %s\n", ConvertRichText(b.BulletedListItem.Text))
			Generate(w, b.BulletedListItem.Children, config)
		case *notionapi.NumberedListItemBlock:
			fmt.Fprintf(w, "1. %s\n", ConvertRichText(b.NumberedListItem.Text))
			Generate(w, b.NumberedListItem.Children, config)
		case *notionapi.ImageBlock:
			log.Println(b.Image.File.URL)
			src, err := getImage(b.Image.File.URL, config)
			if err != nil {
				log.Println("error getting image:", err)
			}
			fmt.Fprintf(w, "![%s](%s)\n\n", ConvertRichText(b.Image.Caption), src)
		default:
			log.Println("unknown", block.GetType())
		}
	}
}
