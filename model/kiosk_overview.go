package model

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

// KioskUsageAggregate is the platform-wide usage total for one time window.
type KioskUsageAggregate struct {
	Calls               int64 `json:"calls" gorm:"column:calls"`
	Quota               int64 `json:"quota" gorm:"column:quota"`
	TokenUsed           int64 `json:"token_used" gorm:"column:token_used"`
	CacheReadTokens     int64 `json:"cache_read_tokens" gorm:"column:cache_read_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens" gorm:"column:cache_creation_tokens"`
	ActiveUsers         int64 `json:"active_users" gorm:"column:active_users"`
}

// KioskUsagePoint is one time bucket of the usage series.
type KioskUsagePoint struct {
	Bucket    int64 `json:"bucket" gorm:"column:bucket"`
	Calls     int64 `json:"calls" gorm:"column:calls"`
	Quota     int64 `json:"quota" gorm:"column:quota"`
	TokenUsed int64 `json:"token_used" gorm:"column:token_used"`
}

// KioskDimensionItem is a Top N row for one grouping dimension.
type KioskDimensionItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Calls     int64  `json:"calls"`
	Quota     int64  `json:"quota"`
	TokenUsed int64  `json:"token_used"`
}

// KioskRevenueAggregate is the successful top-up total for one time window.
type KioskRevenueAggregate struct {
	Amount     float64 `json:"amount" gorm:"column:amount"`
	OrderCount int64   `json:"order_count" gorm:"column:order_count"`
}

// KioskRevenuePoint is one time bucket of the revenue series.
type KioskRevenuePoint struct {
	Bucket     int64   `json:"bucket" gorm:"column:bucket"`
	Amount     float64 `json:"amount" gorm:"column:amount"`
	OrderCount int64   `json:"order_count" gorm:"column:order_count"`
}

// KioskRevenueMethod is the successful top-up total for one payment method.
type KioskRevenueMethod struct {
	Method     string  `json:"method" gorm:"column:method"`
	Amount     float64 `json:"amount" gorm:"column:amount"`
	OrderCount int64   `json:"order_count" gorm:"column:order_count"`
	Quota      int64   `json:"quota" gorm:"column:quota"`
}

// KioskHealthAggregate is the platform relay success and latency total for one
// time window, used by the kiosk header without per-model breakdown.
type KioskHealthAggregate struct {
	RequestCount   int64 `json:"request_count" gorm:"column:request_count"`
	SuccessCount   int64 `json:"success_count" gorm:"column:success_count"`
	TotalLatencyMs int64 `json:"total_latency_ms" gorm:"column:total_latency_ms"`
}

var kioskDimensionColumns = map[string]string{
	"model": "model_name",
	"group": "use_group",
	"user":  "username",
}

func kioskBucketExpr(column string, bucketSize int64) string {
	return fmt.Sprintf("%s - (%s %% %d)", column, column, bucketSize)
}

func GetKioskUsageAggregate(startTime int64, endTime int64) (*KioskUsageAggregate, error) {
	aggregate := &KioskUsageAggregate{}
	err := DB.Table("quota_data").
		Select("COALESCE(SUM(count), 0) as calls, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(token_used), 0) as token_used, "+
			"COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens, COALESCE(SUM(cache_creation_tokens), 0) as cache_creation_tokens, "+
			"COUNT(DISTINCT user_id) as active_users").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Scan(aggregate).Error
	return aggregate, err
}

func GetKioskUsageSeries(startTime int64, endTime int64, bucketSize int64) ([]KioskUsagePoint, error) {
	bucketExpr := kioskBucketExpr("created_at", bucketSize)
	var points []KioskUsagePoint
	err := DB.Table("quota_data").
		Select(bucketExpr+" as bucket, COALESCE(SUM(count), 0) as calls, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(token_used), 0) as token_used").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group(bucketExpr).
		Order("bucket ASC").
		Scan(&points).Error
	return points, err
}

// GetKioskTopDimension ranks a whitelisted quota_data grouping column. The
// column name never comes from the caller, so it cannot become raw SQL input.
func GetKioskTopDimension(startTime int64, endTime int64, dimension string, limit int) ([]KioskDimensionItem, error) {
	column, ok := kioskDimensionColumns[dimension]
	if !ok {
		return nil, fmt.Errorf("invalid kiosk dimension: %s", dimension)
	}
	if limit <= 0 {
		return nil, nil
	}
	var rows []struct {
		Name      string `gorm:"column:dimension_name"`
		Calls     int64  `gorm:"column:calls"`
		Quota     int64  `gorm:"column:quota"`
		TokenUsed int64  `gorm:"column:token_used"`
	}
	err := DB.Table("quota_data").
		Select(column+" as dimension_name, COALESCE(SUM(count), 0) as calls, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(token_used), 0) as token_used").
		Where("created_at >= ? AND created_at <= ? AND "+column+" <> ''", startTime, endTime).
		Group(column).
		Order("token_used DESC, calls DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]KioskDimensionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, KioskDimensionItem{Name: row.Name, Calls: row.Calls, Quota: row.Quota, TokenUsed: row.TokenUsed})
	}
	return items, nil
}

