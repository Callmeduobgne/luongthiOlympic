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

import { useState, useRef, useEffect, useLayoutEffect, useCallback } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import { LogOut, User } from 'lucide-react'
import {
  LayoutDashboard,
  Blocks,
  BarChart3,
  Network,
  Settings,
  Code,
  QrCode,
  Radio,
} from 'lucide-react'
import { authService } from '@features/authentication/services/authService'
import { useNavigate } from 'react-router-dom'
import { cn } from '@shared/utils/cn'
import { ProfilePopup } from '@shared/components/common/ProfilePopup'
import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

interface HeaderProps {
  onMenuClick?: () => void
  isSidebarOpen?: boolean
  // Props kept for backward compatibility but not used in horizontal menu layout
}

interface NavItem {
  label: string
  path: string
  icon: React.ComponentType<{ className?: string }>
  badge?: string | number
}

const navItems: NavItem[] = [
  {
    label: 'Dashboard',
    path: '/dashboard',
    icon: LayoutDashboard,
  },
  {
    label: 'Deploy Chaincode',
    path: '/deploy-chaincode',
    icon: Code,
  },
  {
    label: 'Blockchain Explorer',
    path: '/explorer',
    icon: Blocks,
  },
  {
    label: 'Analytics',
    path: '/analytics',
    icon: BarChart3,
  },
  {
    label: 'Network',
    path: '/network',
    icon: Network,
  },
  {
    label: 'QR Code Generator',
    path: '/qr-generator',
    icon: QrCode,
  },
  {
    label: 'NFC Manager',
    path: '/dashboard/nfc',
    icon: Radio,
  },
  {
    label: 'Settings',
    path: '/settings',
    icon: Settings,
  },
]

