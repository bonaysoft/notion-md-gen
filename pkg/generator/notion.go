package generator

import (
	"context"
	"log"

	"notion-md-gen/pkg/config"

	"github.com/dstotijn/go-notion"
	"github.com/janeczku/go-spinner"
)

func filterFromConfig(config config.BlogConfig) *notion.DatabaseQueryFilter {
	if config.FilterProp == "" || len(config.FilterValue) == 0 {
		return nil
	}

	properties := make([]notion.DatabaseQueryFilter, len(config.FilterValue))
	for i, val := range config.FilterValue {
		properties[i] = notion.DatabaseQueryFilter{
			Property: config.FilterProp,
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: val,
			},
		}
	}

	return &notion.DatabaseQueryFilter{
		Or: properties,
	}
}

func queryDatabase(client *notion.Client, config config.BlogConfig) (notion.DatabaseQueryResponse, error) {
	spin := spinner.StartNew("Querying Notion database")
	defer spin.Stop()

	query := &notion.DatabaseQuery{
		Filter:   filterFromConfig(config),
		PageSize: 100,
	}
	return client.QueryDatabase(context.Background(), config.DatabaseID, query)
}

func queryBlockChildren(client *notion.Client, blockID string) (blocks []notion.Block, err error) {
	spin := spinner.StartNew("Getting blocks tree")
	defer spin.Stop()
	return retrieveBlockChildren(client, blockID)
}

func retrieveBlockChildren(client *notion.Client, blockID string) (blocks []notion.Block, err error) {
	query := &notion.PaginationQuery{
		PageSize: 100,
	}
	res, err := client.FindBlockChildrenByID(context.Background(), blockID, query)
	if err != nil {
		return nil, err
	}

	blocks = res.Results
	if len(blocks) == 0 {
		return
	}

	for _, block := range res.Results {
		if !block.HasChildren {
			continue
		}

		switch block.Type {
		case notion.BlockTypeParagraph:
			block.Paragraph.Children, err = retrieveBlockChildren(client, block.ID)
		case notion.BlockTypeCallout:
			block.Callout.Children, err = retrieveBlockChildren(client, block.ID)
		case notion.BlockTypeQuote:
			block.Quote.Children, err = retrieveBlockChildren(client, block.ID)
		case notion.BlockTypeBulletedListItem:
			block.BulletedListItem.Children, err = retrieveBlockChildren(client, block.ID)
		case notion.BlockTypeNumberedListItem:
			block.NumberedListItem.Children, err = retrieveBlockChildren(client, block.ID)
		case notion.BlockTypeTable:
			block.Table.Children, err = retrieveBlockChildren(client, block.ID)
		}

		if err != nil {
			return
		}
	}

	return blocks, nil
}

// changeStatus changes the Notion article status to the published value if set.
// It returns true if status changed.
func changeStatus(client *notion.Client, p notion.Page, config config.BlogConfig) bool {
	// No published value or filter prop to change
	if config.FilterProp == "" || config.PublishedValue == "" {
		return false
	}

	if v, ok := p.Properties.(notion.DatabasePageProperties)[config.FilterProp]; ok {
		if v.Select.Name == config.PublishedValue {
			return false
		}
	} else { // No filter prop in page, can't change it
		return false
	}

	updatedProps := make(notion.DatabasePageProperties)
	updatedProps[config.FilterProp] = notion.DatabasePageProperty{
		Select: &notion.SelectOptions{
			Name: config.PublishedValue,
		},
	}

	_, err := client.UpdatePage(context.Background(), p.ID,
		notion.UpdatePageParams{
			DatabasePageProperties: &updatedProps,
		},
	)
	if err != nil {
		log.Println("error changing status:", err)
	}

	return err == nil
}
