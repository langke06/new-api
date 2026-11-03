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
import {
  createFileRoute,
  redirect,
  useNavigate,
  useSearch,
} from '@tanstack/react-router'
import z from 'zod'

import { Kiosk } from '@/features/kiosk'
import {
  KIOSK_DEFAULT_REFRESH_SECONDS,
  KIOSK_DEFAULT_WINDOW,
  KIOSK_WINDOW_VALUES,
} from '@/features/kiosk/constants'
import { resolveAuthentication } from '@/lib/auth-session'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

const kioskSearchSchema = z.object({
  window: z.enum(KIOSK_WINDOW_VALUES).catch(KIOSK_DEFAULT_WINDOW),
  refresh: z
    .union([z.literal(0), z.literal(30), z.literal(60), z.literal(300)])
    .catch(KIOSK_DEFAULT_REFRESH_SECONDS),
})

export const Route = createFileRoute('/kiosk/')({
  beforeLoad: async ({ location }) => {
    await resolveAuthentication()

    const { auth } = useAuthStore.getState()

    if (!auth.user) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
      })
    }
    if (auth.user.role < ROLE.ADMIN) {
      throw redirect({ to: '/403' })
    }
  },
  validateSearch: kioskSearchSchema,
  component: KioskPage,
})

function KioskPage() {
  const search = useSearch({ from: '/kiosk/' })
  const navigate = useNavigate()

  return (
    <Kiosk
      window={search.window}
      refreshSeconds={search.refresh}
      onWindowChange={(value) =>
        navigate({
          to: '/kiosk',
          search: { window: value, refresh: search.refresh },
        })
      }
      onRefreshChange={(value) =>
        navigate({
          to: '/kiosk',
          search: { window: search.window, refresh: value },
        })
      }
    />
  )
}
