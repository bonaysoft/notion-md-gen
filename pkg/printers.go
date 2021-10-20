package notion_blog

import (
	"fmt"
	"io"
	"log"
	"strings"
)

func fprintf(w io.Writer, prefixes []string, format string, vals ...interface{}) {
	// Prepend prefixes to the format
	format = strings.Join(prefixes, "") + format + "\n"
	// Print to writer
	_, err := fmt.Fprintf(w, format, vals...)
	if err != nil {
		log.Printf("couldn't write `%s` to file: %s\n", format, err)
	}
}
func fprintln(w io.Writer, prefixes []string, vals ...interface{}) {
	prefix := strings.Join(prefixes, "")
	args := make([]interface{}, len(vals)+1)
	args[0] = prefix
	copy(args[1:], vals)
	// Print to writer
	fprintf(w, prefixes, "%s", fmt.Sprint(vals...))
}
