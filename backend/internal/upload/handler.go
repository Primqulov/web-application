// Package upload exposes REST endpoints for uploading files to S3 and
// removing them. Every upload is attributed to the authenticated user.
package upload

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ishchibormi/backend/pkg/httpx"
	"github.com/ishchibormi/backend/pkg/storage"
)

type Handler struct {
	Storage *storage.Service
}

func NewHandler(s *storage.Service) *Handler { return &Handler{Storage: s} }

// allowed kinds & their constraints
var allowed = map[string]struct {
	prefix   string
	mimes    []string
	maxBytes int64
}{
	"avatar":     {prefix: "avatars",      mimes: []string{"image/jpeg", "image/png", "image/webp"}, maxBytes: 5 << 20},
	"elon":       {prefix: "elons",        mimes: []string{"image/jpeg", "image/png", "image/webp"}, maxBytes: 8 << 20},
	"chat":       {prefix: "chat",         mimes: []string{"image/jpeg", "image/png", "image/webp", "application/pdf", "application/zip"}, maxBytes: 20 << 20},
}

// POST /api/uploads?kind=avatar|elon|chat   (multipart form, field: "file")
// Returns: {key, url}
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil {
		httpx.Err(w, httpx.NewError(503, "storage_disabled", "fayl yuklash sozlanmagan"))
		return
	}
	uid := httpx.UserID(r)
	if uid == "" {
		httpx.Err(w, httpx.NewError(401, "unauthorized", "kirish talab qilinadi"))
		return
	}
	kind := strings.ToLower(r.URL.Query().Get("kind"))
	rule, ok := allowed[kind]
	if !ok {
		httpx.Err(w, httpx.NewError(400, "bad_kind", "noma'lum fayl turi"))
		return
	}
	// Limit total request size (multipart overhead included).
	r.Body = http.MaxBytesReader(w, r.Body, rule.maxBytes+1<<20)
	if err := r.ParseMultipartForm(rule.maxBytes); err != nil {
		httpx.Err(w, httpx.NewError(413, "too_large", "fayl hajmi katta"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "no_file", "fayl topilmadi"))
		return
	}
	defer file.Close()

	if header.Size > rule.maxBytes {
		httpx.Err(w, httpx.NewError(413, "too_large", "fayl hajmi katta"))
		return
	}

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = guessContentType(header.Filename)
	}
	if !mimeOK(ct, rule.mimes) {
		httpx.Err(w, httpx.NewError(415, "bad_type", "fayl turi qabul qilinmaydi"))
		return
	}

	// Build a prefix that scopes the object to this user (and entity, if any).
	prefix := rule.prefix + "/" + uid
	if scope := r.URL.Query().Get("scope"); scope != "" {
		prefix = rule.prefix + "/" + sanitize(scope)
	}

	out, err := h.Storage.Upload(r.Context(), prefix, header.Filename, ct, file)
	if err != nil {
		httpx.Err(w, httpx.NewError(500, "upload_failed", err.Error()))
		return
	}
	httpx.JSON(w, 201, out)
}

// DELETE /api/uploads?key=...  or  DELETE /api/uploads?url=...
// Deletes the underlying S3 object. Authenticated users only — strict
// ownership check is performed by the calling domain (user/elon) before
// invoking this; this endpoint is best-effort for cleanup.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil {
		httpx.JSON(w, 200, map[string]bool{"ok": true})
		return
	}
	if httpx.UserID(r) == "" {
		httpx.Err(w, httpx.NewError(401, "unauthorized", "kirish talab qilinadi"))
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		key = h.Storage.KeyFromURL(r.URL.Query().Get("url"))
	}
	if key == "" {
		httpx.JSON(w, 200, map[string]bool{"ok": true})
		return
	}
	if err := h.Storage.Delete(r.Context(), key); err != nil {
		httpx.Err(w, httpx.NewError(500, "delete_failed", err.Error()))
		return
	}
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

func mimeOK(ct string, list []string) bool {
	ct = strings.ToLower(strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]))
	for _, m := range list {
		if m == ct {
			return true
		}
	}
	return false
}

func guessContentType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	}
	return ""
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "misc"
	}
	return b.String()
}

// Helper for other domains: silently best-effort delete by URL.
// Use this in user/elon delete paths so storage stays in sync with Mongo.
func DeleteByURL(s *storage.Service, url string) {
	if s == nil || url == "" {
		return
	}
	_ = s.DeleteByURL(context.Background(), url)
}
