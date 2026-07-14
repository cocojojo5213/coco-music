package charts

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cocojojo5213/coco-music/internal/upstream"
)

const (
	chartID   = "community-search"
	chartName = "站友搜索榜"
	chartDesc = "按大家实际搜索次数实时排序"
)

type Chart struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	UpdatedAt   string            `json:"updatedAt"`
	Items       []json.RawMessage `json:"items"`
}

type Service struct {
	up        *upstream.Client
	statsPath string

	mu    sync.Mutex
	stats *statsDoc

	boardMu     sync.Mutex
	board       Chart
	boardExp    time.Time
	boardTTL    time.Duration
	refreshing  bool
	buildMu     sync.Mutex // single in-flight full rebuild
}

type statsDoc struct {
	Version   int                 `json:"version"`
	UpdatedAt string              `json:"updatedAt"`
	Searches  map[string]termStat `json:"searches"`
}

type termStat struct {
	Term    string `json:"term"`
	Count   int    `json:"count"`
	LastAt  string `json:"lastAt"`
	FirstAt string `json:"firstAt"`
}

func New(up *upstream.Client, dataDir string) *Service {
	if dataDir == "" {
		dataDir = "data"
	}
	_ = os.MkdirAll(dataDir, 0o755)
	s := &Service{
		up:        up,
		statsPath: filepath.Join(dataDir, "search-stats.json"),
		boardTTL:  10 * time.Minute,
	}
	s.stats = s.loadStats()
	s.pruneNoiseStats()
	return s
}

// Warmup builds the board in the background so first visitor is not blocked.
func (s *Service) Warmup() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
		defer cancel()
		_, _ = s.Get(ctx, chartID, 24)
	}()
}

func (s *Service) Catalog() []map[string]string {
	return []map[string]string{{
		"id":          chartID,
		"name":        chartName,
		"description": chartDesc,
	}}
}

// RecordSearch increments local search count for the query.
func (s *Service) RecordSearch(term string) {
	term = normalizeTerm(term)
	if term == "" || utf8.RuneCountInString(term) > 40 || isNoiseTerm(term) {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	s.mu.Lock()
	if s.stats.Searches == nil {
		s.stats.Searches = map[string]termStat{}
	}
	entry := s.stats.Searches[term]
	if entry.Term == "" {
		entry = termStat{Term: term, FirstAt: now}
	}
	entry.Count++
	entry.LastAt = now
	s.stats.Searches[term] = entry
	s.stats.UpdatedAt = now
	_ = s.persistStatsLocked()
	s.mu.Unlock()

	// Soft-expire only: keep last board as stale fallback so users are never blocked 50s+.
	s.boardMu.Lock()
	s.boardExp = time.Time{}
	s.boardMu.Unlock()
	// async re-rank
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		_, _ = s.refresh(ctx, 24)
	}()
}

func (s *Service) List(ctx context.Context, perChart int) ([]Chart, error) {
	c, err := s.Get(ctx, chartID, perChart)
	if err != nil {
		return nil, err
	}
	return []Chart{c}, nil
}

func (s *Service) Get(ctx context.Context, id string, limit int) (Chart, error) {
	_ = id // single board
	if limit <= 0 {
		limit = 24
	}

	s.boardMu.Lock()
	fresh := time.Now().Before(s.boardExp) && len(s.board.Items) > 0
	stale := len(s.board.Items) > 0
	if fresh {
		c := clipChart(s.board, limit)
		s.boardMu.Unlock()
		return c, nil
	}
	// Stale-while-revalidate: return last good board immediately.
	if stale {
		c := clipChart(s.board, limit)
		needRefresh := !s.refreshing
		if needRefresh {
			s.refreshing = true
		}
		s.boardMu.Unlock()
		if needRefresh {
			go func() {
				rctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				_, _ = s.refresh(rctx, 24)
				s.boardMu.Lock()
				s.refreshing = false
				s.boardMu.Unlock()
			}()
		}
		return c, nil
	}
	s.boardMu.Unlock()

	// No cache yet — must build once (startup / first request).
	return s.refresh(ctx, limit)
}

