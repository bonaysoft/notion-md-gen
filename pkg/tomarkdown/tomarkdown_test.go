package tomarkdown

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"testing"

	"github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testdatas embed.FS

func TestName(t *testing.T) {
	fs.WalkDir(testdatas, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		blockBytes, err := ioutil.ReadFile(path)
		assert.NoError(t, err)

		fmt.Printf("===== Testing %s =====\n", path)
		blocks := make([]notion.Block, 0)
		assert.NoError(t, json.Unmarshal(blockBytes, &blocks))
		assert.NoError(t, Gen(nil, blocks))
		return nil
	})
}
