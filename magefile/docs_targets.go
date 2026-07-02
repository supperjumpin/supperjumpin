//go:build mage

package main

import (
	"fmt"
	"net/http"
	"os"
)

// Docs serves Swagger UI for apps/api/openapi.yaml.
func Docs() error {
	port := valueOrDefault(os.Getenv("DOCS_PORT"), DefaultDocsPort)
	openAPIPath := repoPath("apps", "api", "openapi.yaml")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <title>Supperjumpin API Docs</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({ url: '/openapi.yaml', dom_id: '#swagger-ui' });
    </script>
  </body>
</html>`)))
	})
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, openAPIPath)
	})

	addr := ":" + port
	fmt.Printf("Serving docs on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}
