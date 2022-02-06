package tomarkdown

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/dstotijn/go-notion"
	"github.com/otiai10/opengraph"
	"gopkg.in/yaml.v3"
)

//go:embed templates
var mdTemplatesFS embed.FS

var (
	extendedSyntaxBlocks            = []notion.BlockType{notion.BlockTypeBookmark, notion.BlockTypeCallout}
	blockTypeInExtendedSyntaxBlocks = func(bType notion.BlockType) bool {
		for _, blockType := range extendedSyntaxBlocks {
			if blockType == bType {
				return true
			}
		}

		return false
	}
)

type MdBlock struct {
	notion.Block
	Depth int
	Extra map[string]interface{}
}

type ToMarkdown struct {
	FrontMatter     map[string]interface{}
	ContentBuffer   *bytes.Buffer
	ImgSavePath     string
	ImgVisitPath    string
	ContentTemplate string

	extra map[string]interface{}
}

func New() *ToMarkdown {
	return &ToMarkdown{
		FrontMatter:   make(map[string]interface{}),
		ContentBuffer: new(bytes.Buffer),
		extra:         make(map[string]interface{}),
	}
}

func (tm *ToMarkdown) WithFrontMatter(page notion.Page) {
	tm.injectFrontMatterCover(page.Cover)
	pageProps := page.Properties.(notion.DatabasePageProperties)
	for fmKey, property := range pageProps {
		tm.injectFrontMatter(fmKey, property)
	}
}

func (tm *ToMarkdown) EnableExtendedSyntax(target string) {
	tm.extra["ExtendedSyntaxEnabled"] = true
	tm.extra["ExtendedSyntaxTarget"] = target
}

func (tm *ToMarkdown) ExtendedSyntaxEnabled() bool {
	if v, ok := tm.extra["ExtendedSyntaxEnabled"].(bool); ok {
		return v
	}

	return false
}

func (tm *ToMarkdown) shouldSkipRender(bType notion.BlockType) bool {
	return !tm.ExtendedSyntaxEnabled() && blockTypeInExtendedSyntaxBlocks(bType)
}

func (tm *ToMarkdown) GenerateTo(page notion.Page, blocks []notion.Block, writer io.Writer) error {
	if err := tm.GenFrontMatter(writer); err != nil {
		return err
	}

	if err := tm.GenContentBlocks(blocks, 0); err != nil {
		return err
	}

	if tm.ContentTemplate != "" {
		t, err := template.ParseFiles(tm.ContentTemplate)
		if err != nil {
			return err
		}
		return t.Execute(writer, tm)
	}

	_, err := io.Copy(writer, tm.ContentBuffer)
	return err
}

func (tm *ToMarkdown) GenFrontMatter(writer io.Writer) error {
	if len(tm.FrontMatter) == 0 {
		return nil
	}

	nfm := make(map[string]interface{})
	for key, value := range tm.FrontMatter {
		nfm[strings.ToLower(key)] = value
	}

	frontMatters, err := yaml.Marshal(nfm)
	if err != nil {
		return nil
	}

	buffer := new(bytes.Buffer)
	buffer.WriteString("---\n")
	buffer.Write(frontMatters)
	buffer.WriteString("---\n\n")
	_, err = io.Copy(writer, buffer)
	return err
}

func (tm *ToMarkdown) GenContentBlocks(blocks []notion.Block, depth int) error {
	var sameBlockIdx int
	var lastBlockType notion.BlockType
	for _, block := range blocks {
		if tm.shouldSkipRender(block.Type) {
			continue
		}
		mdb := MdBlock{
			Block: block,
			Depth: depth,
			Extra: tm.extra,
		}

		sameBlockIdx++
		if block.Type != lastBlockType {
			sameBlockIdx = 0
		}
		mdb.Extra["SameBlockIdx"] = sameBlockIdx

		var err error
		switch block.Type {
		case notion.BlockTypeImage:
			err = tm.downloadImage(block.Image)
		case notion.BlockTypeBookmark:
			err = tm.injectBookmarkInfo(block.Bookmark, &mdb.Extra)
		}
		if err != nil {
			return err
		}

		if err := tm.GenBlock(block.Type, mdb); err != nil {
			return err
		}
		lastBlockType = block.Type
	}

	return nil
}

func (tm *ToMarkdown) GenBlock(bType notion.BlockType, block MdBlock) error {
	funcs := sprig.TxtFuncMap()
	funcs["deref"] = func(i *bool) bool { return *i }
	funcs["rich2md"] = ConvertRichText
	t := template.New(fmt.Sprintf("%s.gohtml", bType)).Funcs(funcs)
	tpl, err := t.ParseFS(mdTemplatesFS, fmt.Sprintf("templates/%s.*", bType))
	if err != nil {
		return err
	}

	if err := tpl.Execute(tm.ContentBuffer, block); err != nil {
		return err
	}

	if block.HasChildren {
		block.Depth++
		return tm.GenContentBlocks(getChildrenBlocks(block), block.Depth)
	}

	return nil
}

