package model

import (
	"os"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Hour-aligned fixtures below rely on 3600 buckets:
// created 3600/3601/3700/7199 collapse into bucket 3600, created 7200 into bucket 7200.
func seedKioskQuotaData(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.Create(&Channel{Id: 1, Name: "east"}).Error)
	require.NoError(t, DB.Create(&Channel{Id: 2, Name: "west"}).Error)

	rows := []QuotaData{
		{UserID: 1, Username: "alice", ModelName: "gpt-a", UseGroup: "vip", ChannelID: 1, CreatedAt: 3600, Count: 2, Quota: 100, TokenUsed: 40, CacheReadTokens: 5, CacheCreationTokens: 7},
		{UserID: 1, Username: "alice", ModelName: "gpt-a", UseGroup: "vip", ChannelID: 1, CreatedAt: 3601, Count: 1, Quota: 50, TokenUsed: 20, CacheReadTokens: 1, CacheCreationTokens: 2},
		{UserID: 2, Username: "bob", ModelName: "gpt-b", UseGroup: "free", ChannelID: 2, CreatedAt: 7199, Count: 5, Quota: 20, TokenUsed: 60},
		{UserID: 3, Username: "carol", ModelName: "gpt-a", UseGroup: "free", ChannelID: 1, CreatedAt: 7200, Count: 4, Quota: 8, TokenUsed: 8},
		// Missing dimensions stay out of Top N while still counting toward totals.
		{UserID: 4, Username: "dave", ModelName: "", UseGroup: "", ChannelID: 0, CreatedAt: 3700, Count: 9, Quota: 9, TokenUsed: 9},
	}
	for _, row := range rows {
		seed := row
		require.NoError(t, DB.Create(&seed).Error)
	}
}

func seedKioskTopUps(t *testing.T) {
	t.Helper()
	topUps := []TopUp{
		{UserId: 1, Amount: 500000, Money: 10.5, TradeNo: "t-stripe-1", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CompleteTime: 3600},
		{UserId: 2, Amount: 1000000, Money: 20.25, TradeNo: "t-stripe-2", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CompleteTime: 7199},
		{UserId: 3, Amount: 250000, Money: 100, TradeNo: "t-waffo-old", PaymentMethod: "waffo", Status: common.TopUpStatusSuccess, CompleteTime: 0},
		{UserId: 4, Amount: 100000, Money: 2, TradeNo: "t-creem-late", PaymentMethod: "creem", Status: common.TopUpStatusSuccess, CompleteTime: 7200},
		// Non-success orders must never appear in revenue.
		{UserId: 5, Amount: 50000, Money: 5, TradeNo: "t-pending", PaymentMethod: "epay", Status: common.TopUpStatusPending, CompleteTime: 4000},
		{UserId: 5, Amount: 50000, Money: 7, TradeNo: "t-failed", PaymentMethod: "epay", Status: common.TopUpStatusFailed, CompleteTime: 5000},
		{UserId: 5, Amount: 50000, Money: 3, TradeNo: "t-expired", PaymentMethod: "waffo", Status: common.TopUpStatusExpired, CompleteTime: 5000},
	}
	for _, topUp := range topUps {
		seed := topUp
		require.NoError(t, DB.Create(&seed).Error)
	}
}

func TestGetKioskAggregatesOnEmptyTablesReturnZero(t *testing.T) {
	truncateTables(t)

	usage, err := GetKioskUsageAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, &KioskUsageAggregate{}, usage)

	series, err := GetKioskUsageSeries(0, 10000, 3600)
	require.NoError(t, err)
	require.Empty(t, series)

	revenue, err := GetKioskRevenueAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, &KioskRevenueAggregate{}, revenue)

	health, err := GetKioskHealthAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, &KioskHealthAggregate{}, health)
}

func TestGetKioskUsageAggregateSumsWindowAndDistinctUsers(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)

	window, err := GetKioskUsageAggregate(3600, 7199)
	require.NoError(t, err)
	require.Equal(t, &KioskUsageAggregate{
		Calls: 17, Quota: 179, TokenUsed: 129,
		CacheReadTokens: 6, CacheCreationTokens: 9, ActiveUsers: 3,
	}, window)

	all, err := GetKioskUsageAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, int64(21), all.Calls)
	require.Equal(t, int64(187), all.Quota)
	require.Equal(t, int64(137), all.TokenUsed)
	require.Equal(t, int64(4), all.ActiveUsers)
}

func TestGetKioskUsageSeriesBucketsByModulo(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)

	points, err := GetKioskUsageSeries(0, 10000, 3600)
	require.NoError(t, err)
	require.Equal(t, []KioskUsagePoint{
		{Bucket: 3600, Calls: 17, Quota: 179, TokenUsed: 129},
		{Bucket: 7200, Calls: 4, Quota: 8, TokenUsed: 8},
	}, points)
}

