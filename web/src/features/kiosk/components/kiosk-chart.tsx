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
import { VChart } from '@visactor/react-vchart'
import { useMemo, type ReactNode } from 'react'

import { PanelWrapper } from '@/features/dashboard/components/ui/panel-wrapper'
import { useChartTheme } from '@/lib/use-chart-theme'
import { cn } from '@/lib/utils'
import { VCHART_OPTION } from '@/lib/vchart'

import type { KioskChartRecord } from '../lib/series'

// The kiosk shell is always dark, so the axis ink is fixed to the dark values
// the rest of the app derives from `resolvedTheme`.
const CHART_TEXT_COLOR = 'rgba(255, 255, 255, 0.68)'
const CHART_GRID_COLOR = 'rgba(255, 255, 255, 0.12)'

type KioskChartVariant = 'area' | 'bar'

interface KioskChartProps {
  title: string
  description?: string
  records: KioskChartRecord[]
  formatValue: (value: number) => string
  tooltipKey: string
  variant?: KioskChartVariant
  emptyMessage?: string
  loading?: boolean
  className?: string
  headerActions?: ReactNode
}

export function KioskChart(props: KioskChartProps) {
  const { themeReady } = useChartTheme()
  const { formatValue, records, tooltipKey } = props
  const variant = props.variant ?? 'area'
  const hasData = records.length > 0

  const spec = useMemo(() => {
    if (!hasData) return null

    const seriesStyles =
      variant === 'bar'
        ? { bar: { style: { cornerRadius: 3 } }, paddingInner: 0.35 }
        : {
            line: { style: { lineWidth: 2, curveType: 'monotone' } },
            area: { style: { fillOpacity: 0.16, curveType: 'monotone' } },
            point: { visible: false },
          }

    return {
      type: variant,
      data: [{ id: 'kiosk-series', values: records }],
      xField: 'label',
      yField: 'value',
      legends: { visible: false },
      ...seriesStyles,
      axes: [
        {
          orient: 'bottom',
          label: {
            style: { fill: CHART_TEXT_COLOR, fontSize: 11 },
            autoHide: true,
            autoLimit: true,
          },
          tick: { visible: false },
        },
        {
          orient: 'left',
          min: 0,
          label: {
            style: { fill: CHART_TEXT_COLOR, fontSize: 11 },
            formatMethod: (value: number | string) =>
              formatValue(Number(value)),
          },
          grid: {
            visible: true,
            style: { lineDash: [3, 3], stroke: CHART_GRID_COLOR },
          },
        },
      ],
      tooltip: {
        mark: {
          title: {
            value: (datum: Record<string, unknown>) =>
              String(datum?.label ?? ''),
          },
          content: [
            {
              key: () => tooltipKey,
              value: (datum: Record<string, unknown>) =>
                formatValue(Number(datum?.value) || 0),
            },
          ],
        },
      },
      animationAppear: { duration: 400 },
    }
  }, [formatValue, hasData, records, tooltipKey, variant])

  return (
    <PanelWrapper
      title={props.title}
      description={props.description}
      loading={props.loading}
      empty={!props.loading && !hasData}
      emptyMessage={props.emptyMessage}
      height='h-full'
      className={cn('flex min-h-0 flex-col', props.className)}
      contentClassName='min-h-0 flex-1'
      headerActions={props.headerActions}
    >
      {spec && themeReady ? (
        <div className='h-full min-h-[12rem] w-full'>
          <VChart
            key={`kiosk-chart-${variant}`}
            spec={{ ...spec, theme: 'dark', background: 'transparent' }}
            option={VCHART_OPTION}
          />
        </div>
      ) : null}
    </PanelWrapper>
  )
}