func (tm *ToMarkdown) downloadImage(image *notion.FileBlock) error {
	download := func(imgURL string) (string, error) {
		resp, err := http.Get(imgURL)
		if err != nil {
			return "", err
		}

		imgFilename, err := tm.saveTo(resp.Body, imgURL, tm.ImgSavePath)
		if err != nil {
			return "", err
		}

		return filepath.Join(tm.ImgVisitPath, imgFilename), nil
	}

	var err error
	if image.Type == notion.FileTypeExternal {
		image.External.URL, err = download(image.External.URL)
	}
	if image.Type == notion.FileTypeFile {
		image.File.URL, err = download(image.File.URL)
	}

	return err
}

func (tm *ToMarkdown) saveTo(reader io.Reader, rawURL, distDir string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("malformed url: %s", err)
	}

	// gen file name
	splitPaths := strings.Split(u.Path, "/")
	imageFilename := splitPaths[len(splitPaths)-1]
	if strings.HasPrefix(imageFilename, "Untitled.") {
		imageFilename = splitPaths[len(splitPaths)-2] + filepath.Ext(u.Path)
	}

	if err := os.MkdirAll(distDir, 0755); err != nil {
		return "", fmt.Errorf("%s: %s", distDir, err)
	}

	filename := fmt.Sprintf("%s_%s", u.Hostname(), imageFilename)
	out, err := os.Create(filepath.Join(distDir, filename))
	if err != nil {
		return "", fmt.Errorf("couldn't create image file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return filename, err
}

// injectBookmarkInfo set bookmark info into the extra map field
func (tm *ToMarkdown) injectBookmarkInfo(bookmark *notion.Bookmark, extra *map[string]interface{}) error {
	og, err := opengraph.Fetch(bookmark.URL)
	if err != nil {
		return err
	}
	og.ToAbsURL()
	for _, img := range og.Image {
		if img != nil && img.URL != "" {
			(*extra)["Image"] = img.URL
			break
		}
	}
	(*extra)["Title"] = og.Title
	(*extra)["Description"] = og.Description
	return nil
}

// injectFrontMatter convert the prop to the front-matter
func (tm *ToMarkdown) injectFrontMatter(key string, property notion.DatabasePageProperty) {
	var fmv interface{}
	switch prop := property.Value().(type) {
	case *notion.SelectOptions:
		fmv = prop.Name
	case []notion.SelectOptions:
		opts := make([]string, 0)
		for _, options := range prop {
			opts = append(opts, options.Name)
		}
		fmv = opts
	case []notion.RichText:
		fmv = ConvertRichText(prop)
	case *time.Time:
		fmv = prop.Format("2006-01-02T15:04:05+07:00")
	case *notion.Date:
		fmv = prop.Start.Format("2006-01-02T15:04:05+07:00")
	case *notion.User:
		fmv = prop.Name
	case *string:
		fmv = *prop
	case *float64:
		fmv = *prop
	default:
		fmt.Printf("Unsupport prop: %s - %T\n", prop, prop)
	}

	if fmv == nil {
		return
	}

	// todo support settings mapping relation
	tm.FrontMatter[key] = fmv
}

func (tm *ToMarkdown) injectFrontMatterCover(cover *notion.Cover) {
	if cover == nil {
		return
	}

	image := &notion.FileBlock{
		Type:     cover.Type,
		File:     cover.File,
		External: cover.External,
	}
	if err := tm.downloadImage(image); err != nil {
		return
	}

	if image.Type == notion.FileTypeExternal {
		tm.FrontMatter["cover"] = image.External.URL
	}
	if image.Type == notion.FileTypeFile {
		tm.FrontMatter["cover"] = image.File.URL
	}
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

func getChildrenBlocks(block MdBlock) []notion.Block {
	switch block.Type {
	case notion.BlockTypeQuote:
		return block.Quote.Children
	case notion.BlockTypeToggle:
		return block.Toggle.Children
	case notion.BlockTypeParagraph:
		return block.Paragraph.Children
	case notion.BlockTypeCallout:
		return block.Callout.Children
	case notion.BlockTypeBulletedListItem:
		return block.BulletedListItem.Children
	case notion.BlockTypeNumberedListItem:
		return block.NumberedListItem.Children
	case notion.BlockTypeToDo:
		return block.ToDo.Children
	case notion.BlockTypeCode:
		return block.Code.Children
	case notion.BlockTypeColumn:
		return block.Column.Children
	case notion.BlockTypeColumnList:
		return block.ColumnList.Children
	case notion.BlockTypeTable:
		return block.Table.Children
	case notion.BlockTypeSyncedBlock:
		return block.SyncedBlock.Children
	case notion.BlockTypeTemplate:
		return block.Template.Children
	}

	return nil
}