func TestGetKioskTopDimensionOrdersByTokensThenCalls(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)

	models, err := GetKioskTopDimension(0, 10000, "model", 10)
	require.NoError(t, err)
	require.Equal(t, []KioskDimensionItem{
		{Name: "gpt-a", Calls: 7, Quota: 158, TokenUsed: 68},
		{Name: "gpt-b", Calls: 5, Quota: 20, TokenUsed: 60},
	}, models)

	groups, err := GetKioskTopDimension(0, 10000, "group", 10)
	require.NoError(t, err)
	require.Equal(t, []KioskDimensionItem{
		{Name: "free", Calls: 9, Quota: 28, TokenUsed: 68},
		{Name: "vip", Calls: 3, Quota: 150, TokenUsed: 60},
	}, groups)

	// bob and alice tie on tokens; the calls tie-breaker orders bob first.
	users, err := GetKioskTopDimension(0, 10000, "user", 2)
	require.NoError(t, err)
	require.Equal(t, []KioskDimensionItem{
		{Name: "bob", Calls: 5, Quota: 20, TokenUsed: 60},
		{Name: "alice", Calls: 3, Quota: 150, TokenUsed: 60},
	}, users)
}

func TestGetKioskTopDimensionRejectsUnknownColumn(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)

	for _, dimension := range []string{"", "quota", "model_name; DROP TABLE quota_data", "created_at"} {
		items, err := GetKioskTopDimension(0, 10000, dimension, 10)
		require.Error(t, err)
		require.Nil(t, items)
	}

	items, err := GetKioskTopDimension(0, 10000, "model", 0)
	require.NoError(t, err)
	require.Nil(t, items)
}

func TestGetKioskTopChannelsSkipsUnattributedRows(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)

	channels, err := GetKioskTopChannels(0, 10000, 10)
	require.NoError(t, err)
	require.Equal(t, []KioskDimensionItem{
		{ID: 1, Calls: 7, Quota: 158, TokenUsed: 68},
		{ID: 2, Calls: 5, Quota: 20, TokenUsed: 60},
	}, channels)

	names, err := GetKioskChannelNames([]int64{1, 2, 3})
	require.NoError(t, err)
	require.Equal(t, map[int64]string{1: "east", 2: "west"}, names)

	empty, err := GetKioskChannelNames(nil)
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestGetKioskRevenueCountsOnlyCompletedSuccessOrders(t *testing.T) {
	truncateTables(t)
	seedKioskTopUps(t)

	window, err := GetKioskRevenueAggregate(1000, 7199)
	require.NoError(t, err)
	require.Equal(t, &KioskRevenueAggregate{Amount: 30.75, OrderCount: 2}, window)

	// A zero start includes the epoch-completed order and everything later.
	total, err := GetKioskRevenueAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, &KioskRevenueAggregate{Amount: 132.75, OrderCount: 4}, total)

	pending, err := GetKioskPendingRevenueCount()
	require.NoError(t, err)
	require.Equal(t, int64(1), pending)
}

func TestGetKioskRevenueSeriesAndMethods(t *testing.T) {
	truncateTables(t)
	seedKioskTopUps(t)

	series, err := GetKioskRevenueSeries(1000, 7199, 3600)
	require.NoError(t, err)
	require.Equal(t, []KioskRevenuePoint{
		{Bucket: 3600, Amount: 30.75, OrderCount: 2},
	}, series)

	methods, err := GetKioskRevenueByMethod(1000, 7199)
	require.NoError(t, err)
	require.Equal(t, []KioskRevenueMethod{
		{Method: "stripe", Amount: 30.75, OrderCount: 2, Quota: 1500000},
	}, methods)
}

func TestGetKioskHealthAggregateAcrossBuckets(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&PerfMetric{ModelName: "gpt-a", Group: "vip", BucketTs: 3600, RequestCount: 10, SuccessCount: 9, TotalLatencyMs: 1000}).Error)
	require.NoError(t, DB.Create(&PerfMetric{ModelName: "gpt-b", Group: "free", BucketTs: 7200, RequestCount: 5, SuccessCount: 3, TotalLatencyMs: 500}).Error)
	require.NoError(t, DB.Create(&PerfMetric{ModelName: "gpt-a", Group: "vip", BucketTs: 20000, RequestCount: 100, SuccessCount: 1, TotalLatencyMs: 1}).Error)

	aggregate, err := GetKioskHealthAggregate(0, 10000)
	require.NoError(t, err)
	require.Equal(t, &KioskHealthAggregate{RequestCount: 15, SuccessCount: 12, TotalLatencyMs: 1500}, aggregate)
}

