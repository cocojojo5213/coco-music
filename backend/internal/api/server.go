package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cocojojo5213/coco-music/internal/charts"
	"github.com/cocojojo5213/coco-music/internal/config"
	"github.com/cocojojo5213/coco-music/internal/upstream"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	cfg    config.Config
	up     *upstream.Client
	charts *charts.Service
	http   http.Handler
}

func New(cfg config.Config, up *upstream.Client, chartSvc *charts.Service) *Server {
	s := &Server{cfg: cfg, up: up, charts: chartSvc}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(90 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Range"},
		ExposedHeaders:   []string{"Accept-Ranges", "Content-Length", "Content-Range", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/health", s.health)

	// coco-play music APIs (BFF)
	r.Get("/api/music/hot", s.musicHotPreferred)
	r.Get("/api/music/search", s.musicSearch)
	r.Post("/api/music/lyrics", s.musicLyrics)
	r.Post("/api/music/play", s.musicPlay)

	// 站友搜索榜（按搜索次数排序）
	r.Get("/api/music/charts", s.listCharts)
	r.Get("/api/music/charts/{id}", s.getChart)

	// media reverse-proxy (playback + client-side download; no server cache)
	r.Get("/api/music/stream", s.proxyPath)
	r.Get("/api/proxy", s.proxyPath)

	// legacy aliases
	r.Get("/api/tracks", s.musicHotAsTracks)
	r.Get("/api/search", s.legacySearch)

	if staticDir := envStatic(); staticDir != "" {
		fileServer(r, "/", http.Dir(staticDir))
	}

	s.http = r
	return s
}

func (s *Server) Handler() http.Handler { return s.http }

func envStatic() string {
	if v := os.Getenv("STATIC_DIR"); v != "" {
		return v
	}
	for _, c := range []string{"../frontend/dist", "./frontend/dist", "/app/frontend/dist"} {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return ""
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		r.Head(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		// never SPA-fallback API paths
		if strings.HasPrefix(req.URL.Path, "/api/") {
			writeJSON(w, 404, map[string]string{"error": "api not found"})
			return
		}
		fs := http.FileServer(root)
		upath := strings.TrimPrefix(req.URL.Path, path)
		if upath == "" {
			upath = "index.html"
		}
		if f, err := root.Open(upath); err != nil {
			req.URL.Path = "/"
			http.StripPrefix(strings.TrimSuffix(path, "/"), fs).ServeHTTP(w, req)
			return
		} else {
			_ = f.Close()
		}
		http.StripPrefix(strings.TrimSuffix(path, "/"), fs).ServeHTTP(w, req)
	}
	r.Get(path+"*", handler)
	r.Head(path+"*", handler)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=30")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeRaw(w http.ResponseWriter, status int, raw json.RawMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=30")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		_, _ = w.Write([]byte("\n"))
	}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"status":  "ok",
		"service": "coco-music",
		"brand":   "摇摆熊",
		// do not expose private upstream endpoints publicly
		"upstreamConfigured": strings.TrimSpace(s.cfg.CocoPlayBase) != "",
		"time":               time.Now().UTC().Format(time.RFC3339),
		"features": map[string]bool{
			"serverDownload":       false,
			"clientDownload":       true,
			"favoritesLocal":       true,
			"favoritesAccount":     false,
			"linuxdoLogin":         false,
			"charts":               true,
			"clientDirectDownload": true,
		},
	})
}

func (s *Server) listCharts(w http.ResponseWriter, r *http.Request) {
	if s.charts == nil {
		writeJSON(w, 503, map[string]string{"error": "charts unavailable"})
		return
	}
	count := 12
	list, err := s.charts.List(r.Context(), count)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"items": list})
}

func (s *Server) getChart(w http.ResponseWriter, r *http.Request) {
	if s.charts == nil {
		writeJSON(w, 503, map[string]string{"error": "charts unavailable"})
		return
	}
	id := chi.URLParam(r, "id")
	c, err := s.charts.Get(r.Context(), id, 30)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "chart not found"})
		return
	}
	writeJSON(w, 200, c)
}

// musicHot serves 站友搜索榜 (search-count ranking).
func (s *Server) musicHotPreferred(w http.ResponseWriter, r *http.Request) {
	if s.charts != nil {
		c, err := s.charts.Get(r.Context(), "community-search", 24)
		if err == nil && len(c.Items) > 0 {
			writeJSON(w, 200, map[string]any{
				"items":    c.Items,
				"provider": "coco-music-charts",
				"ranking":  c.ID,
				"note":     c.Name,
				"cached":   false,
			})
			return
		}
	}
	s.musicHot(w, r)
}

func (s *Server) musicHot(w http.ResponseWriter, r *http.Request) {
	s.proxyMusicJSON(w, r, "/api/music/hot")
}

