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
  Activity,
  CircleDollarSign,
  Coins,
  Gauge,
  TrendingUp,
  Users,
} from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { StatCard } from '@/features/dashboard/components/ui/stat-card'
import { GrowthText } from '@/features/rankings/components/growth-text'
import {
  formatLocalCurrencyAmount,
  formatQuotaWithCurrency,
} from '@/lib/currency'
import { formatCompactNumber, formatNumber, formatPercent } from '@/lib/format'

import { KIOSK_SPARKLINE_BUCKETS, KIOSK_WINDOW_LABELS } from '../constants'
import { growthRatio } from '../lib/series'
import type { KioskOverview, KioskWindow } from '../types'

type KioskMetricRowProps = {
  overview: KioskOverview
  window: KioskWindow
}

function renderGrowth(value: number, base: number): ReactNode {
  const ratio = growthRatio(value, base)
  return ratio == null ? undefined : <GrowthText value={ratio} />
}

export function KioskMetricRow(props: KioskMetricRowProps) {
  const { t } = useTranslation()
  const { current, previous, today, series, revenue, health, totals } =
    props.overview
  const windowLabel = t(KIOSK_WINDOW_LABELS[props.window])

  const callSparkline = series
    .slice(-KIOSK_SPARKLINE_BUCKETS)
    .map((p) => p.calls)
  const tokenSparkline = series
    .slice(-KIOSK_SPARKLINE_BUCKETS)
    .map((p) => p.token_used)

  return (
    <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-6'>
      <StatCard
        title={t('API Calls')}
        value={formatNumber(current.calls)}
        description={t('{{count}} today', {
          count: formatCompactNumber(today.calls),
        })}
        icon={Activity}
        tone='accent-1'
        sparkline={callSparkline}
        sparklineVariant='line'
        action={renderGrowth(current.calls, previous.calls)}
      />
      <StatCard
        title={t('Tokens')}
        value={formatCompactNumber(current.token_used)}
        description={t('{{count}} today', {
          count: formatCompactNumber(today.token_used),
        })}
        icon={Coins}
        tone='accent-2'
        sparkline={tokenSparkline}
      />
      <StatCard
        title={t('Consumption')}
        value={formatQuotaWithCurrency(current.quota)}
        description={t('Platform spend in {{window}}', {
          window: windowLabel,
        })}
        icon={TrendingUp}
        tone='accent-3'
        details={[
          { label: t('Today'), value: formatQuotaWithCurrency(today.quota) },
          {
            label: t('All time'),
            value: formatQuotaWithCurrency(totals.quota),
          },
        ]}
      />
      <StatCard
        title={t('Recharged')}
        value={formatLocalCurrencyAmount(revenue.window_amount)}
        description={t('Successful orders in {{window}}', {
          window: windowLabel,
        })}
        icon={CircleDollarSign}
        tone='accent-1'
        details={[
          {
            label: t('Today'),
            value: formatLocalCurrencyAmount(revenue.today_amount),
          },
          {
            label: t('Pending'),
            value: formatNumber(revenue.pending_count),
            tone: revenue.pending_count > 0 ? 'warning' : 'muted',
          },
        ]}
      />
      <StatCard
        title={t('Active Users')}
        value={formatNumber(current.active_users)}
        description={t('Total users {{count}}', {
          count: formatNumber(totals.total_users),
        })}
        icon={Users}
        tone='accent-2'
        details={[
          { label: t('Today'), value: formatNumber(today.active_users) },
          {
            label: t('All time calls'),
            value: formatCompactNumber(totals.calls),
          },
        ]}
      />
      <StatCard
        title={t('Success Rate')}
        value={formatPercent(health.success_rate)}
        description={t('{{count}} requests sampled', {
          count: formatNumber(health.request_count),
        })}
        icon={Gauge}
        tone='accent-3'
        details={[
          { label: t('Average latency'), value: `${health.avg_latency_ms} ms` },
          {
            label: t('Failures'),
            value: formatNumber(health.request_count - health.success_count),
            tone:
              health.request_count - health.success_count > 0
                ? 'destructive'
                : 'success',
          },
        ]}
      />
    </div>
  )
}