func TestGetKioskPlatformTotals(t *testing.T) {
	truncateTables(t)
	seedKioskQuotaData(t)
	require.NoError(t, DB.Create(&User{Id: 1, Username: "alice", AffCode: "aff1"}).Error)
	require.NoError(t, DB.Create(&User{Id: 2, Username: "bob", AffCode: "aff2"}).Error)
	require.NoError(t, DB.Create(&User{Id: 3, Username: "carol", AffCode: "aff3"}).Error)

	calls, quota, tokenUsed, totalUsers, err := GetKioskPlatformTotals()
	require.NoError(t, err)
	require.Equal(t, int64(21), calls)
	require.Equal(t, int64(187), quota)
	require.Equal(t, int64(137), tokenUsed)
	require.Equal(t, int64(3), totalUsers)
}

// TestKioskAggregatesOnExternalDatabases checks the kiosk SQL against a real
// MySQL/PostgreSQL schema: the modulo bucketing, alias ordering and GROUP BY
// shapes are what the in-memory SQLite suite cannot prove. Fixtures live in the
// 1970 epoch window, where a running platform writes nothing, so the queries see
// only these rows and cleanup deletes exactly the rows this test inserted.
func TestKioskAggregatesOnExternalDatabases(t *testing.T) {
	for _, tc := range []struct{ dialect, env string }{
		{dialect: "mysql", env: "TEST_MYSQL_DSN"},
		{dialect: "postgres", env: "TEST_POSTGRES_DSN"},
	} {
		t.Run(tc.dialect, func(t *testing.T) {
			dsn := os.Getenv(tc.env)
			if dsn == "" {
				t.Skipf("%s is not configured", tc.env)
			}

			var driver gorm.Dialector
			switch tc.dialect {
			case "mysql":
				driver = mysql.Open(dsn)
			case "postgres":
				driver = postgres.Open(dsn)
			}
			db, err := gorm.Open(driver, &gorm.Config{})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = sqlDB.Close() })

			var version string
			require.NoError(t, db.Raw("select version()").Scan(&version).Error)
			t.Logf("%s version: %s", tc.dialect, version)

			previous := DB
			DB = db
			t.Cleanup(func() { DB = previous })

			for _, row := range []QuotaData{
				{UserID: 1, Username: "alice", ModelName: "gpt-a", UseGroup: "vip", ChannelID: 1, CreatedAt: 3600, Count: 2, Quota: 100, TokenUsed: 40, CacheReadTokens: 5, CacheCreationTokens: 7},
				{UserID: 1, Username: "alice", ModelName: "gpt-a", UseGroup: "vip", ChannelID: 1, CreatedAt: 3601, Count: 1, Quota: 50, TokenUsed: 20, CacheReadTokens: 1, CacheCreationTokens: 2},
				{UserID: 2, Username: "bob", ModelName: "gpt-b", UseGroup: "free", ChannelID: 2, CreatedAt: 7199, Count: 5, Quota: 20, TokenUsed: 60},
				{UserID: 3, Username: "carol", ModelName: "gpt-a", UseGroup: "free", ChannelID: 1, CreatedAt: 7200, Count: 4, Quota: 8, TokenUsed: 8},
				{UserID: 4, Username: "dave", ModelName: "", UseGroup: "", ChannelID: 0, CreatedAt: 3700, Count: 9, Quota: 9, TokenUsed: 9},
			} {
				seeded := row
				require.NoError(t, db.Create(&seeded).Error)
				t.Cleanup(func() { db.Exec("DELETE FROM quota_data WHERE id = ?", seeded.Id) })
			}

			for _, topUp := range []TopUp{
				{UserId: 1, Amount: 500000, Money: 10.5, TradeNo: "t-kiosk-x-stripe", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CompleteTime: 3600},
				{UserId: 2, Amount: 1000000, Money: 20.25, TradeNo: "t-kiosk-x-stripe-2", PaymentMethod: "stripe", Status: common.TopUpStatusSuccess, CompleteTime: 7199},
				{UserId: 3, Amount: 250000, Money: 100, TradeNo: "t-kiosk-x-waffo", PaymentMethod: "waffo", Status: common.TopUpStatusSuccess, CompleteTime: 0},
				{UserId: 5, Amount: 50000, Money: 5, TradeNo: "t-kiosk-x-pending", PaymentMethod: "epay", Status: common.TopUpStatusPending, CompleteTime: 4000},
			} {
				seeded := topUp
				require.NoError(t, db.Create(&seeded).Error)
				t.Cleanup(func() { db.Exec("DELETE FROM top_ups WHERE id = ?", seeded.Id) })
			}

			for _, metric := range []PerfMetric{
				{ModelName: "gpt-a", Group: "vip", BucketTs: 3600, RequestCount: 10, SuccessCount: 9, TotalLatencyMs: 1000},
				{ModelName: "gpt-b", Group: "free", BucketTs: 7200, RequestCount: 5, SuccessCount: 3, TotalLatencyMs: 500},
				{ModelName: "gpt-a", Group: "vip", BucketTs: 20000, RequestCount: 100, SuccessCount: 1, TotalLatencyMs: 1},
			} {
				seeded := metric
				require.NoError(t, db.Create(&seeded).Error)
				t.Cleanup(func() { db.Exec("DELETE FROM perf_metrics WHERE id = ?", seeded.Id) })
			}

			window, err := GetKioskUsageAggregate(3600, 7199)
			require.NoError(t, err)
			require.Equal(t, &KioskUsageAggregate{
				Calls: 17, Quota: 179, TokenUsed: 129,
				CacheReadTokens: 6, CacheCreationTokens: 9, ActiveUsers: 3,
			}, window)

			series, err := GetKioskUsageSeries(0, 10000, 3600)
			require.NoError(t, err)
			require.Equal(t, []KioskUsagePoint{
				{Bucket: 3600, Calls: 17, Quota: 179, TokenUsed: 129},
				{Bucket: 7200, Calls: 4, Quota: 8, TokenUsed: 8},
			}, series)

			daySeries, err := GetKioskUsageSeries(0, 100000, 86400)
			require.NoError(t, err)
			require.Equal(t, []KioskUsagePoint{
				{Bucket: 0, Calls: 21, Quota: 187, TokenUsed: 137},
			}, daySeries)

			for _, dimension := range []struct {
				name     string
				expected []KioskDimensionItem
			}{
				{name: "model", expected: []KioskDimensionItem{
					{Name: "gpt-a", Calls: 7, Quota: 158, TokenUsed: 68},
					{Name: "gpt-b", Calls: 5, Quota: 20, TokenUsed: 60},
				}},
				{name: "group", expected: []KioskDimensionItem{
					{Name: "free", Calls: 9, Quota: 28, TokenUsed: 68},
					{Name: "vip", Calls: 3, Quota: 150, TokenUsed: 60},
				}},
				{name: "user", expected: []KioskDimensionItem{
					{Name: "bob", Calls: 5, Quota: 20, TokenUsed: 60},
					{Name: "alice", Calls: 3, Quota: 150, TokenUsed: 60},
					{Name: "dave", Calls: 9, Quota: 9, TokenUsed: 9},
					{Name: "carol", Calls: 4, Quota: 8, TokenUsed: 8},
				}},
			} {
				items, err := GetKioskTopDimension(0, 10000, dimension.name, 10)
				require.NoError(t, err)
				require.Equal(t, dimension.expected, items, dimension.name)
			}

			channels, err := GetKioskTopChannels(0, 10000, 10)
			require.NoError(t, err)
			require.Equal(t, []KioskDimensionItem{
				{ID: 1, Calls: 7, Quota: 158, TokenUsed: 68},
				{ID: 2, Calls: 5, Quota: 20, TokenUsed: 60},
			}, channels)

			// The id list path must work even when no channel row matches, so the
			// kiosk can fall back to the raw id instead of failing the overview.
			names, err := GetKioskChannelNames([]int64{-1})
			require.NoError(t, err)
			require.Empty(t, names)

			revenue, err := GetKioskRevenueAggregate(1000, 7199)
			require.NoError(t, err)
			require.Equal(t, &KioskRevenueAggregate{Amount: 30.75, OrderCount: 2}, revenue)

			revenueSeries, err := GetKioskRevenueSeries(0, 10000, 3600)
			require.NoError(t, err)
			require.Equal(t, []KioskRevenuePoint{
				{Bucket: 0, Amount: 100, OrderCount: 1},
				{Bucket: 3600, Amount: 30.75, OrderCount: 2},
			}, revenueSeries)

			methods, err := GetKioskRevenueByMethod(1000, 7199)
			require.NoError(t, err)
			require.Equal(t, []KioskRevenueMethod{
				{Method: "stripe", Amount: 30.75, OrderCount: 2, Quota: 1500000},
			}, methods)

			health, err := GetKioskHealthAggregate(0, 10000)
			require.NoError(t, err)
			require.Equal(t, &KioskHealthAggregate{RequestCount: 15, SuccessCount: 12, TotalLatencyMs: 1500}, health)
		})
	}
}
