package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"gopkg.in/yaml.v3"
)

const openAPISpecPath = "api/openapi.yaml"

func registerSwaggerEndpoints(srv *khttp.Server) {
	if srv == nil {
		return
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL == nil:
			http.NotFound(w, r)
		case r.URL.Path == "/swagger" || r.URL.Path == "/swagger/" || strings.HasSuffix(r.URL.Path, "/index.html"):
			serveSwaggerIndex(w)
		case strings.HasSuffix(r.URL.Path, "/doc.json") || strings.HasSuffix(r.URL.Path, "/openapi.json"):
			serveOpenAPIJSON(w, r)
		case strings.HasSuffix(r.URL.Path, "/openapi.yaml"):
			serveOpenAPIYAML(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	srv.Handle("/swagger", handler)
	srv.HandlePrefix("/swagger/", handler)
}

func serveSwaggerIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>nai-tizi kratos swagger</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; background: #f5f7fb; color: #1f2937; }
    header { padding: 24px 32px; background: #111827; color: #fff; }
    main { max-width: 1080px; margin: 0 auto; padding: 24px 32px 48px; }
    .card { background: #fff; border-radius: 12px; box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08); padding: 20px 24px; margin-bottom: 24px; }
    a { color: #2563eb; text-decoration: none; }
    code { background: #eef2ff; border-radius: 6px; padding: 2px 6px; }
  </style>
  <script src="https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"></script>
</head>
<body>
  <header>
    <h1>nai-tizi kratos API</h1>
    <p>Generated from protobuf definitions.</p>
  </header>
  <main>
    <div class="card">
      <p>OpenAPI spec: <a href="/swagger/openapi.yaml"><code>/swagger/openapi.yaml</code></a></p>
      <p>OpenAPI JSON: <a href="/swagger/doc.json"><code>/swagger/doc.json</code></a></p>
      <p>If the embedded documentation does not render, you can still download the YAML spec directly.</p>
    </div>
    <redoc spec-url="/swagger/doc.json"></redoc>
  </main>
</body>
</html>`))
}

func serveOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	content, modTime, err := readOpenAPIContent()
	if err != nil {
		http.Error(w, "openapi spec not found, run `make proto-all` first", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	http.ServeContent(w, r, filepath.Base(cleanOpenAPISpecPath()), modTime, bytes.NewReader(content))
}

func serveOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	content, modTime, err := readOpenAPIContent()
	if err != nil {
		http.Error(w, "openapi spec not found, run `make proto-all` first", http.StatusNotFound)
		return
	}
	var payload any
	if err = yaml.Unmarshal(content, &payload); err != nil {
		http.Error(w, "failed to parse openapi spec", http.StatusInternalServerError)
		return
	}
	jsonContent, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode openapi spec", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	http.ServeContent(w, r, "openapi.json", modTime, bytes.NewReader(jsonContent))
}

func readOpenAPIContent() ([]byte, time.Time, error) {
	specPath := cleanOpenAPISpecPath()
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, time.Time{}, err
	}
	modTime := time.Time{}
	if info, statErr := os.Stat(specPath); statErr == nil {
		modTime = info.ModTime()
	}
	return content, modTime, nil
}

func cleanOpenAPISpecPath() string {
	specPath := openAPISpecPath
	if filepath.IsAbs(specPath) {
		return specPath
	}
	return filepath.Clean(specPath)
}