func (s *Service) refresh(ctx context.Context, limit int) (Chart, error) {
	s.buildMu.Lock()
	defer s.buildMu.Unlock()

	// another waiter may have filled cache while we queued
	s.boardMu.Lock()
	if time.Now().Before(s.boardExp) && len(s.board.Items) > 0 {
		c := clipChart(s.board, limit)
		s.boardMu.Unlock()
		return c, nil
	}
	s.boardMu.Unlock()

	items, err := s.build(ctx, limit)
	if err != nil {
		// if build fails but we have stale, keep serving it
		s.boardMu.Lock()
		if len(s.board.Items) > 0 {
			c := clipChart(s.board, limit)
			s.boardMu.Unlock()
			return c, nil
		}
		s.boardMu.Unlock()
		return Chart{}, err
	}
	chart := Chart{
		ID:          chartID,
		Name:        chartName,
		Description: chartDesc,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		Items:       items,
	}
	s.boardMu.Lock()
	s.board = chart
	s.boardExp = time.Now().Add(s.boardTTL)
	s.boardMu.Unlock()
	return clipChart(chart, limit), nil
}

func clipChart(c Chart, limit int) Chart {
	if limit > 0 && len(c.Items) > limit {
		c.Items = append([]json.RawMessage{}, c.Items[:limit]...)
	}
	return c
}

func (s *Service) build(ctx context.Context, limit int) ([]json.RawMessage, error) {
	terms := s.topTerms(24)
	if len(terms) < 8 {
		if seeded := s.seedFromUpstream(ctx); len(seeded) > 0 {
			terms = s.topTerms(24)
		}
	}
	if len(terms) == 0 {
		now := time.Now().UTC().Format(time.RFC3339)
		s.mu.Lock()
		if s.stats.Searches == nil {
			s.stats.Searches = map[string]termStat{}
		}
		for _, t := range []string{"热歌", "流行", "周杰伦", "林俊杰", "邓紫棋", "李荣浩"} {
			term := normalizeTerm(t)
			if isNoiseTerm(term) {
				continue
			}
			s.stats.Searches[term] = termStat{Term: term, Count: 1, FirstAt: now, LastAt: now}
		}
		s.stats.UpdatedAt = now
		_ = s.persistStatsLocked()
		s.mu.Unlock()
		terms = s.topTerms(24)
	}

	// Resolve top terms in parallel (upstream search is the bottleneck).
	type result struct {
		idx  int
		ts   termStat
		item json.RawMessage
		ok   bool
	}
	workerN := 4
	if len(terms) < workerN {
		workerN = len(terms)
	}
	jobs := make(chan int, len(terms))
	outCh := make(chan result, len(terms))
	var wg sync.WaitGroup
	for w := 0; w < workerN; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				ts := terms[i]
				item, ok := s.resolveTerm(ctx, ts.Term)
				outCh <- result{idx: i, ts: ts, item: item, ok: ok}
			}
		}()
	}
	for i := range terms {
		jobs <- i
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(outCh)
	}()

	byIdx := make(map[int]result, len(terms))
	for r := range outCh {
		byIdx[r.idx] = r
	}

	seen := map[string]bool{}
	var out []json.RawMessage
	for i := 0; i < len(terms) && len(out) < limit; i++ {
		r, ok := byIdx[i]
		if !ok || !r.ok {
			continue
		}
		item := r.item
		// hard requirement: chart tracks must be browser-direct CDN only
		fixed, okDirect := upstream.EnsureTrackDirect(item)
		if !okDirect {
			continue
		}
		item = fixed
		key := trackDedupeKey(item)
		if key == "" || seen[key] || isWeakAlternate(item) {
			continue
		}
		seen[key] = true
		var obj map[string]any
		if json.Unmarshal(item, &obj) == nil {
			obj["rank"] = len(out) + 1
			obj["chartId"] = chartID
			obj["searchCount"] = r.ts.Count
			obj["searchTerm"] = r.ts.Term
			obj["clientDirect"] = true
			if b, err := json.Marshal(obj); err == nil {
				item = b
			}
		}
		out = append(out, item)
	}
	if out == nil {
		out = []json.RawMessage{}
	}
	return out, nil
}

