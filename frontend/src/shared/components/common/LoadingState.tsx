import type { ReactNode } from 'react'
import { Spinner } from '@shared/components/ui/Spinner'
import { cn } from '@shared/utils/cn'

interface LoadingStateProps {
  size?: 'sm' | 'md' | 'lg'
  text?: string
  fullScreen?: boolean
  className?: string
}

export const LoadingState = ({
  size = 'md',
  text = 'Loading...',
  fullScreen = false,
  className,
}: LoadingStateProps) => {
  const content = (
    <div
      className={cn(
        'flex flex-col items-center justify-center gap-3 text-white',
        className
      )}
    >
      <Spinner size={size} variant="white" />
      {text && <p className="text-sm text-gray-300">{text}</p>}
    </div>
  )

  if (fullScreen) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-black/60 backdrop-blur-xl">
        {content}
      </div>
    )
  }

  return content
}

interface LoadingOverlayProps {
  isLoading: boolean
  children: ReactNode
  text?: string
}

export const LoadingOverlay = ({
  isLoading,
  children,
  text = 'Loading...',
}: LoadingOverlayProps) => {
  if (!isLoading) return <>{children}</>

  return (
    <div className="relative">
      <div className="opacity-50 pointer-events-none">{children}</div>
      <div className="absolute inset-0 flex items-center justify-center bg-black/70 backdrop-blur">
        <LoadingState text={text} />
      </div>
    </div>
  )
}

