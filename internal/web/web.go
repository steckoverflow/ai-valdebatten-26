// Package web embeds the built Svelte single-page app and serves it.
//
// The Vite build writes its output into the dist/ directory beside this file
// (see frontend/vite.config.ts -> build.outDir). go:embed then bakes those
// files into the binary, so the final server is a single self-contained
// executable that serves both the UI and the WebSocket endpoint.
//
// dist/ is a build artifact and is gitignored except for a .gitkeep, so the
// embed compiles on a fresh checkout even before the frontend is built. When no
// real index.html is present, Handler serves a small fallback page.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

// distFS holds the embedded production build. The "all:" prefix ensures files
// beginning with "." or "_" (like .gitkeep) are included too.
//
//go:embed all:dist
var distFS embed.FS

const fallbackHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>aivaldebatten</title></head>
<body style="font-family:system-ui;background:#020617;color:#e2e8f0;padding:2rem">
<h1>aivaldebatten</h1>
<p>The backend is running, but the frontend has not been built yet.</p>
<p>Run <code>make build-frontend</code> (or <code>npm run build</code> in <code>frontend/</code>) and rebuild.</p>
</body></html>`

// Handler returns an http.Handler that serves the SPA. Unknown paths fall back
// to index.html so client-side routing works on deep links. If no built
// index.html is embedded, a small placeholder page is served instead.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err) // only fails if the embed is misconfigured at build time
	}

	// Detect whether a real build is present.
	_, statErr := fs.Stat(sub, "index.html")
	hasIndex := statErr == nil

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasIndex {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(fallbackHTML))
			return
		}

		// Serve the requested file if it exists; otherwise serve index.html so
		// the SPA can handle the route.
		path := trimLeadingSlash(r.URL.Path)
		if path != "" {
			if _, err := fs.Stat(sub, path); err != nil {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func trimLeadingSlash(p string) string {
	if len(p) > 0 && p[0] == '/' {
		return p[1:]
	}
	return p
}