func (s *Service) resolveTerm(ctx context.Context, term string) (json.RawMessage, bool) {
	termCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	raw, status, err := s.up.GetJSON(termCtx, "/api/music/search", url.Values{"q": {term}})
	if err != nil || status >= 400 {
		return nil, false
	}
	rewritten, err := upstream.RewriteTrackJSON(raw, s.up.Public())
	if err != nil {
		rewritten = raw
	}
	var payload struct {
		Items []json.RawMessage `json:"items"`
	}
	if json.Unmarshal(rewritten, &payload) != nil || len(payload.Items) == 0 {
		return nil, false
	}
	// Only accept tracks with third-party CDN direct URLs (no stream/proxy via our hosts).
	for _, item := range prioritizeItems(payload.Items, term) {
		if isWeakAlternate(item) {
			continue
		}
		if fixed, ok := upstream.EnsureTrackDirect(item); ok {
			return fixed, true
		}
	}
	// last pass: any non-weak with direct, even lower score
	for _, item := range payload.Items {
		if isWeakAlternate(item) {
			continue
		}
		if fixed, ok := upstream.EnsureTrackDirect(item); ok {
			return fixed, true
		}
	}
	return nil, false
}

func (s *Service) seedFromUpstream(ctx context.Context) []termStat {
	termCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	raw, status, err := s.up.GetJSON(termCtx, "/api/music/hot", url.Values{"count": {"12"}})
	if err != nil || status >= 400 {
		return nil
	}
	var payload struct {
		Stats struct {
			TopSearches []struct {
				Term  string `json:"term"`
				Count int    `json:"count"`
			} `json:"topSearches"`
		} `json:"stats"`
		Items []json.RawMessage `json:"items"`
	}
	if json.Unmarshal(raw, &payload) != nil {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stats.Searches == nil {
		s.stats.Searches = map[string]termStat{}
	}
	for _, t := range payload.Stats.TopSearches {
		term := normalizeTerm(t.Term)
		if term == "" || t.Count <= 0 || isNoiseTerm(term) {
			continue
		}
		seedCount := t.Count
		if seedCount > 50 {
			seedCount = 50 + (seedCount-50)/10
		}
		cur := s.stats.Searches[term]
		if cur.Term == "" {
			s.stats.Searches[term] = termStat{Term: term, Count: seedCount, FirstAt: now, LastAt: now}
		} else if cur.Count < seedCount {
			cur.Count = seedCount
			cur.LastAt = now
			s.stats.Searches[term] = cur
		}
	}
	for _, item := range payload.Items {
		var t struct {
			Title string `json:"title"`
		}
		if json.Unmarshal(item, &t) != nil {
			continue
		}
		term := normalizeTerm(normalizeTitle(t.Title))
		if term == "" || isNoiseTerm(term) {
			continue
		}
		if _, ok := s.stats.Searches[term]; !ok {
			s.stats.Searches[term] = termStat{Term: term, Count: 3, FirstAt: now, LastAt: now}
		}
	}
	s.stats.UpdatedAt = now
	_ = s.persistStatsLocked()

	out := make([]termStat, 0, len(s.stats.Searches))
	for _, v := range s.stats.Searches {
		out = append(out, v)
	}
	return out
}

func (s *Service) topTerms(limit int) []termStat {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := make([]termStat, 0, len(s.stats.Searches))
	for _, e := range s.stats.Searches {
		if e.Term != "" && e.Count > 0 && !isNoiseTerm(e.Term) {
			entries = append(entries, e)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].LastAt > entries[j].LastAt
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}

func (s *Service) loadStats() *statsDoc {
	doc := &statsDoc{Version: 1, Searches: map[string]termStat{}}
	b, err := os.ReadFile(s.statsPath)
	if err != nil {
		return doc
	}
	if json.Unmarshal(b, doc) != nil || doc.Searches == nil {
		return &statsDoc{Version: 1, Searches: map[string]termStat{}}
	}
	return doc
}

func (s *Service) pruneNoiseStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stats.Searches == nil {
		return
	}
	changed := false
	for k, v := range s.stats.Searches {
		term := normalizeTerm(v.Term)
		if term == "" || isNoiseTerm(term) {
			delete(s.stats.Searches, k)
			changed = true
			continue
		}
		if term != k || term != v.Term {
			delete(s.stats.Searches, k)
			v.Term = term
			s.stats.Searches[term] = v
			changed = true
		}
	}
	if changed {
		_ = s.persistStatsLocked()
	}
}

func (s *Service) persistStatsLocked() error {
	if len(s.stats.Searches) > 300 {
		entries := make([]termStat, 0, len(s.stats.Searches))
		for _, e := range s.stats.Searches {
			entries = append(entries, e)
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Count != entries[j].Count {
				return entries[i].Count > entries[j].Count
			}
			return entries[i].LastAt > entries[j].LastAt
		})
		keep := map[string]termStat{}
		for i, e := range entries {
			if i >= 200 {
				break
			}
			keep[e.Term] = e
		}
		s.stats.Searches = keep
	}
	b, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.statsPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.statsPath)
}