export const Header = ({ }: HeaderProps) => {
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [showProfilePopup, setShowProfilePopup] = useState(false)
  const [userAvatar, setUserAvatar] = useState<string | null>(null)
  const navigate = useNavigate()
  const isAuthenticated = authService.isAuthenticated()
  const location = useLocation()
  const navRefs = useRef<{ [key: string]: HTMLAnchorElement | null }>({})
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [highlightStyle, setHighlightStyle] = useState({ width: 0, x: 100, opacity: 100 })

  const updateHighlight = useCallback(() => {
    // Tìm mục menu active (có thể là exact match hoặc startsWith)
    const activePath = navItems.find(
      (item) =>
        location.pathname === item.path ||
        (item.path !== '/' && location.pathname.startsWith(item.path))
    )?.path

    if (!activePath) {
      setHighlightStyle((prev: { width: number; x: number; opacity: number }) => ({ ...prev, opacity: 0 }))
      return
    }

    const activeRef = navRefs.current[activePath]
    const container = containerRef.current

    if (activeRef && container) {
      const containerRect = container.getBoundingClientRect()
      const activeRect = activeRef.getBoundingClientRect()

      // ============================================
      // ĐIỀU CHỈNH KÍCH THƯỚC HIGHLIGHT Ở ĐÂY
      // ============================================
      // widthOffset: Thêm/bớt độ rộng (px)
      //   - Số dương: highlight rộng hơn NavLink
      //   - Số âm: highlight hẹp hơn NavLink
      //   - Ví dụ: 16 = rộng thêm 8px mỗi bên
      const widthOffset = 0 // Thay đổi giá trị này để điều chỉnh độ rộng

      // xOffset: Điều chỉnh vị trí ngang (px)
      //   - Số dương: dịch sang phải
      //   - Số âm: dịch sang trái
      const xOffset = -80 // Thay đổi giá trị này để điều chỉnh vị trí ngang

      // Tính toán vị trí và kích thước
      const x = activeRect.left - containerRect.left - (widthOffset / 2) + xOffset
      const width = activeRect.width + widthOffset

      setHighlightStyle({
        width,
        x,
        opacity: 1,
      })
    } else if (!activeRef) {
      // Nếu chưa có ref, thử lại trong frame tiếp theo
      requestAnimationFrame(() => {
        requestAnimationFrame(updateHighlight)
      })
    }
  }, [location.pathname])

  // ============================================
  // TỐI ƯU TÍNH TOÁN VỊ TRÍ - ĐIỀU CHỈNH Ở ĐÂY
  // ============================================
  // useLayoutEffect: Tính toán TRƯỚC khi browser paint
  //   - Tránh flicker, animation mượt hơn
  //   - Đã tối ưu, không cần thay đổi
  useLayoutEffect(() => {
    updateHighlight()
  }, [updateHighlight])

  // useEffect: Tính toán lại khi resize (với debounce)
  useEffect(() => {
    let resizeTimer: ReturnType<typeof setTimeout>

    const handleResize = () => {
      clearTimeout(resizeTimer)
      // Debounce: Chờ 16ms (~1 frame) trước khi tính toán lại
      //   - 16ms: 1 frame ở 60fps (khuyến nghị)
      //   - Có thể tăng lên 32ms nếu muốn giảm tải hơn
      resizeTimer = setTimeout(() => {
        updateHighlight()
      }, 16) // Thay đổi giá trị này nếu cần
    }

    // Double requestAnimationFrame: Đảm bảo tính toán đúng thời điểm
    //   - Frame 1: Đợi layout hoàn tất
    //   - Frame 2: Tính toán vị trí chính xác
    //   - Đã tối ưu, không cần thay đổi
    requestAnimationFrame(() => {
      requestAnimationFrame(updateHighlight)
    })

    // passive: true = Tối ưu scroll performance
    window.addEventListener('resize', handleResize, { passive: true })

    return () => {
      clearTimeout(resizeTimer)
      window.removeEventListener('resize', handleResize)
    }
  }, [updateHighlight])

  // Fetch user avatar from API or localStorage
  useEffect(() => {
    if (isAuthenticated) {
      fetchUserAvatar()

      // Listen for avatar updates from ProfilePopup
      const handleAvatarUpdate = () => {
        fetchUserAvatar()
      }
      window.addEventListener('avatarUpdated', handleAvatarUpdate)

      return () => {
        window.removeEventListener('avatarUpdated', handleAvatarUpdate)
      }
    } else {
      setUserAvatar(null)
    }
  }, [isAuthenticated])

  const fetchUserAvatar = async () => {
    try {
      // Check localStorage first for quick display
      const savedAvatar = localStorage.getItem('user_avatar')
      if (savedAvatar) {
        setUserAvatar(savedAvatar)
      }

      // Fetch from API to get latest from DB
      const response = await api.get<{ success: boolean; data: { avatar_url?: string; avatarUrl?: string } }>(
        API_ENDPOINTS.AUTH.PROFILE
      )

      if (response.data.success && response.data.data) {
        // Check both avatar_url (snake_case from backend) and avatarUrl (camelCase)
        const avatarUrl = response.data.data.avatar_url || response.data.data.avatarUrl
        if (avatarUrl) {
          setUserAvatar(avatarUrl)
          // Update localStorage with latest from DB
          localStorage.setItem('user_avatar', avatarUrl)
        } else if (savedAvatar) {
          // Keep localStorage avatar if API doesn't return one
          setUserAvatar(savedAvatar)
        }
      } else if (savedAvatar) {
        // Keep localStorage avatar if API doesn't return one
        setUserAvatar(savedAvatar)
      }
    } catch (error) {
      console.error('Failed to fetch user avatar:', error)
      // Fallback to localStorage if API fails
      const savedAvatar = localStorage.getItem('user_avatar')
      if (savedAvatar) {
        setUserAvatar(savedAvatar)
      }
    }
  }

  const handleLogout = () => {
    authService.logout()
    localStorage.removeItem('user_avatar') // Clear avatar on logout
    setUserAvatar(null)
    navigate('/login')
  }

  return (
    <header className="sticky top-0 z-40 w-full border-b border-white/10 bg-black/60 backdrop-blur-2xl">
      <div className="flex flex-col">
        {/* Top bar: Logo + User menu */}
        <div className="flex h-16 items-center justify-between px-4 lg:px-6 text-white">
          {/* Left: Logo */}
          <button
            onClick={() => navigate('/page')}
            className="flex items-center gap-2 hover:opacity-80 transition-opacity cursor-pointer"
            aria-label="Go to Platform Page"
          >
            <div className="h-9 w-9 rounded-2xl border border-white/30 bg-white/10 flex items-center justify-center backdrop-blur">
              <span className="text-white font-bold text-sm tracking-wide">IBN</span>
            </div>
            <span className="font-semibold text-lg">
              IBN Network
            </span>
          </button>

          {/* Right: User menu or Login button */}
          {!isAuthenticated ? (
            <button
              onClick={() => navigate('/login')}
              className="px-4 py-2 rounded-xl border border-white/15 bg-white/5 hover:bg-white/10 transition text-sm font-medium"
            >
              Đăng nhập
            </button>
          ) : (
            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-2 p-2 rounded-xl border border-white/15 bg-white/5 hover:bg-white/10 transition focus:outline-none focus:ring-2 focus:ring-white/30"
                aria-label="User menu"
              >
                <div className="h-8 w-8 rounded-full bg-white/20 flex items-center justify-center overflow-hidden border border-white/30">
                  {userAvatar ? (
                    <img
                      src={userAvatar}
                      alt="User avatar"
                      className="w-full h-full object-cover"
                      onError={() => {
                        // Fallback to icon if image fails to load
                        setUserAvatar(null)
                      }}
                    />
                  ) : (
                    <User className="h-5 w-5 text-white/80" />
                  )}
                </div>
              </button>

              {/* Dropdown menu */}
              {showUserMenu && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowUserMenu(false)}
                  />
                  <div className="absolute right-0 mt-2 w-48 rounded-2xl shadow-2xl bg-black/85 border border-white/15 backdrop-blur-xl z-20">
                    <div className="py-1">
                      <button
                        onClick={() => {
                          setShowUserMenu(false)
                          navigate('/dashboard')
                        }}
                        className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-white/5 flex items-center gap-2 transition"
                      >
                        <LayoutDashboard className="h-4 w-4" />
                        Dashboard
                      </button>
                      <button
                        onClick={() => {
                          setShowUserMenu(false)
                          setShowProfilePopup(true)
                        }}
                        className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-white/5 flex items-center gap-2 transition"
                      >
                        <User className="h-4 w-4" />
                        Profile
                      </button>
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-white/5 flex items-center gap-2 transition"
                      >
                        <LogOut className="h-4 w-4" />
                        Logout
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>
          )}
        </div>

        {/* Navigation menu: Horizontal, centered - Only show when authenticated and not on /page */}
        {isAuthenticated && location.pathname !== '/page' && (
          <nav className="flex items-center justify-center border-t border-white/10 bg-black/40 backdrop-blur-xl">
            <div
              ref={containerRef}
              className="relative flex items-center gap-5 px-20 py-2"
            >
              {/* Animated highlight background - sử dụng transform để GPU accelerate */}
              {/* 
                ĐIỀU CHỈNH CHIỀU CAO (HEIGHT):
                - Thay đổi "inset-y-2" trong className bên dưới
                - inset-y-0 = không có khoảng cách (cao nhất)
                - inset-y-1 = 4px khoảng cách mỗi bên
                - inset-y-2 = 8px khoảng cách mỗi bên (hiện tại)
                - inset-y-3 = 12px khoảng cách mỗi bên
                - inset-y-4 = 16px khoảng cách mỗi bên (thấp nhất)
              */}
              <motion.span
                className="absolute inset-y-2 rounded-2xl border border-white/35 bg-white/15 shadow-[0_10px_25px_rgba(15,15,15,0.6)] pointer-events-none"
                style={{
                  // ============================================
                  // TỐI ƯU CSS PERFORMANCE - ĐIỀU CHỈNH Ở ĐÂY
                  // ============================================
                  // willChange: Báo cho browser biết property nào sẽ thay đổi
                  //   - Giúp browser tối ưu rendering trước
                  //   - Chỉ nên dùng khi thực sự cần (đã có sẵn)
                  willChange: 'transform, width, opacity',

                  // transformOrigin: Điểm gốc của transform
                  //   - 'left center': Transform từ bên trái (khuyến nghị)
                  transformOrigin: 'left center',

                  // backfaceVisibility: Ẩn mặt sau khi rotate (tối ưu rendering)
                  //   - 'hidden': Tối ưu tốt nhất (khuyến nghị)
                  backfaceVisibility: 'hidden',
                  WebkitBackfaceVisibility: 'hidden',

                  // perspective: Tạo không gian 3D (kích hoạt GPU acceleration)
                  //   - 1000: Giá trị tốt (khuyến nghị)
                  //   - Có thể tăng lên 2000 nếu cần
                  perspective: 2000,

                  // Có thể thêm các tối ưu khác:
                  // transform: 'translateZ(0)', // Force GPU acceleration
                  // isolation: 'isolate', // Tạo stacking context mới
                }}
                initial={false}
                animate={{
                  width: highlightStyle.width,
                  x: highlightStyle.x,
                  opacity: highlightStyle.opacity,
                }}
                transition={{
                  // ============================================
                  // TỐI ƯU ANIMATION MƯỢT MÀ - ĐIỀU CHỈNH Ở ĐÂY
                  // ============================================
                  width: {
                    type: 'spring',
                    // stiffness: Độ cứng của spring (càng thấp càng mềm mại)
                    //   - 100-150: Mềm mại, chậm (mượt mà nhất)
                    //   - 150-200: Cân bằng (khuyến nghị)
                    //   - 200-300: Nhanh, cứng (có thể giật)
                    stiffness: 100, // Giảm xuống 120-130 để mượt hơn

                    // damping: Độ giảm dao động (càng cao càng ít dao động)
                    //   - 20-25: Nhiều dao động (bouncy)
                    //   - 25-35: Cân bằng (khuyến nghị)
                    //   - 35-45: Ít dao động (mượt, ít bật)
                    damping: 25, // Tăng lên 35-40 để ít dao động hơn

                    // mass: Khối lượng (càng cao càng nặng, chậm hơn)
                    //   - 0.8-1.0: Nhẹ, nhanh
                    //   - 1.0-1.5: Cân bằng (khuyến nghị)
                    //   - 1.5-2.0: Nặng, chậm (mượt mà hơn)
                    mass: 1.9, // Tăng lên 1.5-1.8 để mượt hơn

                    // restDelta: Ngưỡng dừng (càng nhỏ càng chính xác)
                    restDelta: 0.001, // Giữ nguyên hoặc giảm xuống 0.001

                    // restSpeed: Tốc độ dừng (càng thấp càng mượt)
                    restSpeed: 0.3, // Giảm xuống 0.3-0.4 để mượt hơn
                  },
                  x: {
                    type: 'spring',
                    // Các tham số giống như width ở trên
                    stiffness: 120, // Giảm xuống 120-130 để mượt hơn
                    damping: 30, // Tăng lên 35-40 để ít dao động
                    mass: 1.8, // Tăng lên 1.5-1.8 để mượt hơn
                    restDelta: 0.001,
                    restSpeed: 0.3, // Giảm xuống 0.3-0.4
                  },
                  opacity: {
                    // duration: Thời gian fade (giây)
                    //   - 0.08-0.12: Nhanh (khuyến nghị)
                    //   - 0.12-0.18: Cân bằng
                    //   - 0.18-0.25: Chậm (có thể lag)
                    duration: 0.1, // Giữ nguyên hoặc giảm xuống 0.1

                    // ease: Hàm easing (càng mượt càng tốt)
                    //   - [0.25, 0.1, 0.25, 1]: Mượt mà (khuyến nghị)
                    //   - [0.4, 0, 0.2, 1]: Nhanh hơn
                    //   - [0.2, 0, 0, 1]: Rất mượt
                    ease: [0.2, 0., 0., 1], // Hoặc thử [0.2, 0, 0, 1]
                  },
                }}
              />

              {navItems.map((item) => {
                const Icon = item.icon
                const isActive = location.pathname === item.path ||
                  (item.path !== '/' && location.pathname.startsWith(item.path))

                return (
                  <NavLink
                    key={item.path}
                    ref={(el: HTMLAnchorElement | null) => {
                      navRefs.current[item.path] = el
                    }}
                    to={item.path}
                    className={cn(
                      'relative flex items-center gap-3 px-8 py-3 text-sm font-medium transition-all duration-300 border border-transparent overflow-hidden group rounded-2xl bg-white/5 text-gray-200 hover:bg-white/10 hover:border-white/20 z-10',
                      isActive
                        ? 'text-white'
                        : ''
                    )}
                  >
                    {/* Glow effect on hover */}
                    <span className="absolute inset-0 bg-gradient-to-r from-transparent via-white/25 to-transparent opacity-0 group-hover:opacity-100 group-hover:animate-shimmer transition-opacity duration-300" />

                    <Icon className="h-4 w-4 relative z-10 group-hover:scale-110 transition-transform duration-300 text-white" />
                    <span className="relative z-10">{item.label}</span>
                    {item.badge && (
                      <span className="px-2 py-0.5 text-xs rounded-full bg-white/20 text-white relative z-10">
                        {item.badge}
                      </span>
                    )}

                    {/* Shine effect */}
                    <span className="absolute inset-0 -translate-x-full group-hover:translate-x-full transition-transform duration-700 bg-gradient-to-r from-transparent via-white/60 to-transparent" />
                  </NavLink>
                )
              })}
            </div>
          </nav>
        )}
      </div>

      {/* Profile Popup - Render ở root level */}
      <ProfilePopup
        isOpen={showProfilePopup}
        onClose={() => setShowProfilePopup(false)}
      />
    </header>
  )
}