func GetKioskTopChannels(startTime int64, endTime int64, limit int) ([]KioskDimensionItem, error) {
	if limit <= 0 {
		return nil, nil
	}
	var rows []struct {
		ChannelID int64 `gorm:"column:channel_id"`
		Calls     int64 `gorm:"column:calls"`
		Quota     int64 `gorm:"column:quota"`
		TokenUsed int64 `gorm:"column:token_used"`
	}
	err := DB.Table("quota_data").
		Select("channel_id, COALESCE(SUM(count), 0) as calls, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(token_used), 0) as token_used").
		Where("created_at >= ? AND created_at <= ? AND channel_id > 0", startTime, endTime).
		Group("channel_id").
		Order("token_used DESC, calls DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]KioskDimensionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, KioskDimensionItem{ID: row.ChannelID, Calls: row.Calls, Quota: row.Quota, TokenUsed: row.TokenUsed})
	}
	return items, nil
}

func GetKioskRevenueAggregate(startTime int64, endTime int64) (*KioskRevenueAggregate, error) {
	aggregate := &KioskRevenueAggregate{}
	err := DB.Model(&TopUp{}).
		Select("COALESCE(SUM(money), 0) as amount, COUNT(id) as order_count").
		Where("status = ? AND complete_time >= ? AND complete_time <= ?", common.TopUpStatusSuccess, startTime, endTime).
		Scan(aggregate).Error
	return aggregate, err
}

func GetKioskPendingRevenueCount() (int64, error) {
	var count int64
	err := DB.Model(&TopUp{}).
		Where("status = ?", common.TopUpStatusPending).
		Count(&count).Error
	return count, err
}

func GetKioskRevenueSeries(startTime int64, endTime int64, bucketSize int64) ([]KioskRevenuePoint, error) {
	bucketExpr := kioskBucketExpr("complete_time", bucketSize)
	var points []KioskRevenuePoint
	err := DB.Model(&TopUp{}).
		Select(bucketExpr+" as bucket, COALESCE(SUM(money), 0) as amount, COUNT(id) as order_count").
		Where("status = ? AND complete_time >= ? AND complete_time <= ?", common.TopUpStatusSuccess, startTime, endTime).
		Group(bucketExpr).
		Order("bucket ASC").
		Scan(&points).Error
	return points, err
}

func GetKioskRevenueByMethod(startTime int64, endTime int64) ([]KioskRevenueMethod, error) {
	var methods []KioskRevenueMethod
	err := DB.Model(&TopUp{}).
		Select("payment_method as method, COALESCE(SUM(money), 0) as amount, COUNT(id) as order_count, COALESCE(SUM(amount), 0) as quota").
		Where("status = ? AND complete_time >= ? AND complete_time <= ?", common.TopUpStatusSuccess, startTime, endTime).
		Group("payment_method").
		Order("amount DESC").
		Scan(&methods).Error
	return methods, err
}

func GetKioskHealthAggregate(startTime int64, endTime int64) (*KioskHealthAggregate, error) {
	aggregate := &KioskHealthAggregate{}
	err := DB.Model(&PerfMetric{}).
		Select("COALESCE(SUM(request_count), 0) as request_count, COALESCE(SUM(success_count), 0) as success_count, COALESCE(SUM(total_latency_ms), 0) as total_latency_ms").
		Where("bucket_ts >= ? AND bucket_ts <= ?", startTime, endTime).
		Scan(aggregate).Error
	return aggregate, err
}

// GetKioskChannelNames resolves channel ids to display names in a single query,
// avoiding the per-id lookups that CacheGetChannel costs when memory cache is off.
func GetKioskChannelNames(channelIDs []int64) (map[int64]string, error) {
	names := make(map[int64]string, len(channelIDs))
	if len(channelIDs) == 0 {
		return names, nil
	}
	var channels []struct {
		Id   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	if err := DB.Table("channels").Select("id, name").Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
		return nil, err
	}
	for _, channel := range channels {
		names[channel.Id] = channel.Name
	}
	return names, nil
}

func GetKioskPlatformTotals() (calls int64, quota int64, tokenUsed int64, totalUsers int64, err error) {
	var aggregate struct {
		Calls     int64 `gorm:"column:calls"`
		Quota     int64 `gorm:"column:quota"`
		TokenUsed int64 `gorm:"column:token_used"`
	}
	if err = DB.Table("quota_data").
		Select("COALESCE(SUM(count), 0) as calls, COALESCE(SUM(quota), 0) as quota, COALESCE(SUM(token_used), 0) as token_used").
		Scan(&aggregate).Error; err != nil {
		return 0, 0, 0, 0, err
	}
	if err = DB.Model(&User{}).Count(&totalUsers).Error; err != nil {
		return 0, 0, 0, 0, err
	}
	return aggregate.Calls, aggregate.Quota, aggregate.TokenUsed, totalUsers, nil
}
