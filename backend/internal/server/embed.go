package server

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:web
var embeddedWeb embed.FS

// EmbeddedWebFS returns the embedded console assets rooted at "web".
// It is empty when the binary was built without running the console export
// step first; in that case the caller should fall back to -web on disk.
func EmbeddedWebFS() (fs.FS, bool) {
	sub, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}

// EmbeddedWebHandler serves the embedded console with Next.js static-export
// trailing-slash semantics. It bypasses http.FileServer's automatic redirect
// from /index.html → /, which would otherwise turn a request for "/" into a
// 301 with no body.
func EmbeddedWebHandler(sub fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if p == "" {
			serveFile(w, r, sub, "index.html")
			return
		}
		if info, err := fs.Stat(sub, p); err == nil {
			if info.IsDir() {
				if _, err := fs.Stat(sub, p+"/index.html"); err == nil {
					serveFile(w, r, sub, p+"/index.html")
					return
				}
				http.NotFound(w, r)
				return
			}
			serveFile(w, r, sub, p)
			return
		}
		// Next.js export wrote /apps/index.html for the /apps route, so a
		// request for "apps" without a trailing slash should still resolve.
		if _, err := fs.Stat(sub, p+"/index.html"); err == nil {
			serveFile(w, r, sub, p+"/index.html")
			return
		}
		http.NotFound(w, r)
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, sub fs.FS, name string) {
	f, err := sub.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if seeker, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, r, name, info.ModTime(), seeker)
		return
	}
	_, _ = io.Copy(w, f)
}
