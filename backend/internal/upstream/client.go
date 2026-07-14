package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base   string
	public string // coco-play public origin for absolute fallback API links
	http   *http.Client
}

func New(base, publicOrigin string) *Client {
	return &Client{
		base:   strings.TrimRight(base, "/"),
		public: strings.TrimRight(publicOrigin, "/"),
		http: &http.Client{
			Timeout: 45 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (c *Client) Base() string   { return c.base }
func (c *Client) Public() string { return c.public }

func (c *Client) GetJSON(ctx context.Context, path string, query url.Values) (json.RawMessage, int, error) {
	u := c.base + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return json.RawMessage(body), resp.StatusCode, nil
}

func (c *Client) PostJSON(ctx context.Context, path string, payload any) (json.RawMessage, int, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, bytes.NewReader(b))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return json.RawMessage(body), resp.StatusCode, nil
}

// Proxy streams a request to coco-play, preserving Range and selected headers.
// Redirects are NOT followed here — caller decides how to resolve Location.
func (c *Client) Proxy(ctx context.Context, method, pathWithQuery string, hdr http.Header, body io.Reader) (*http.Response, error) {
	u := c.base + pathWithQuery
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	copyHop(hdr, req.Header)
	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client.Do(req)
}

func copyHop(src, dst http.Header) {
	for _, k := range []string{"Range", "Accept", "Accept-Encoding", "If-Range", "User-Agent"} {
		if v := src.Get(k); v != "" {
			dst.Set(k, v)
		}
	}
}

// ExtractDirectMediaURL returns a browser-fetchable CDN/media URL without going
// through coco-music or coco-play reverse proxies.
func ExtractDirectMediaURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// relative /api/proxy?url=...
	if strings.HasPrefix(raw, "/api/proxy") {
		if u, err := url.Parse(raw); err == nil {
			if d := strings.TrimSpace(u.Query().Get("url")); isAbsoluteHTTP(d) && !isOwnMediaHost(d) {
				return d
			}
		}
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return ""
	}
	if u.Path == "/api/proxy" {
		if d := strings.TrimSpace(u.Query().Get("url")); isAbsoluteHTTP(d) && !isOwnMediaHost(d) {
			return d
		}
		return ""
	}
	// already absolute: only treat pure CDN/media as "direct"
	if isAbsoluteHTTP(raw) && !isOwnMediaHost(raw) && !isProxyOrStreamPath(u.Path) {
		return raw
	}
	return ""
}

func isAbsoluteHTTP(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func isOwnMediaHost(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	h := strings.ToLower(u.Hostname())
	return strings.Contains(h, "52131415.xyz") ||
		strings.Contains(h, "coco5213.me") ||
		strings.Contains(h, "coco5213.io") ||
		h == "localhost" ||
		h == "127.0.0.1" ||
		strings.HasPrefix(h, "100.")
}

func isProxyOrStreamPath(path string) bool {
	return path == "/api/proxy" || path == "/api/music/stream" || strings.HasPrefix(path, "/api/music/stream/")
}

// AbsolutizeAPIURL turns coco-play relative API paths into public absolute URLs
// so the browser does not hit the coco-music BFF for media fallbacks.
func AbsolutizeAPIURL(raw, publicOrigin string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if isAbsoluteHTTP(raw) {
		return raw
	}
	if strings.HasPrefix(raw, "/api/") {
		return strings.TrimRight(publicOrigin, "/") + raw
	}
	return raw
}

// RewriteTrackJSON preserves CDN direct links for client-side download/play.
// It no longer wraps external media into /api/proxy (that would burn server egress).
func RewriteTrackJSON(raw json.RawMessage, publicOrigin string) (json.RawMessage, error) {
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return raw, err
	}
	rewriteValue(root, strings.TrimRight(publicOrigin, "/"))
	return json.Marshal(root)
}

func rewriteValue(v any, public string) {
	switch t := v.(type) {
	case map[string]any:
		origURL, _ := t["url"].(string)
		origProxy, _ := t["proxyUrl"].(string)

		direct := ExtractDirectMediaURL(origURL)
		if direct == "" {
			direct = ExtractDirectMediaURL(origProxy)
		}
		if direct != "" {
			t["directUrl"] = direct
			// Prefer CDN direct for both play + download to avoid server egress.
			t["url"] = direct
		} else if origURL != "" {
			// e.g. signed stream on coco-play — point at public play host, not music BFF
			t["url"] = AbsolutizeAPIURL(origURL, public)
		}

		if origProxy != "" {
			// keep as absolute public fallback only (still not used for download)
			if d := ExtractDirectMediaURL(origProxy); d != "" {
				// proxy was just a wrapper around CDN; no need to keep own proxy
				if t["proxyUrl"] == nil || t["proxyUrl"] == "" {
					t["proxyUrl"] = ""
				}
				// leave proxy empty when we already have directUrl
				if direct != "" {
					delete(t, "proxyUrl")
				} else {
					t["proxyUrl"] = AbsolutizeAPIURL(origProxy, public)
				}
			} else {
				t["proxyUrl"] = AbsolutizeAPIURL(origProxy, public)
			}
		}

		// duration convenience for frontend
		if _, ok := t["duration"]; !ok {
			if ms, ok := t["durationMs"].(float64); ok && ms > 0 {
				t["duration"] = int(ms / 1000)
			}
		}

		// covers: prefer original artwork CDN, not reverse-proxy
		if cover, ok := t["artwork"].(string); ok && cover != "" {
			t["coverUrl"] = cover
		}
		if ap, ok := t["artworkProxyUrl"].(string); ok && ap != "" {
			if d := ExtractDirectMediaURL(ap); d != "" {
				t["artworkProxyUrl"] = d
			} else if !isAbsoluteHTTP(ap) {
				// drop relative artwork proxy to avoid music BFF image egress
				if cover, ok := t["artwork"].(string); ok && cover != "" {
					t["artworkProxyUrl"] = cover
				} else {
					delete(t, "artworkProxyUrl")
				}
			}
		}

		for _, child := range t {
			rewriteValue(child, public)
		}
	case []any:
		for _, child := range t {
			rewriteValue(child, public)
		}
	}
}
