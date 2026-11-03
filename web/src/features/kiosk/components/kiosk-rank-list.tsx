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
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { PanelWrapper } from '@/features/dashboard/components/ui/panel-wrapper'
import { formatNumber } from '@/lib/format'
import { cn } from '@/lib/utils'

import type { KioskDimensionItem } from '../types'

interface KioskRankListProps {
  title: string
  description?: string
  items: KioskDimensionItem[]
  metric: 'calls' | 'token_used'
  formatValue: (value: number) => string
  loading?: boolean
  className?: string
  headerActions?: ReactNode
}

export function KioskRankList(props: KioskRankListProps) {
  const { t } = useTranslation()
  const hasData = props.items.length > 0
  const max = Math.max(...props.items.map((item) => item[props.metric]), 0)

  return (
    <PanelWrapper
      title={props.title}
      description={props.description}
      loading={props.loading}
      empty={!props.loading && !hasData}
      emptyMessage={t('No data available')}
      height='h-full'
      className={cn('flex min-h-0 flex-col', props.className)}
      contentClassName='min-h-0 flex-1 overflow-hidden'
      headerActions={props.headerActions}
    >
      <ol className='divide-border/60 divide-y'>
        {props.items.map((item, index) => {
          const value = item[props.metric]
          const share = max > 0 ? (value / max) * 100 : 0
          return (
            <li
              key={`${item.id}-${item.name}`}
              className='grid grid-cols-[1.75rem_minmax(0,1fr)_auto] items-center gap-x-3 py-2'
            >
              <span className='text-muted-foreground/80 text-right font-mono text-xs tabular-nums'>
                {index + 1}
              </span>
              <div className='min-w-0'>
                <div className='text-foreground truncate text-sm font-medium'>
                  {item.name}
                </div>
                <div className='bg-muted mt-1.5 h-1 overflow-hidden rounded-full'>
                  <div
                    className='bg-primary h-full rounded-full'
                    style={{ width: `${Math.max(share, 2)}%` }}
                  />
                </div>
              </div>
              <div className='shrink-0 text-right'>
                <div className='text-foreground font-mono text-sm font-semibold tabular-nums'>
                  {props.formatValue(value)}
                </div>
                {props.metric !== 'calls' && (
                  <div className='text-muted-foreground/80 text-[11px] tabular-nums'>
                    {formatNumber(item.calls)} {t('calls')}
                  </div>
                )}
              </div>
            </li>
          )
        })}
      </ol>
    </PanelWrapper>
  )
}
