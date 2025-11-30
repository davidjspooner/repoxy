package reactui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/davidjspooner/go-http-server/pkg/handler"
)

//go:embed dist
var dist embed.FS

// Handler returns an http.Handler that serves the embedded React UI at /ui/.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("load embedded react ui: %w", err)
	}
	return handler.NewSinglePageApp(http.FS(sub), "/ui/", "index.html"), nil
}
