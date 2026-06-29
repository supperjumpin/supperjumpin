package evidence

import (
	"os"
	"path/filepath"
	"strings"
)

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func filenameFromURL(u string) string {
	idx := strings.LastIndex(u, "/")
	if idx < 0 {
		return u
	}
	return u[idx+1:]
}

// filepathExt returns the extension of a URL path (e.g. ".png" for "photo.png").
// It's a tiny helper to keep the tests scannable.
var _ = filepath.Ext
