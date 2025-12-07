import { forwardRef } from 'react'
import type { HTMLAttributes } from 'react'
import { cn } from '@shared/utils/cn'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'outlined'
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'default', children, ...props }, ref) => {
    const baseStyles = 'rounded-2xl transition-all duration-300 backdrop-blur-2xl text-white'
    const variants = {
      default:
        'bg-white/5 border border-white/10 shadow-[0_18px_45px_rgba(0,0,0,0.45)] hover:border-white/25',
      outlined:
        'border border-white/15 bg-black/30 shadow-[0_12px_35px_rgba(0,0,0,0.4)]',
    }

    return (
      <div
        ref={ref}
        className={cn(baseStyles, variants[variant], className)}
        {...props}
      >
        {children}
      </div>
    )
  }
)

Card.displayName = 'Card'