func (s *Server) musicSearch(w http.ResponseWriter, r *http.Request) {
	if s.charts != nil {
		if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
			s.charts.RecordSearch(q)
		}
	}
	s.proxyMusicJSON(w, r, "/api/music/search")
}

func (s *Server) musicLyrics(w http.ResponseWriter, r *http.Request) {
	var body any
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
	raw, status, err := s.up.PostJSON(r.Context(), "/api/music/lyrics", body)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "upstream lyrics failed"})
		return
	}
	writeRaw(w, status, raw)
}

func (s *Server) musicPlay(w http.ResponseWriter, r *http.Request) {
	var body any
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
	raw, status, err := s.up.PostJSON(r.Context(), "/api/music/play", body)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "upstream play failed"})
		return
	}
	writeRaw(w, status, raw)
}

func (s *Server) proxyMusicJSON(w http.ResponseWriter, r *http.Request, path string) {
	raw, status, err := s.up.GetJSON(r.Context(), path, r.URL.Query())
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "upstream failed: " + err.Error()})
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}
	rewritten, err := upstream.RewriteTrackJSON(raw, s.up.Public())
	if err != nil {
		writeRaw(w, status, raw)
		return
	}
	writeRaw(w, status, rewritten)
}

func (s *Server) musicHotAsTracks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("count") == "" {
		q.Set("count", "24")
	}
	raw, status, err := s.up.GetJSON(r.Context(), "/api/music/hot", q)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "upstream failed"})
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}
	rewritten, err := upstream.RewriteTrackJSON(raw, s.up.Public())
	if err != nil {
		writeRaw(w, status, raw)
		return
	}
	var payload struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(rewritten, &payload); err != nil || payload.Items == nil {
		writeRaw(w, status, rewritten)
		return
	}
	out, _ := json.Marshal(payload.Items)
	writeRaw(w, 200, out)
}

func (s *Server) legacySearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	raw, status, err := s.up.GetJSON(r.Context(), "/api/music/search", q)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "upstream failed"})
		return
	}
	if status >= 400 {
		writeRaw(w, status, raw)
		return
	}
	rewritten, _ := upstream.RewriteTrackJSON(raw, s.up.Public())
	var payload struct {
		Items []json.RawMessage `json:"items"`
	}
	_ = json.Unmarshal(rewritten, &payload)
	writeJSON(w, 200, map[string]any{
		"tracks":  payload.Items,
		"albums":  []any{},
		"artists": []any{},
	})
}

func (s *Server) proxyPath(w http.ResponseWriter, r *http.Request) {
	pathQuery := r.URL.RequestURI()
	if !strings.HasPrefix(r.URL.Path, "/api/music/stream") && r.URL.Path != "/api/proxy" {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if r.URL.Path == "/api/proxy" {
		u := r.URL.Query().Get("url")
		if u != "" {
			if parsed, err := url.Parse(u); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				writeJSON(w, 400, map[string]string{"error": "invalid url"})
				return
			}
		}
	}

	resp, err := s.fetchUpstreamMedia(r, pathQuery, 0)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	for _, k := range []string{
		"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges",
		"Cache-Control", "ETag", "Last-Modified",
	} {
		if v := resp.Header.Get(k); v != "" {
			w.Header().Set(k, v)
		}
	}
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges, Content-Type")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) fetchUpstreamMedia(r *http.Request, pathQuery string, depth int) (*http.Response, error) {
	if depth > 4 {
		return nil, errString("too many redirects")
	}
	resp, err := s.up.Proxy(r.Context(), http.MethodGet, pathQuery, r.Header, nil)
	if err != nil {
		return nil, errString("proxy failed")
	}
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return resp, nil
	}
	loc := strings.TrimSpace(resp.Header.Get("Location"))
	_ = resp.Body.Close()
	if loc == "" {
		return nil, errString("empty redirect")
	}
	if strings.HasPrefix(loc, "/") {
		return s.fetchUpstreamMedia(r, loc, depth+1)
	}
	u, err := url.Parse(loc)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, errString("bad redirect")
	}
	if u.Path == "/api/music/stream" || u.Path == "/api/proxy" {
		pq := u.Path
		if u.RawQuery != "" {
			pq += "?" + u.RawQuery
		}
		return s.fetchUpstreamMedia(r, pq, depth+1)
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}
	if rng := r.Header.Get("Range"); rng != "" {
		req.Header.Set("Range", rng)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; coco-music/1.0)")
	req.Header.Set("Accept", "*/*")
	client := &http.Client{Timeout: 0, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errString("cdn redirect loop")
		}
		req.Header.Del("Referer")
		return nil
	}}
	return client.Do(req)
}

type errString string

func (e errString) Error() string { return string(e) }
