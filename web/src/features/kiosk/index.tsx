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
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { useCountdown } from '@/hooks/use-countdown'
import { useFullscreen } from '@/hooks/use-fullscreen'
import { formatLocalCurrencyAmount } from '@/lib/currency'
import { formatCompactNumber, formatNumber } from '@/lib/format'

import { KioskChart } from './components/kiosk-chart'
import { KioskHeader } from './components/kiosk-header'
import { KioskMetricRow } from './components/kiosk-metric-row'
import { KioskRankList } from './components/kiosk-rank-list'
import {
  KIOSK_DEFAULT_REFRESH_SECONDS,
  KIOSK_RANK_LIMIT,
  KIOSK_WINDOW_LABELS,
  type KioskRefreshValue,
} from './constants'
import { useKioskOverview } from './hooks/use-kiosk-overview'
import { buildRevenueRecords, buildUsageRecords } from './lib/series'
import type { KioskWindow } from './types'

type UsageMetric = 'calls' | 'token_used'

type KioskProps = {
  window: KioskWindow
  refreshSeconds: KioskRefreshValue
  onWindowChange: (value: KioskWindow) => void
  onRefreshChange: (value: KioskRefreshValue) => void
}

export function Kiosk(props: KioskProps) {
  const { t } = useTranslation()
  const [usageMetric, setUsageMetric] = useState<UsageMetric>('calls')
  const query = useKioskOverview(props.window, props.refreshSeconds)
  const overview = query.data
  const {
    isFullscreen,
    supported: fullscreenSupported,
    toggle: toggleFullscreen,
  } = useFullscreen()

  const { start, stop, secondsLeft } = useCountdown({
    initialSeconds: props.refreshSeconds || KIOSK_DEFAULT_REFRESH_SECONDS,
  })
  const dataUpdatedAt = query.dataUpdatedAt

  useEffect(() => {
    if (props.refreshSeconds > 0 && dataUpdatedAt > 0) {
      start(props.refreshSeconds)
      return
    }
    stop()
  }, [dataUpdatedAt, props.refreshSeconds, start, stop])

  const usageRecords = useMemo(
    () =>
      overview
        ? buildUsageRecords(
            overview.series,
            overview.bucket_seconds,
            usageMetric
          )
        : [],
    [overview, usageMetric]
  )
  const revenueRecords = useMemo(
    () =>
      overview
        ? buildRevenueRecords(overview.revenue.series, overview.bucket_seconds)
        : [],
    [overview]
  )

  const headerProps = {
    ...props,
    secondsLeft: overview ? secondsLeft : 0,
    isFullscreen,
    fullscreenSupported,
    onToggleFullscreen: toggleFullscreen,
    isStale: overview == null || query.isError || query.isPaused,
    isFetching: query.isFetching,
    generatedAt: overview?.generated_at ?? 0,
  }

  if (!overview) {
    return (
      <KioskShell>
        <KioskHeader {...headerProps} />
        <div className='bg-card flex flex-1 flex-col items-center justify-center gap-3 rounded-2xl border text-center'>
          {query.isError ? (
            <>
              <p className='text-muted-foreground text-sm'>
                {t('Unable to load the kiosk data')}
              </p>
              <Button
                variant='outline'
                size='sm'
                onClick={() => query.refetch()}
                className='gap-1.5'
              >
                <RefreshCw className='size-4' />
                {t('Retry')}
              </Button>
            </>
          ) : (
            <p className='text-muted-foreground text-sm'>{t('Loading...')}</p>
          )}
        </div>
      </KioskShell>
    )
  }

  const resolvedWindowLabel = t(KIOSK_WINDOW_LABELS[props.window])

  return (
    <KioskShell>
      <KioskHeader {...headerProps} />

      {!overview.data_export_enabled && (
        <Alert variant='default'>
          <AlertTriangle />
          <AlertTitle>{t('Usage recording is disabled')}</AlertTitle>
          <AlertDescription>
            {t(
              'Call and token figures stay empty while the server option Data Export is off.'
            )}
          </AlertDescription>
        </Alert>
      )}

      {query.isError && (
        <Alert variant='destructive'>
          <AlertTriangle />
          <AlertTitle>{t('Showing the last loaded figures')}</AlertTitle>
          <AlertDescription>{query.error.message}</AlertDescription>
        </Alert>
      )}

      <KioskMetricRow overview={overview} window={props.window} />

      <div className='grid min-h-0 flex-1 gap-3 xl:grid-cols-12 xl:grid-rows-[minmax(0,3fr)_minmax(0,2fr)]'>
        <KioskChart
          className='min-h-[18rem] xl:col-span-7'
          title={t('Call Volume Trend')}
          description={resolvedWindowLabel}
          records={usageRecords}
          loading={query.isFetching}
          tooltipKey={usageMetric === 'calls' ? t('Calls') : t('Tokens')}
          formatValue={
            usageMetric === 'calls'
              ? (value) => formatNumber(value)
              : (value) => formatCompactNumber(value)
          }
          emptyMessage={t('No usage recorded in this window')}
          headerActions={
            <ToggleGroup
              value={[usageMetric]}
              onValueChange={(values) => {
                const next = values[0] as UsageMetric | undefined
                if (next) setUsageMetric(next)
              }}
              variant='outline'
              size='sm'
              aria-label={t('Metric')}
            >
              <ToggleGroupItem value='calls'>{t('Calls')}</ToggleGroupItem>
              <ToggleGroupItem value='token_used'>
                {t('Tokens')}
              </ToggleGroupItem>
            </ToggleGroup>
          }
        />

        <KioskChart
          className='min-h-[18rem] xl:col-span-5'
          variant='bar'
          title={t('Recharge Trend')}
          description={t('Paid amount of completed orders')}
          records={revenueRecords}
          loading={query.isFetching}
          tooltipKey={t('Amount')}
          formatValue={(value) => formatLocalCurrencyAmount(value)}
          emptyMessage={t('No orders completed in this window')}
        />

        <KioskRankList
          className='min-h-[14rem] xl:col-span-3'
          title={t('Top Models')}
          description={t('Ranked by tokens')}
          items={overview.top_models.slice(0, KIOSK_RANK_LIMIT)}
          metric='token_used'
          formatValue={(value) => formatCompactNumber(value)}
        />
        <KioskRankList
          className='min-h-[14rem] xl:col-span-3'
          title={t('Top Channels')}
          description={t('Ranked by tokens')}
          items={overview.top_channels.slice(0, KIOSK_RANK_LIMIT)}
          metric='token_used'
          formatValue={(value) => formatCompactNumber(value)}
        />
        <KioskRankList
          className='min-h-[14rem] xl:col-span-3'
          title={t('Top Groups')}
          description={t('Ranked by tokens')}
          items={overview.top_groups.slice(0, KIOSK_RANK_LIMIT)}
          metric='token_used'
          formatValue={(value) => formatCompactNumber(value)}
        />
        <KioskRankList
          className='min-h-[14rem] xl:col-span-3'
          title={t('Top Users')}
          description={t('Ranked by calls')}
          items={overview.top_users.slice(0, KIOSK_RANK_LIMIT)}
          metric='calls'
          formatValue={(value) => formatNumber(value)}
        />
      </div>
    </KioskShell>
  )
}

function KioskShell(props: { children: ReactNode }) {
  return (
    <div className='dark bg-background text-foreground min-h-dvh p-3 sm:p-4 lg:h-dvh lg:overflow-hidden'>
      <div className='flex h-full flex-col gap-3'>{props.children}</div>
    </div>
  )
}