func normalizeTerm(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func isNoiseTerm(term string) bool {
	if term == "" {
		return true
	}
	// pure years
	if len(term) == 4 && term >= "1900" && term <= "2100" {
		return true
	}
	// catalog junk like "2026- 心雨-1"
	if strings.Contains(term, "2026-") || strings.Contains(term, "2025-") {
		return true
	}
	// mostly digits / id-like
	digits := 0
	letters := 0
	for _, r := range term {
		if unicode.IsDigit(r) {
			digits++
		}
		if unicode.IsLetter(r) || r > 127 {
			letters++
		}
	}
	if digits >= 4 && letters <= 2 {
		return true
	}
	for _, w := range []string{"dj", "女版", "男版", "节奏版", "默涵", "伴奏", "直播", "test"} {
		if strings.Contains(term, w) {
			return true
		}
	}
	return false
}

func trackDedupeKey(raw json.RawMessage) string {
	var t struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
		ID     string `json:"id"`
	}
	if json.Unmarshal(raw, &t) != nil {
		return ""
	}
	title := normalizeTitle(t.Title)
	if title == "" {
		return t.ID
	}
	return title
}

func normalizeTitle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, cut := range []string{"(", "（", "[", "【"} {
		if i := strings.Index(s, cut); i > 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return s
}

func isWeakAlternate(raw json.RawMessage) bool {
	var t struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
	}
	if json.Unmarshal(raw, &t) != nil {
		return false
	}
	s := strings.ToLower(t.Title + " " + t.Artist)
	for _, w := range []string{
		"dj", "女版", "男版", "翻自", "cover", "0.8", "伴奏", "直播",
		"现场", "remix", "加速", "降调", "节奏版", "氛围", "喊麦",
		"同学", "翻唱", "热搜版", "口风琴", "钢琴版", "古筝",
	} {
		if strings.Contains(s, w) {
			return true
		}
	}
	if strings.Contains(t.Title, "《") || strings.Contains(t.Title, "》") {
		return true
	}
	return false
}

func prioritizeItems(items []json.RawMessage, term string) []json.RawMessage {
	if len(items) <= 1 {
		return items
	}
	type scored struct {
		raw   json.RawMessage
		score int
	}
	term = strings.ToLower(strings.TrimSpace(term))
	parts := strings.Fields(term)
	out := make([]scored, 0, len(items))
	for _, item := range items {
		var t struct {
			Title  string `json:"title"`
			Artist string `json:"artist"`
		}
		_ = json.Unmarshal(item, &t)
		title := strings.ToLower(t.Title)
		artist := strings.ToLower(t.Artist)
		score := 0
		if len(parts) > 0 && (title == parts[0] || strings.HasPrefix(title, parts[0])) {
			score += 50
		}
		for _, p := range parts {
			if strings.Contains(title, p) {
				score += 10
			}
			if strings.Contains(artist, p) {
				score += 20
			}
		}
		if !isWeakAlternate(item) {
			score += 15
		}
		if upstream.TrackDirectURL(item) != "" {
			score += 120 // strongly prefer browser-direct CDN
		} else {
			score -= 250
		}
		score -= len([]rune(t.Title)) / 8
		out = append(out, scored{raw: item, score: score})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	res := make([]json.RawMessage, len(out))
	for i, s := range out {
		res[i] = s.raw
	}
	return res
}
