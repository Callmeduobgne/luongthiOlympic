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
import { useLocation } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { Header } from './Header'
import { Footer } from './Footer'

interface LayoutProps {
  children: ReactNode
  showSidebar?: boolean
  showFooter?: boolean
}

export const Layout = ({
  children,
  showFooter = true,
}: LayoutProps) => {
  const location = useLocation()

  return (
    <div className="min-h-screen flex flex-col bg-black text-white relative">
      {/* Background Image */}
      <div
        className="fixed inset-0 z-0"
        style={{
          backgroundImage: 'url(/images/315a80aeaefdcc33d8098190d4a21c9c.jpg)',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          backgroundRepeat: 'no-repeat',
          backgroundAttachment: 'scroll',
        }}
      >
        {/* Dark overlay để text dễ đọc hơn */}
        <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      </div>

      {/* Content */}
      <div className="relative z-10 flex flex-col min-h-screen">
        <Header />
        <AnimatePresence mode="wait">
          <motion.main
            key={location.pathname}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -16 }}
            transition={{ duration: 0.35, ease: 'easeInOut' }}
            className="flex-1"
          >
            <div className="p-4 lg:p-6">{children}</div>
          </motion.main>
        </AnimatePresence>
        {showFooter && <Footer />}
      </div>
    </div>
  )
}
