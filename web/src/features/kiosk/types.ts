/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
export type KioskWindow = '24h' | '7d' | '30d'

export interface KioskUsageAggregate {
  calls: number
  quota: number
  token_used: number
  cache_read_tokens: number
  cache_creation_tokens: number
  active_users: number
}

export interface KioskUsagePoint {
  bucket: number
  calls: number
  quota: number
  token_used: number
}

export interface KioskDimensionItem {
  id: number
  name: string
  calls: number
  quota: number
  token_used: number
}

export interface KioskRevenuePoint {
  bucket: number
  amount: number
  order_count: number
}

export interface KioskRevenueMethod {
  method: string
  amount: number
  order_count: number
  quota: number
}

export interface KioskRevenue {
  today_amount: number
  window_amount: number
  total_amount: number
  success_count: number
  pending_count: number
  series: KioskRevenuePoint[]
  by_method: KioskRevenueMethod[]
}

export interface KioskHealth {
  request_count: number
  success_count: number
  success_rate: number
  avg_latency_ms: number
}

export interface KioskPlatformTotals {
  calls: number
  quota: number
  token_used: number
  total_users: number
}

export interface KioskOverview {
  generated_at: number
  window: KioskWindow
  bucket_seconds: number
  data_export_enabled: boolean
  today: KioskUsageAggregate
  current: KioskUsageAggregate
  previous: KioskUsageAggregate
  series: KioskUsagePoint[]
  top_models: KioskDimensionItem[]
  top_channels: KioskDimensionItem[]
  top_groups: KioskDimensionItem[]
  top_users: KioskDimensionItem[]
  revenue: KioskRevenue
  health: KioskHealth
  totals: KioskPlatformTotals
}
