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

