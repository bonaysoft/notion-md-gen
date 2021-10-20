package notion_blog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"net/url"

	"github.com/janeczku/go-spinner"
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
		if t.Text.Link != nil {
			return fmt.Sprintf(
				emphFormat(t.Annotations),
				fmt.Sprintf("[%s](%s)", t.Text.Content, t.Text.Link.Url),
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

func getImage(imgURL string, config BlogConfig) (_ string, err error) {
	// Split image url to get host and file name
	splittedURL, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("malformed url: %s", err)
	}

	// Get file name
	filePath := splittedURL.Path
	filePath = filePath[strings.LastIndex(filePath, "/")+1:]

	name := fmt.Sprintf("%s_%s", splittedURL.Hostname(), filePath)

	spin := spinner.StartNew(fmt.Sprintf("Getting image `%s`", name))
	defer func() {
		spin.Stop()
		if err != nil {
			fmt.Println(fmt.Sprintf("❌ Getting image `%s`: %s", name, err))
		} else {
			fmt.Println(fmt.Sprintf("✔ Getting image `%s`: Completed", name))
		}
	}()

	resp, err := http.Get(imgURL)
	if err != nil {
		return "", fmt.Errorf("couldn't download image: %s", err)
	}
	defer resp.Body.Close()

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

func Generate(w io.Writer, page notionapi.Page, blocks []notionapi.Block, config BlogConfig) error {
	// Parse template file
	t := template.New(path.Base(config.ArchetypeFile)).Delims("[[", "]]")
	t.Funcs(template.FuncMap{
		"rich": ConvertRichText,
	})

	t, err := t.ParseFiles(config.ArchetypeFile)
	if err != nil {
		return fmt.Errorf("error parsing archetype file: %s", err)
	}

	// Generate markdown content
	buffer := &bytes.Buffer{}
	GenerateContent(buffer, blocks, config)

	// Dump markdown content into output according to archetype file
	fileArchetype := MakeArchetypeFields(page, config)
	fileArchetype.Content = buffer.String()
	err = t.Execute(w, fileArchetype)
	if err != nil {
		return fmt.Errorf("error filling archetype file: %s", err)
	}

	return nil
}

func GenerateContent(w io.Writer, blocks []notionapi.Block, config BlogConfig, prefixes ...string) {
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
			fprintln(w, prefixes, ConvertRichText(b.Paragraph.Text)+"\n")
			GenerateContent(w, b.Paragraph.Children, config)
		case *notionapi.Heading1Block:
			fprintf(w, prefixes, "# %s", ConvertRichText(b.Heading1.Text))
		case *notionapi.Heading2Block:
			fprintf(w, prefixes, "## %s", ConvertRichText(b.Heading2.Text))
		case *notionapi.Heading3Block:
			fprintf(w, prefixes, "### %s", ConvertRichText(b.Heading3.Text))
		case *notionapi.CalloutBlock:
			if !config.UseShortcodes {
				continue
			}
			if b.Callout.Icon != nil {
				if b.Callout.Icon.Emoji != nil {
					fprintf(w, prefixes, `{{%% callout emoji="%s" %%}}`, *b.Callout.Icon.Emoji)
				} else {
					fprintf(w, prefixes, `{{%% callout image="%s" %%}}`, b.Callout.Icon.GetURL())
				}
			}
			fprintln(w, prefixes, ConvertRichText(b.Callout.Text))
			GenerateContent(w, b.Callout.Children, config, prefixes...)
			fprintln(w, prefixes, "{{% /callout %}}")

		case *notionapi.BookmarkBlock:
			if !config.UseShortcodes {
				// Simply generate the url link
				fprintf(w, prefixes, "[%s](%s)", b.Bookmark.URL, b.Bookmark.URL)
				continue
			}
			// Parse external page metadata
			og, err := parseMetadata(b.Bookmark.URL, config)
			if err != nil {
				log.Println("error getting bookmark metadata:", err)
			}

			// GenerateContent shortcode with given metadata
			fprintf(w, prefixes,
				`{{< bookmark url="%s" title="%s" img="%s" >}}`,
				og.URL,
				og.Title,
				og.Image,
			)
			fprintln(w, prefixes, og.Description)
			fprintln(w, prefixes, "{{< /bookmark >}}")

		case *notionapi.QuoteBlock:
			fprintf(w, prefixes, "> %s", ConvertRichText(b.Quote.Text))
			GenerateContent(w, b.Quote.Children, config,
				append([]string{"> "}, prefixes...)...)

		case *notionapi.BulletedListItemBlock:
			bulletedList = true
			fprintf(w, prefixes, "- %s", ConvertRichText(b.BulletedListItem.Text))
			GenerateContent(w, b.BulletedListItem.Children, config,
				append([]string{"    "}, prefixes...)...)

		case *notionapi.NumberedListItemBlock:
			numberedList = true
			fprintf(w, prefixes, "1. %s", ConvertRichText(b.NumberedListItem.Text))
			GenerateContent(w, b.NumberedListItem.Children, config,
				append([]string{"    "}, prefixes...)...)

		case *notionapi.ImageBlock:
			src, _ := getImage(b.Image.File.URL, config)
			fprintf(w, prefixes, "![%s](%s)\n", ConvertRichText(b.Image.Caption), src)

		case *notionapi.CodeBlock:
			if b.Code.Language == "plain text" {
				fprintln(w, prefixes, "```")
			} else {
				fprintf(w, prefixes, "```%s", b.Code.Language)
			}
			fprintln(w, prefixes, ConvertRichText(b.Code.Text))
			fprintln(w, prefixes, "```")

		case *notionapi.UnsupportedBlock:
			if b.GetType() != "unsupported" {
				fmt.Println("ℹ Unimplemented block", b.GetType())
			} else {
				fmt.Println("ℹ Unsupported block type")
			}
		default:
			fmt.Println("ℹ Unimplemented block", b.GetType())
		}
	}
}
