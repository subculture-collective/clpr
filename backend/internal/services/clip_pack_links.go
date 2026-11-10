package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// ClipPackLinks signs and checks the expiring download links of clip packs
// (#86). The API signs them for the broadcaster; the render host checks
// them before serving the zip. Both need the same Secret.
type ClipPackLinks struct {
	Secret []byte
	// BaseURL is where the render host serves packs, for example
	// https://packs.clpr.tv. Links are BaseURL/packs/<id>.zip.
	BaseURL string
	// TTL is how long a link lasts; default 24 hours. It never outlasts the
	// pack.
	TTL time.Duration
}

func (l ClipPackLinks) signature(id uuid.UUID, expires int64) string {
	mac := hmac.New(sha256.New, l.Secret)
	mac.Write([]byte("clip-pack:" + id.String() + ":" + strconv.FormatInt(expires, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

// URL returns a link to the pack's zip that expires after TTL, or when the
// pack does, whichever is sooner.
func (l ClipPackLinks) URL(id uuid.UUID, packExpires, now time.Time) string {
	ttl := l.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	expires := now.Add(ttl)
	if packExpires.Before(expires) {
		expires = packExpires
	}
	unix := expires.Unix()
	return strings.TrimRight(l.BaseURL, "/") + "/packs/" + id.String() + ".zip?expires=" + strconv.FormatInt(unix, 10) +
		"&sig=" + l.signature(id, unix)
}

// Valid reports whether expires and sig were signed for the pack and the
// link has not expired.
func (l ClipPackLinks) Valid(id uuid.UUID, expires, sig string, now time.Time) bool {
	if len(l.Secret) == 0 {
		return false
	}
	unix, err := strconv.ParseInt(expires, 10, 64)
	if err != nil || now.Unix() >= unix {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(l.signature(id, unix)))
}

// ClipPackFiles looks up a pack's zip.
type ClipPackFiles interface {
	PackFile(ctx context.Context, id uuid.UUID) (*models.ClipPackFile, error)
}

// ClipPackServer serves GET /packs/<id>.zip?expires=…&sig=… from the pack
// directory, for a valid link to a ready pack that has not expired.
type ClipPackServer struct {
	Store ClipPackFiles
	Dir   string
	Links ClipPackLinks
	Now   func() time.Time
}

func (s *ClipPackServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rest, ok := strings.CutPrefix(r.URL.Path, "/packs/")
	raw, isZip := strings.CutSuffix(rest, ".zip")
	id, err := uuid.Parse(raw)
	if !ok || !isZip || err != nil {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query()
	if !s.Links.Valid(id, query.Get("expires"), query.Get("sig"), now) {
		http.Error(w, "this link has expired or is not valid", http.StatusForbidden)
		return
	}
	file, err := s.Store.PackFile(r.Context(), id)
	if err != nil {
		utils.Error("Clip pack lookup failed", err, map[string]interface{}{"pack_id": id.String()})
		http.Error(w, "the pack could not be loaded", http.StatusInternalServerError)
		return
	}
	if file == nil || file.Status != models.ClipPackReady || file.Deleted || file.FileName == "" ||
		file.ExpiresAt == nil || !now.Before(*file.ExpiresAt) {
		http.Error(w, "this pack is no longer available", http.StatusGone)
		return
	}
	root, err := os.OpenRoot(s.Dir)
	if err != nil {
		utils.Error("Clip pack directory unavailable", err, nil)
		http.Error(w, "the pack could not be loaded", http.StatusInternalServerError)
		return
	}
	defer root.Close()
	name := filepath.ToSlash(filepath.Clean(file.FileName))
	f, err := root.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		http.Error(w, "this pack is no longer available", http.StatusGone)
		return
	}
	if err != nil {
		utils.Error("Clip pack open failed", err, map[string]interface{}{"pack_id": id.String()})
		http.Error(w, "the pack could not be loaded", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "the pack could not be loaded", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(name)+`"`)
	http.ServeContent(w, r, filepath.Base(name), info.ModTime(), f)
}
