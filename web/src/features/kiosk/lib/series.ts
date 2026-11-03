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
import dayjs from 'dayjs'

import { KIOSK_DAY_BUCKET_SECONDS } from '../constants'
import type { KioskRevenuePoint, KioskUsagePoint } from '../types'

export interface KioskChartRecord {
  label: string
  value: number
}

export function formatBucketLabel(
  bucket: number,
  bucketSeconds: number
): string {
  return dayjs
    .unix(bucket)
    .format(bucketSeconds >= KIOSK_DAY_BUCKET_SECONDS ? 'MM-DD' : 'HH:mm')
}

export function buildUsageRecords(
  points: KioskUsagePoint[],
  bucketSeconds: number,
  metric: 'calls' | 'token_used'
): KioskChartRecord[] {
  return points.map((point) => ({
    label: formatBucketLabel(point.bucket, bucketSeconds),
    value: point[metric],
  }))
}

export function buildRevenueRecords(
  points: KioskRevenuePoint[],
  bucketSeconds: number
): KioskChartRecord[] {
  return points.map((point) => ({
    label: formatBucketLabel(point.bucket, bucketSeconds),
    value: point.amount,
  }))
}

export function growthRatio(current: number, previous: number): number | null {
  if (previous <= 0) return null
  return ((current - previous) / previous) * 100
}
