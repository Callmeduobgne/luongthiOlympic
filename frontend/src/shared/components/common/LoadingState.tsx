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

