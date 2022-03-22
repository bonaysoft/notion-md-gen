package tomarkdown

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testdatas embed.FS

func testTarget(t *testing.T, target string) {
	fs.WalkDir(testdatas, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		blockBytes, err := ioutil.ReadFile(path)
		assert.NoError(t, err)

		fmt.Printf("===== Testing %s =====\n", path)
		blocks := make([]notion.Block, 0)
		assert.NoError(t, json.Unmarshal(blockBytes, &blocks))
		tom := New()
		tom.ImgSavePath = "/tmp/"
		tom.EnableExtendedSyntax(target)
		assert.NoError(t, tom.GenerateTo(blocks, os.Stdout))
		return nil
	})
}
func TestOne(t *testing.T) {
	testTarget(t, "vuepress")
}

func TestAllTarget(t *testing.T) {
	targets := []string{"hugo", "hexo", "vuepress"}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			testTarget(t, target)
		})
	}
}
