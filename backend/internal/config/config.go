package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr            string
	CocoPlayBase    string
	UpstreamPublic  string
	PublicOrigin    string
	DataDir         string
	CORS            []string
}

func Load() Config {
	public := strings.TrimRight(env("PUBLIC_ORIGIN", "https://music.52131415.xyz"), "/")
	// Upstream endpoints must come from env / local deploy files — never ship real internals in git defaults.
	base := strings.TrimRight(env("COCO_PLAY_BASE", ""), "/")
	upPublic := strings.TrimRight(env("UPSTREAM_PUBLIC", env("COCO_PLAY_PUBLIC", "")), "/")
	return Config{
		Addr:           env("ADDR", ":18280"),
		CocoPlayBase:   base,
		UpstreamPublic: upPublic,
		PublicOrigin:   public,
		DataDir:        env("DATA_DIR", "data"),
		CORS: splitCSV(env("CORS_ORIGINS", strings.Join([]string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:18280",
			public,
		}, ","))),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
