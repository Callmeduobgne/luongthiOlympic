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

import type { ReactNode } from 'react'
import { Package, Search, Inbox, CheckCircle } from 'lucide-react'
import { Button } from '@shared/components/ui/Button'
import { cn } from '@shared/utils/cn'

interface EmptyStateProps {
  icon?: 'package' | 'search' | 'inbox' | 'check-circle' | ReactNode
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  className?: string
}

const iconMap = {
  package: Package,
  search: Search,
  inbox: Inbox,
  'check-circle': CheckCircle,
}

export const EmptyState = ({
  icon = 'package',
  title,
  description,
  action,
  className,
}: EmptyStateProps) => {
  const IconComponent =
    typeof icon === 'string' ? (iconMap[icon as keyof typeof iconMap]) : undefined
  const CustomIcon = typeof icon !== 'string' ? icon : undefined

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-12 px-4 text-center rounded-3xl border border-white/10 bg-white/5 backdrop-blur-2xl text-white shadow-[0_20px_45px_rgba(0,0,0,0.45)]',
        className
      )}
    >
      <div className="mb-4">
        {IconComponent && (
          <IconComponent className="h-12 w-12 text-white/50" />
        )}
        {CustomIcon && <div className="h-12 w-12">{CustomIcon}</div>}
      </div>
      <h3 className="text-lg font-semibold mb-2">
        {title}
      </h3>
      {description && (
        <p className="text-sm text-gray-300 mb-6 max-w-sm">
          {description}
        </p>
      )}
      {action && (
        <Button onClick={action.onClick} variant="primary" size="md">
          {action.label}
        </Button>
      )}
    </div>
  )
}

