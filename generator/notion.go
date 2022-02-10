package generator

import (
	"context"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dstotijn/go-notion"
)

var spin = spinner.New(spinner.CharSets[14], time.Millisecond*100)

func filterFromConfig(config Notion) *notion.DatabaseQueryFilter {
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

func queryDatabase(client *notion.Client, config Notion) (notion.DatabaseQueryResponse, error) {
	spin.Suffix = " Querying Notion database..."
	spin.Start()
	defer spin.Stop()

	query := &notion.DatabaseQuery{
		Filter:   filterFromConfig(config),
		PageSize: 100,
	}
	return client.QueryDatabase(context.Background(), config.DatabaseID, query)
}

func queryBlockChildren(client *notion.Client, blockID string) (blocks []notion.Block, err error) {
	spin.Suffix = " Fetching blocks tree..."
	spin.Start()
	defer spin.Stop()
	return retrieveBlockChildren(client, blockID)
}

func retrieveBlockChildrenLoop(client *notion.Client, blockID, cursor string) (blocks []notion.Block, err error) {
	for {
		query := &notion.PaginationQuery{
			StartCursor: cursor,
			PageSize:    100,
		}
		res, err := client.FindBlockChildrenByID(context.Background(), blockID, query)
		if err != nil {
			return nil, err
		}

		if len(res.Results) == 0 {
			return blocks, nil
		}

		blocks = append(blocks, res.Results...)
		if !res.HasMore {
			return blocks, nil
		}
		cursor = *res.NextCursor
	}
}

func retrieveBlockChildren(client *notion.Client, blockID string) (blocks []notion.Block, err error) {
	blocks, err = retrieveBlockChildrenLoop(client, blockID, "")
	if err != nil {
		return
	}

	for _, block := range blocks {
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
func changeStatus(client *notion.Client, p notion.Page, config Notion) bool {
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
