package service

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"golang.org/x/sync/singleflight"
)

const (
	kioskCacheTTL          = 5 * time.Minute
	kioskTopLimit          = 10
	kioskRedisKeyPrefix    = "kiosk:overview:"
	kioskUnknownChannelFmt = "channel-%d"
	kioskDefaultWindow     = "7d"
	kioskSeriesBucketHour  = int64(3600)
	kioskSeriesBucketDay   = int64(24 * 3600)
)

type kioskWindowConfig struct {
	id         string
	duration   time.Duration
	bucketSize int64
}

var kioskWindows = map[string]kioskWindowConfig{
	"24h": {id: "24h", duration: 24 * time.Hour, bucketSize: kioskSeriesBucketHour},
	"7d":  {id: "7d", duration: 7 * 24 * time.Hour, bucketSize: kioskSeriesBucketDay},
	"30d": {id: "30d", duration: 30 * 24 * time.Hour, bucketSize: kioskSeriesBucketDay},
}

type KioskHealth struct {
	RequestCount int64   `json:"request_count"`
	SuccessCount int64   `json:"success_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
}

type KioskRevenue struct {
	TodayAmount  float64                    `json:"today_amount"`
	WindowAmount float64                    `json:"window_amount"`
	TotalAmount  float64                    `json:"total_amount"`
	SuccessCount int64                      `json:"success_count"`
	PendingCount int64                      `json:"pending_count"`
	Series       []model.KioskRevenuePoint  `json:"series"`
	ByMethod     []model.KioskRevenueMethod `json:"by_method"`
}

type KioskPlatformTotals struct {
	Calls      int64 `json:"calls"`
	Quota      int64 `json:"quota"`
	TokenUsed  int64 `json:"token_used"`
	TotalUsers int64 `json:"total_users"`
}

type KioskOverviewResponse struct {
	GeneratedAt       int64                      `json:"generated_at"`
	Window            string                     `json:"window"`
	BucketSeconds     int64                      `json:"bucket_seconds"`
	DataExportEnabled bool                       `json:"data_export_enabled"`
	Today             *model.KioskUsageAggregate `json:"today"`
	Current           *model.KioskUsageAggregate `json:"current"`
	Previous          *model.KioskUsageAggregate `json:"previous"`
	Series            []model.KioskUsagePoint    `json:"series"`
	TopModels         []model.KioskDimensionItem `json:"top_models"`
	TopChannels       []model.KioskDimensionItem `json:"top_channels"`
	TopGroups         []model.KioskDimensionItem `json:"top_groups"`
	TopUsers          []model.KioskDimensionItem `json:"top_users"`
	Revenue           KioskRevenue               `json:"revenue"`
	Health            KioskHealth                `json:"health"`
	Totals            KioskPlatformTotals        `json:"totals"`
}

type kioskCacheItem struct {
	expiresAt time.Time
	data      *KioskOverviewResponse
}

var (
	kioskCacheMu sync.Mutex
	kioskCache   = map[string]kioskCacheItem{}
	kioskLoader  singleflight.Group
)

func GetKioskOverview(window string) (*KioskOverviewResponse, error) {
	config, err := resolveKioskWindow(window)
	if err != nil {
		return nil, err
	}

	// A kiosk polls on a timer, so concurrent misses for the same window must
	// collapse into one full-table aggregation instead of stampeding the database.
	shared, err, _ := kioskLoader.Do(config.id, func() (any, error) {
		return loadKioskOverview(config)
	})
	if err != nil {
		return nil, err
	}
	data, ok := shared.(*KioskOverviewResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected kiosk overview cache type for window %s", config.id)
	}
	return data, nil
}

func resolveKioskWindow(window string) (kioskWindowConfig, error) {
	if window == "" {
		window = kioskDefaultWindow
	}
	config, ok := kioskWindows[window]
	if !ok {
		return kioskWindowConfig{}, fmt.Errorf("invalid kiosk window: %s", window)
	}
	return config, nil
}

func loadKioskOverview(config kioskWindowConfig) (*KioskOverviewResponse, error) {
	now := time.Now()
	if item, ok := readKioskCache(config.id, now); ok {
		return item, nil
	}
	if data := readKioskRedis(config.id); data != nil {
		writeKioskCache(config.id, data, now)
		return data, nil
	}

	data, err := buildKioskOverview(config, now)
	if err != nil {
		return nil, err
	}
	writeKioskCache(config.id, data, now)
	writeKioskRedis(config.id, data)
	return data, nil
}

func readKioskCache(window string, now time.Time) (*KioskOverviewResponse, bool) {
	kioskCacheMu.Lock()
	defer kioskCacheMu.Unlock()
	item, ok := kioskCache[window]
	if !ok || !now.Before(item.expiresAt) {
		return nil, false
	}
	return item.data, true
}

func writeKioskCache(window string, data *KioskOverviewResponse, now time.Time) {
	kioskCacheMu.Lock()
	defer kioskCacheMu.Unlock()
	kioskCache[window] = kioskCacheItem{expiresAt: now.Add(kioskCacheTTL), data: data}
}

func readKioskRedis(window string) *KioskOverviewResponse {
	if !common.RedisEnabled {
		return nil
	}
	raw, err := common.RedisGet(kioskRedisKeyPrefix + window)
	if err != nil {
		return nil
	}
	data := &KioskOverviewResponse{}
	if err := common.Unmarshal([]byte(raw), data); err != nil {
		return nil
	}
	return data
}

func writeKioskRedis(window string, data *KioskOverviewResponse) {
	if !common.RedisEnabled {
		return
	}
	raw, err := common.Marshal(data)
	if err != nil {
		return
	}
	_ = common.RedisSet(kioskRedisKeyPrefix+window, string(raw), kioskCacheTTL)
}

func buildKioskOverview(config kioskWindowConfig, now time.Time) (*KioskOverviewResponse, error) {
	endTime := now.Unix()
	startTime := now.Add(-config.duration).Unix()
	previousEndTime := startTime - 1
	previousStartTime := time.Unix(startTime, 0).Add(-config.duration).Unix()
	todayStartTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	current, err := model.GetKioskUsageAggregate(startTime, endTime)
	if err != nil {
		return nil, err
	}
	previous, err := model.GetKioskUsageAggregate(previousStartTime, previousEndTime)
	if err != nil {
		return nil, err
	}
	today, err := model.GetKioskUsageAggregate(todayStartTime, endTime)
	if err != nil {
		return nil, err
	}
	series, err := model.GetKioskUsageSeries(startTime, endTime, config.bucketSize)
	if err != nil {
		return nil, err
	}
	topModels, err := model.GetKioskTopDimension(startTime, endTime, "model", kioskTopLimit)
	if err != nil {
		return nil, err
	}
	topGroups, err := model.GetKioskTopDimension(startTime, endTime, "group", kioskTopLimit)
	if err != nil {
		return nil, err
	}
	topUsers, err := model.GetKioskTopDimension(startTime, endTime, "user", kioskTopLimit)
	if err != nil {
		return nil, err
	}
	topChannels, err := model.GetKioskTopChannels(startTime, endTime, kioskTopLimit)
	if err != nil {
		return nil, err
	}
	if err := fillKioskChannelNames(topChannels); err != nil {
		return nil, err
	}
	revenue, err := buildKioskRevenue(config.bucketSize, startTime, endTime, todayStartTime)
	if err != nil {
		return nil, err
	}
	health, err := buildKioskHealth(startTime, endTime)
	if err != nil {
		return nil, err
	}
	totals, err := buildKioskTotals()
	if err != nil {
		return nil, err
	}

	return &KioskOverviewResponse{
		GeneratedAt:       endTime,
		Window:            config.id,
		BucketSeconds:     config.bucketSize,
		DataExportEnabled: common.DataExportEnabled,
		Today:             today,
		Current:           current,
		Previous:          previous,
		Series:            series,
		TopModels:         topModels,
		TopChannels:       topChannels,
		TopGroups:         topGroups,
		TopUsers:          topUsers,
		Revenue:           revenue,
		Health:            health,
		Totals:            totals,
	}, nil
}

func fillKioskChannelNames(channels []model.KioskDimensionItem) error {
	if len(channels) == 0 {
		return nil
	}
	channelIDs := make([]int64, 0, len(channels))
	for _, channel := range channels {
		channelIDs = append(channelIDs, channel.ID)
	}
	names, err := model.GetKioskChannelNames(channelIDs)
	if err != nil {
		return err
	}
	for index := range channels {
		if name := names[channels[index].ID]; name != "" {
			channels[index].Name = name
			continue
		}
		channels[index].Name = fmt.Sprintf(kioskUnknownChannelFmt, channels[index].ID)
	}
	return nil
}

func buildKioskRevenue(bucketSize int64, startTime int64, endTime int64, todayStartTime int64) (KioskRevenue, error) {
	revenue := KioskRevenue{Series: []model.KioskRevenuePoint{}, ByMethod: []model.KioskRevenueMethod{}}

	windowAggregate, err := model.GetKioskRevenueAggregate(startTime, endTime)
	if err != nil {
		return revenue, err
	}
	todayAggregate, err := model.GetKioskRevenueAggregate(todayStartTime, endTime)
	if err != nil {
		return revenue, err
	}
	totalAggregate, err := model.GetKioskRevenueAggregate(0, endTime)
	if err != nil {
		return revenue, err
	}
	pendingCount, err := model.GetKioskPendingRevenueCount()
	if err != nil {
		return revenue, err
	}
	series, err := model.GetKioskRevenueSeries(startTime, endTime, bucketSize)
	if err != nil {
		return revenue, err
	}
	byMethod, err := model.GetKioskRevenueByMethod(startTime, endTime)
	if err != nil {
		return revenue, err
	}

	revenue.WindowAmount = windowAggregate.Amount
	revenue.SuccessCount = windowAggregate.OrderCount
	revenue.TodayAmount = todayAggregate.Amount
	revenue.TotalAmount = totalAggregate.Amount
	revenue.PendingCount = pendingCount
	revenue.Series = series
	revenue.ByMethod = byMethod
	return revenue, nil
}

func buildKioskHealth(startTime int64, endTime int64) (KioskHealth, error) {
	aggregate, err := model.GetKioskHealthAggregate(startTime, endTime)
	if err != nil {
		return KioskHealth{}, err
	}
	health := KioskHealth{
		RequestCount: aggregate.RequestCount,
		SuccessCount: aggregate.SuccessCount,
	}
	if aggregate.RequestCount > 0 {
		health.SuccessRate = math.Round(float64(aggregate.SuccessCount)/float64(aggregate.RequestCount)*10000) / 100
		health.AvgLatencyMs = aggregate.TotalLatencyMs / aggregate.RequestCount
	}
	return health, nil
}

func buildKioskTotals() (KioskPlatformTotals, error) {
	calls, quota, tokenUsed, totalUsers, err := model.GetKioskPlatformTotals()
	if err != nil {
		return KioskPlatformTotals{}, err
	}
	return KioskPlatformTotals{
		Calls:      calls,
		Quota:      quota,
		TokenUsed:  tokenUsed,
		TotalUsers: totalUsers,
	}, nil
}
