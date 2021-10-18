package notion_blog

import (
	"fmt"

	"github.com/otiai10/opengraph"
)

type OGMetadata struct {
	Title       string
	Description string
	URL         string
	Image       string
	Favicon     string
}

// parseMetadata returns the OpenGraph metadata of a page so we can generate a
// bookmark.
func parseMetadata(url string) (*OGMetadata, error) {
	og, err := opengraph.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse metadata of `%s`: %s", url, err)
	}
	if og == nil {
		return nil, fmt.Errorf("unexpected error")
	}

	// Change to absolute urls
	og.ToAbsURL()

	imgSrc := ""
	for _, img := range og.Image {
		if img != nil && img.URL != "" {
			imgSrc = img.URL
			break
		}
	}

	return &OGMetadata{
		Title:       og.Title,
		Description: og.Description,
		URL:         url,
		Image:       imgSrc,
		Favicon:     og.Favicon,
	}, nil
}
