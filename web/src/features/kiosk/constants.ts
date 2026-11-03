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
import type { KioskWindow } from './types'

export const KIOSK_DAY_BUCKET_SECONDS = 86400

export const KIOSK_WINDOW_VALUES = ['24h', '7d', '30d'] as const

export const KIOSK_DEFAULT_WINDOW: KioskWindow = '7d'

export const KIOSK_REFRESH_VALUES = [0, 30, 60, 300] as const

export type KioskRefreshValue = (typeof KIOSK_REFRESH_VALUES)[number]

export const KIOSK_DEFAULT_REFRESH_SECONDS: KioskRefreshValue = 60

export const KIOSK_RANK_LIMIT = 8

export const KIOSK_SPARKLINE_BUCKETS = 24

export const KIOSK_WINDOW_LABELS: Record<KioskWindow, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
}
