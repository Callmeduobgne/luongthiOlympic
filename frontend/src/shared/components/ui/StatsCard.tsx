/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Card } from './Card'
import { cn } from '@shared/utils/cn'

interface StatsCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon?: React.ReactNode
  trend?: {
    value: number
    isPositive: boolean
  }
  className?: string
}

export const StatsCard = ({
  title,
  value,
  subtitle,
  icon,
  trend,
  className,
}: StatsCardProps) => {
  return (
    <Card
      variant="outlined"
      className={cn(
        'bg-white/5 border-white/10 backdrop-blur-2xl text-white shadow-[0_18px_35px_rgba(0,0,0,0.45)] hover:border-white/25 transition-all',
        className
      )}
    >
      <div className="p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-300 uppercase tracking-widest">
            {title}
          </h3>
          {icon && (
            <div className="p-2 rounded-full bg-white/10 border border-white/15 text-white">
              {icon}
            </div>
          )}
        </div>
        <div className="flex items-baseline gap-2">
          <p className="text-3xl font-bold text-white">{value}</p>
          {trend && (
            <span
              className={cn(
                'text-sm font-medium px-2 py-0.5 rounded-full',
                trend.isPositive ? 'text-emerald-300 bg-emerald-300/10' : 'text-rose-300 bg-rose-300/10'
              )}
            >
              {trend.isPositive ? '+' : ''}
              {trend.value}%
            </span>
          )}
        </div>
        {subtitle && (
          <p className="text-sm text-gray-400">{subtitle}</p>
        )}
      </div>
    </Card>
  )
}

