package main

import (
	"log"
	"net/http"

	"github.com/cocojojo5213/coco-music/internal/api"
	"github.com/cocojojo5213/coco-music/internal/charts"
	"github.com/cocojojo5213/coco-music/internal/config"
	"github.com/cocojojo5213/coco-music/internal/upstream"
)

func main() {
	cfg := config.Load()
	if cfg.CocoPlayBase == "" {
		log.Fatal("COCO_PLAY_BASE is required (set in local env / systemd, not committed)")
	}
	up := upstream.New(cfg.CocoPlayBase, cfg.UpstreamPublic)
	chartSvc := charts.New(up, cfg.DataDir)
	srv := api.New(cfg, up, chartSvc)
	log.Printf("coco-music listening on %s (upstream configured=%v data=%s)", cfg.Addr, cfg.CocoPlayBase != "", cfg.DataDir)
	if err := http.ListenAndServe(cfg.Addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
