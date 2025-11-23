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

import { forwardRef } from 'react'
import type { HTMLAttributes } from 'react'
import { cn } from '@shared/utils/cn'

export type BadgeVariant =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'created'
  | 'processing'
  | 'verified'
  | 'shipped'
  | 'delivered'
  | 'failed'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  size?: 'sm' | 'md' | 'lg'
}

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant = 'default', size = 'md', children, ...props }, ref) => {
    const baseStyles =
      'inline-flex items-center font-medium rounded-full transition-all backdrop-blur px-3 py-1 text-xs'

    const variants = {
      default: 'bg-white/10 text-white border border-white/20',
      primary: 'bg-white text-black',
      success: 'bg-emerald-400/20 text-emerald-200 border border-emerald-300/40',
      warning: 'bg-amber-400/20 text-amber-100 border border-amber-300/40',
      danger: 'bg-rose-500/20 text-rose-200 border border-rose-400/40',
      info: 'bg-cyan-400/20 text-cyan-100 border border-cyan-300/40',
      // Blockchain-specific status colors
      created: 'bg-white/15 text-white border border-white/30',
      processing: 'bg-amber-400/15 text-amber-100 border border-amber-300/30',
      verified: 'bg-emerald-400/15 text-emerald-100 border border-emerald-300/30',
      shipped: 'bg-purple-400/15 text-purple-100 border border-purple-300/30',
      delivered: 'bg-sky-400/15 text-sky-100 border border-sky-300/30',
      failed: 'bg-rose-500/15 text-rose-100 border border-rose-400/30',
    }

    const sizes = {
      sm: 'text-[11px] px-2 py-0.5',
      md: 'text-sm px-3 py-1',
      lg: 'text-base px-4 py-1.5',
    }

    return (
      <span
        ref={ref}
        className={cn(baseStyles, variants[variant], sizes[size], className)}
        {...props}
      >
        {children}
      </span>
    )
  }
)

Badge.displayName = 'Badge'

