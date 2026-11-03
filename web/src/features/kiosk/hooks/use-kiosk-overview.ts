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
import { keepPreviousData, useQuery } from '@tanstack/react-query'

import { requireServerSuccess } from '@/lib/server-error-message'

import { getKioskOverview } from '../api'
import type { KioskWindow } from '../types'

// A wall-mounted screen keeps showing the last good frame while the projector
// host is offline, so polling failures must stay silent and non-destructive.
export function useKioskOverview(window: KioskWindow, refreshSeconds: number) {
  return useQuery({
    queryKey: ['kiosk', 'overview', window],
    queryFn: async () =>
      requireServerSuccess(await getKioskOverview(window)).data,
    placeholderData: keepPreviousData,
    refetchInterval: refreshSeconds > 0 ? refreshSeconds * 1000 : false,
    refetchIntervalInBackground: false,
    meta: { errorToast: false },
  })
}
