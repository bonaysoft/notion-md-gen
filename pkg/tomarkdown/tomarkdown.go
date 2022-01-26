package tomarkdown

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/dstotijn/go-notion"
)

//go:embed templates
var mdTemplates embed.FS

type MdBlock struct {
	notion.Block
	BlockIndex int
}

type ToMarkdown struct {
	writer io.Writer
}

func New(writer io.Writer) *ToMarkdown {
	return &ToMarkdown{writer: writer}
}

func (tm *ToMarkdown) Gen(page *notion.Page, blocks []notion.Block) error {
	var sameBlockIdx int
	var lastBlockType notion.BlockType
	for _, block := range blocks {
		switch block.Type {
		case notion.BlockTypeImage:
			// todo download the image
		case notion.BlockTypeBookmark:
			// todo implement me
		case notion.BlockTypeCallout:
			// todo implement me

		}
		sameBlockIdx++
		if block.Type != lastBlockType {
			sameBlockIdx = 0
		}

		mdb := MdBlock{BlockIndex: sameBlockIdx, Block: block}
		if err := tm.RenderBlock(block.Type, mdb); err != nil {
			return err
		}
		lastBlockType = block.Type
	}

	return nil
}

func (tm *ToMarkdown) RenderBlock(kind notion.BlockType, block MdBlock) error {
	funcs := template.FuncMap{
		"rich2md": ConvertRichText,
		"add":     func(a, b int) int { return a + b },
		"deref":   func(i *bool) bool { return *i },
	}

	tpl, err := template.New(fmt.Sprintf("%s.gohtml", kind)).Funcs(funcs).ParseFS(mdTemplates, fmt.Sprintf("templates/%s.*", kind))
	if err != nil {
		return err
	}

	return tpl.Execute(tm.writer, block)
}

func ConvertRichText(t []notion.RichText) string {
	buf := &bytes.Buffer{}
	for _, word := range t {
		buf.WriteString(ConvertRich(word))
	}

	return buf.String()
}

func ConvertRich(t notion.RichText) string {
	switch t.Type {
	case notion.RichTextTypeText:
		if t.Text.Link != nil {
			return fmt.Sprintf(
				emphFormat(t.Annotations),
				fmt.Sprintf("[%s](%s)", t.Text.Content, t.Text.Link.URL),
			)
		}
		return fmt.Sprintf(emphFormat(t.Annotations), t.Text.Content)
	case notion.RichTextTypeEquation:
	case notion.RichTextTypeMention:
	}
	return ""
}

func emphFormat(a *notion.Annotations) (s string) {
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
