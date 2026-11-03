import { Link } from '@tanstack/react-router'
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
import { Maximize, Minimize, Monitor } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { formatDateStr, formatTimeStr } from '@/lib/format'
import { cn } from '@/lib/utils'

import {
  KIOSK_REFRESH_VALUES,
  KIOSK_WINDOW_VALUES,
  type KioskRefreshValue,
} from '../constants'
import type { KioskWindow } from '../types'

type KioskHeaderProps = {
  window: KioskWindow
  onWindowChange: (value: KioskWindow) => void
  refreshSeconds: KioskRefreshValue
  onRefreshChange: (value: KioskRefreshValue) => void
  secondsLeft: number
  generatedAt: number
  isStale: boolean
  isFetching: boolean
  fullscreenSupported: boolean
  isFullscreen: boolean
  onToggleFullscreen: () => void
  className?: string
}

export function KioskHeader(props: KioskHeaderProps) {
  const { t } = useTranslation()
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 1000)
    return () => window.clearInterval(timer)
  }, [])

  const refreshItems = KIOSK_REFRESH_VALUES.map((seconds) => ({
    value: String(seconds),
    label: seconds === 0 ? t('Off') : `${seconds}s`,
  }))

  return (
    <header className='bg-card flex flex-wrap items-center gap-x-6 gap-y-3 rounded-2xl border px-4 py-3 sm:px-5'>
      <div className='flex items-center gap-3'>
        <Monitor className='text-primary size-5' />
        <div>
          <h1 className='text-foreground text-base leading-tight font-semibold'>
            {t('Data Kiosk')}
          </h1>
          <p className='text-muted-foreground text-xs'>
            {t('Platform operations overview')}
            {props.generatedAt > 0 &&
              ` · ${t('Updated {{time}}', {
                time: formatTimeStr(new Date(props.generatedAt * 1000)),
              })}`}
          </p>
        </div>
      </div>

      <div className='ms-auto text-right'>
        <div className='text-foreground font-mono text-2xl leading-none font-semibold tabular-nums'>
          {formatTimeStr(now)}
        </div>
        <div className='text-muted-foreground mt-1 font-mono text-xs tabular-nums'>
          {formatDateStr(now)}
        </div>
      </div>

      <div className='flex flex-wrap items-center gap-2 sm:gap-3'>
        <span
          className={cn(
            'flex items-center gap-1.5 text-xs',
            props.isStale ? 'text-warning' : 'text-success'
          )}
        >
          <span
            className={cn(
              'size-1.5 rounded-full',
              props.isStale ? 'bg-warning' : 'bg-success',
              props.isFetching && 'animate-pulse'
            )}
          />
          {props.isStale ? t('Stale data') : t('Live')}
        </span>

        {props.refreshSeconds > 0 && (
          <span className='text-muted-foreground font-mono text-xs tabular-nums'>
            {t('Refresh in {{seconds}}s', { seconds: props.secondsLeft })}
          </span>
        )}

        <ToggleGroup
          value={[props.window]}
          onValueChange={(values) => {
            const next = values[0] as KioskWindow | undefined
            if (next) props.onWindowChange(next)
          }}
          variant='outline'
          size='sm'
          aria-label={t('Time window')}
        >
          {KIOSK_WINDOW_VALUES.map((item) => (
            <ToggleGroupItem key={item} value={item}>
              {item}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>

        <ToggleGroup
          value={[String(props.refreshSeconds)]}
          onValueChange={(values) => {
            const next = KIOSK_REFRESH_VALUES.find(
              (seconds) => String(seconds) === values[0]
            )
            if (next != null) props.onRefreshChange(next)
          }}
          variant='outline'
          size='sm'
          aria-label={t('Auto refresh')}
        >
          {refreshItems.map((item) => (
            <ToggleGroupItem key={item.value} value={item.value}>
              {item.label}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>

        <Button
          variant='outline'
          size='icon'
          render={<Link to='/' />}
          aria-label={t('Back to console')}
        >
          <Monitor className='size-4' />
        </Button>

        {props.fullscreenSupported && (
          <Button
            variant='outline'
            size='icon'
            onClick={props.onToggleFullscreen}
            aria-label={
              props.isFullscreen ? t('Exit fullscreen') : t('Enter fullscreen')
            }
          >
            {props.isFullscreen ? (
              <Minimize className='size-4' />
            ) : (
              <Maximize className='size-4' />
            )}
          </Button>
        )}
      </div>
    </header>
  )
}
