//go:build js && wasm

// Command searchwasm runs Clpr's search route, handler and repository from
// internal/search in js/wasm, against PGlite (PostgreSQL in WebAssembly)
// through a net.Conn that hands wire-protocol bytes to PGlite's
// execProtocolRaw. The portfolio demo loads it in a Web Worker.
//
// It builds only for js/wasm, so `go build ./...`, `go test ./...` and the
// API image skip it. Build and measure it with:
//
//	GOOS=js GOARCH=wasm go build -tags nomsgpack -ldflags="-s -w" -trimpath -o search.wasm ./cmd/searchwasm
//	gzip -9 -c search.wasm | wc -c
//
// The page sets globalThis.pglite to a PGlite instance before starting the
// program, then calls goHandle(method, url) once goReady is true.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"syscall/js"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/search"
)

// failingProvider stands in for an unreachable OpenSearch cluster (demo fault control)
type failingProvider struct{}

func (failingProvider) Search(context.Context, *models.SearchRequest) (*models.SearchResponse, error) {
	return nil, fmt.Errorf("dial tcp: connection refused")
}
func (failingProvider) GetSuggestions(context.Context, string, int) ([]models.SearchSuggestion, error) {
	return nil, fmt.Errorf("dial tcp: connection refused")
}
func (failingProvider) SearchWithScores(context.Context, *models.SearchRequest) (*models.SearchResponseWithScores, error) {
	return nil, fmt.Errorf("dial tcp: connection refused")
}

type pgliteConn struct {
	db      js.Value
	pending bytes.Buffer
	resp    bytes.Buffer
}

func (c *pgliteConn) Write(p []byte) (int, error) { return c.pending.Write(p) }
func (c *pgliteConn) Read(p []byte) (int, error) {
	for c.resp.Len() == 0 {
		if c.pending.Len() == 0 {
			return 0, io.EOF
		}
		msg := js.Global().Get("Uint8Array").New(c.pending.Len())
		js.CopyBytesToJS(msg, c.pending.Bytes())
		c.pending.Reset()
		out, err := await(c.db.Call("execProtocolRaw", msg))
		if err != nil {
			return 0, err
		}
		b := make([]byte, out.Get("length").Int())
		js.CopyBytesToGo(b, out)
		c.resp.Write(b)
	}
	return c.resp.Read(p)
}
func await(p js.Value) (js.Value, error) {
	done := make(chan struct{})
	var val js.Value
	var err error
	ok := js.FuncOf(func(_ js.Value, a []js.Value) any { val = a[0]; close(done); return nil })
	fail := js.FuncOf(func(_ js.Value, a []js.Value) any {
		err = fmt.Errorf("%s", a[0].Call("toString").String())
		close(done)
		return nil
	})
	defer ok.Release()
	defer fail.Release()
	p.Call("then", ok, fail)
	<-done
	return val, err
}
func (c *pgliteConn) Close() error                     { return nil }
func (c *pgliteConn) LocalAddr() net.Addr              { return addr{} }
func (c *pgliteConn) RemoteAddr() net.Addr             { return addr{} }
func (c *pgliteConn) SetDeadline(time.Time) error      { return nil }
func (c *pgliteConn) SetReadDeadline(time.Time) error  { return nil }
func (c *pgliteConn) SetWriteDeadline(time.Time) error { return nil }

type addr struct{}

func (addr) Network() string { return "pglite" }
func (addr) String() string  { return "pglite" }

func main() {
	cfg, err := pgxpool.ParseConfig("postgres://postgres@pglite/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	// PGlite is single-connection: one pooled conn, no health-check pings
	cfg.MaxConns = 1
	cfg.ConnConfig.LookupFunc = func(context.Context, string) ([]string, error) { return []string{"127.0.0.1"}, nil }
	db := js.Global().Get("pglite")
	cfg.ConnConfig.DialFunc = func(context.Context, string, string) (net.Conn, error) { return &pgliteConn{db: db}, nil }
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	repo := search.NewRepository(pool)
	h := search.NewHandler(repo)
	r.GET("/api/v1/search", h.Search)
	// Production's configured path when OpenSearch is reachable at startup: hybrid, no Postgres fallback.
	hybrid := search.NewHandlerWithHybridProvider(repo, failingProvider{})
	r.GET("/demo/fault/hybrid", hybrid.Search)

	js.Global().Set("goHandle", js.FuncOf(func(_ js.Value, a []js.Value) any {
		method, url := a[0].String(), a[1].String()
		executor := js.FuncOf(func(_ js.Value, pr []js.Value) any {
			resolve := pr[0]
			go func() {
				rec := httptest.NewRecorder()
				func() {
					defer func() {
						if p := recover(); p != nil {
							rec = httptest.NewRecorder()
							rec.WriteHeader(400)
							rec.WriteString(fmt.Sprint(p))
						}
					}()
					r.ServeHTTP(rec, httptest.NewRequest(method, url, nil))
				}()
				o := js.Global().Get("Object").New()
				o.Set("status", rec.Code)
				o.Set("body", rec.Body.String())
				hdr := js.Global().Get("Object").New()
				for k, v := range rec.Header() {
					hdr.Set(k, v[0])
				}
				o.Set("headers", hdr)
				resolve.Invoke(o)
			}()
			return nil
		})
		// Promise invokes its executor synchronously. The request goroutine
		// keeps resolve alive, but no longer needs the executor itself.
		defer executor.Release()
		return js.Global().Get("Promise").New(executor)
	}))
	js.Global().Set("goReady", true)
	select {}
}
