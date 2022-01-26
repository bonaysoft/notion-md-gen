package notion_blog

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/janeczku/go-spinner"
)

func getImage(imgURL string, config BlogConfig) (_ string, err error) {
	// Split image url to get host and file name
	u, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("malformed url: %s", err)
	}

	// Get file name
	splitPaths := strings.Split(u.Path, "/")
	imageFilename := splitPaths[len(splitPaths)-1]
	if strings.HasPrefix(imageFilename, "Untitled.") {
		imageFilename = splitPaths[len(splitPaths)-2] + filepath.Ext(u.Path)
	}

	name := fmt.Sprintf("%s_%s", u.Hostname(), imageFilename)
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
