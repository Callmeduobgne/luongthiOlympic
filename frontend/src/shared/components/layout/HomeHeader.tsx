/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

/**
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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

import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { LogOut, User, LayoutDashboard, Menu, X, ChevronRight } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { authService } from '@features/authentication/services/authService'
import { ProfilePopup } from '@shared/components/common/ProfilePopup'
import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

interface UserInfo {
  email: string
  name?: string
  role?: string
}

export const HomeHeader = () => {
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [showProfilePopup, setShowProfilePopup] = useState(false)
  const [userAvatar, setUserAvatar] = useState<string | null>(null)
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null)
  const navigate = useNavigate()
  const isAuthenticated = authService.isAuthenticated()
  const menuRef = useRef<HTMLDivElement>(null)
  const mobileMenuRef = useRef<HTMLDivElement>(null)

  // Fetch user info and avatar from API or localStorage
  useEffect(() => {
    if (isAuthenticated) {
      fetchUserInfo()
      fetchUserAvatar()

      // Listen for avatar updates from ProfilePopup
      const handleAvatarUpdate = () => {
        fetchUserAvatar()
        fetchUserInfo()
      }
      window.addEventListener('avatarUpdated', handleAvatarUpdate)

      return () => {
        window.removeEventListener('avatarUpdated', handleAvatarUpdate)
      }
    } else {
      setUserAvatar(null)
      setUserInfo(null)
    }
  }, [isAuthenticated])

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node

      // Only close if click is outside BOTH menus (and the button itself is not clicked - handled by button logic)
      if (showUserMenu) {
        if (menuRef.current && !menuRef.current.contains(target) &&
          (!mobileMenuRef.current || !mobileMenuRef.current.contains(target))) {
          setShowUserMenu(false)
        }
      }
    }

    if (showUserMenu) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showUserMenu])

  const fetchUserInfo = async () => {
    try {
      // Try to get user info from API
      const response = await api.get<{ success: boolean; data: { email?: string; name?: string; role?: string } }>(
        API_ENDPOINTS.AUTH.PROFILE
      )

      if (response.data.success && response.data.data) {
        const data = response.data.data
        setUserInfo({
          email: data.email || 'user@ibn.vn',
          name: data.name,
          role: data.role,
        })
      } else {
        // Fallback: Get from token
        const token = authService.getAccessToken()
        if (token) {
          try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            setUserInfo({
              email: payload.email || 'user@ibn.vn',
              name: payload.name || payload.username,
              role: payload.role,
            })
          } catch {
            setUserInfo({ email: 'user@ibn.vn' })
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch user info:', error)
      // Fallback: Get from token
      const token = authService.getAccessToken()
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]))
          setUserInfo({
            email: payload.email || 'user@ibn.vn',
            name: payload.name || payload.username,
            role: payload.role,
          })
        } catch {
          setUserInfo({ email: 'user@ibn.vn' })
        }
      }
    }
  }

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
    localStorage.removeItem('user_avatar')
    setUserAvatar(null)
    setUserInfo(null)
    navigate('/')
  }

  if (!isAuthenticated) {
    return null // Don't show header if not authenticated
  }

  return (
    <>
      <header className="bg-white/80 backdrop-blur-md shadow-sm fixed top-0 w-full z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* Left: Logo */}
            <button
              onClick={() => navigate('/')}
              className="flex items-center gap-3 hover:opacity-80 transition-opacity cursor-pointer"
              aria-label="Go to Homepage"
            >
              <img
                src="/images2/image copy.png"
                alt="IBN Tea Logo"
                className="h-12 w-auto object-contain"
              />
              <span className="text-2xl font-bold" style={{ color: '#22C55E' }}>
                Ibn tea
              </span>
            </button>

            {/* Center: Navigation links */}
            <nav className="hidden md:flex items-center gap-8">
              <a href="#products" className="text-gray-700 hover:text-green-600 transition-colors font-medium">
                Sản phẩm
              </a>
              <a
                href="#journey"
                onClick={(e) => {
                  e.preventDefault()
                  const element = document.getElementById('journey')
                  if (element) {
                    element.scrollIntoView({ behavior: 'smooth' })
                  } else {
                    window.location.href = '/#journey'
                  }
                }}
                className="text-gray-700 hover:text-green-600 transition-colors font-medium"
              >
                Câu Chuyện Cây Chè
              </a>
              <a href="#uses" className="text-gray-700 hover:text-green-600 transition-colors font-medium">
                Công dụng
              </a>
              <a href="#about" className="text-gray-700 hover:text-green-600 transition-colors font-medium">
                Giới thiệu
              </a>
            </nav>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setShowUserMenu(!showUserMenu)}
              className="md:hidden p-2 rounded-lg hover:bg-gray-100 transition-colors"
            >
              {showUserMenu ? (
                <X className="h-6 w-6 text-gray-600" />
              ) : (
                <Menu className="h-6 w-6 text-gray-600" />
              )}
            </button>

            {/* Right: User menu (Desktop) */}
            <div className="hidden md:block relative" ref={menuRef}>
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-3 px-3 py-2 rounded-xl border border-gray-200 bg-white hover:bg-gray-50 transition focus:outline-none focus:ring-2 focus:ring-green-500/30"
                aria-label="User menu"
              >
                <div className="h-8 w-8 rounded-full bg-gradient-to-r from-green-600 to-emerald-700 flex items-center justify-center overflow-hidden border-2 border-white shadow-md">
                  {userAvatar ? (
                    <img
                      src={userAvatar}
                      alt="User avatar"
                      className="w-full h-full object-cover"
                      onError={() => {
                        setUserAvatar(null)
                      }}
                    />
                  ) : (
                    <User className="h-5 w-5 text-white" />
                  )}
                </div>
                {userInfo && (
                  <div className="hidden md:block text-left">
                    <div className="text-sm font-medium text-gray-900">
                      {userInfo.name || userInfo.email.split('@')[0]}
                    </div>
                    <div className="text-xs text-gray-500">
                      {userInfo.email}
                    </div>
                  </div>
                )}
              </button>

              {/* Dropdown menu (Desktop) */}
              {showUserMenu && (
                <div className="absolute right-0 mt-2 w-56 rounded-2xl shadow-2xl bg-white border border-gray-200 z-20">
                  {/* User info section */}
                  {userInfo && (
                    <div className="px-4 py-3 border-b border-gray-100">
                      <div className="text-sm font-medium text-gray-900">
                        {userInfo.name || userInfo.email.split('@')[0]}
                      </div>
                      <div className="text-xs text-gray-500 truncate">
                        {userInfo.email}
                      </div>
                      {userInfo.role && (
                        <div className="text-xs text-gray-400 mt-1">
                          {userInfo.role}
                        </div>
                      )}
                    </div>
                  )}
                  <div className="py-1">
                    <button
                      onClick={() => {
                        setShowUserMenu(false)
                        navigate('/dashboard')
                      }}
                      className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition"
                    >
                      <LayoutDashboard className="h-4 w-4" />
                      Dashboard
                    </button>
                    <button
                      onClick={() => {
                        setShowUserMenu(false)
                        setShowProfilePopup(true)
                      }}
                      className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center gap-2 transition"
                    >
                      <User className="h-4 w-4" />
                      Profile
                    </button>
                    <div className="border-t border-gray-100 my-1" />
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2 transition"
                    >
                      <LogOut className="h-4 w-4" />
                      Đăng xuất
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Mobile Menu (Drawer) */}
        <AnimatePresence>
          {showUserMenu && (
            <>
              {/* Overlay */}
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={() => setShowUserMenu(false)}
                className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 md:hidden"
              />

              {/* Drawer */}
              <motion.div
                ref={mobileMenuRef}
                initial={{ x: '100%' }}
                animate={{ x: 0 }}
                exit={{ x: '100%' }}
                transition={{ type: 'spring', damping: 25, stiffness: 200 }}
                className="fixed top-0 right-0 bottom-0 w-[280px] bg-white shadow-2xl z-50 md:hidden overflow-y-auto"
              >
                <div className="p-4 flex flex-col h-full">
                  {/* Header */}
                  <div className="flex justify-between items-center mb-6">
                    <span className="text-xl font-bold text-green-600">Menu</span>
                    <button
                      onClick={() => setShowUserMenu(false)}
                      className="p-2 rounded-full hover:bg-gray-100 transition-colors"
                    >
                      <X className="h-6 w-6 text-gray-500" />
                    </button>
                  </div>

                  {/* User Info (if logged in) */}
                  {userInfo && (
                    <div className="mb-6 p-4 bg-gray-50 rounded-xl border border-gray-100">
                      <div className="flex items-center gap-3 mb-3">
                        <div className="h-10 w-10 rounded-full bg-gradient-to-r from-green-600 to-emerald-700 flex items-center justify-center overflow-hidden border-2 border-white shadow-md">
                          {userAvatar ? (
                            <img
                              src={userAvatar}
                              alt="User avatar"
                              className="w-full h-full object-cover"
                            />
                          ) : (
                            <User className="h-6 w-6 text-white" />
                          )}
                        </div>
                        <div>
                          <div className="font-medium text-gray-900 truncate max-w-[140px]">
                            {userInfo.name || userInfo.email.split('@')[0]}
                          </div>
                          <div className="text-xs text-gray-500 truncate max-w-[140px]">
                            {userInfo.email}
                          </div>
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <button
                          onClick={() => {
                            setShowUserMenu(false)
                            navigate('/dashboard')
                          }}
                          className="flex items-center justify-center gap-2 py-2 bg-white border border-gray-200 rounded-lg text-xs font-medium text-gray-700 hover:bg-gray-50 hover:border-green-200 transition-all"
                        >
                          <LayoutDashboard className="h-3.5 w-3.5" />
                          Dashboard
                        </button>
                        <button
                          onClick={() => {
                            setShowUserMenu(false)
                            setShowProfilePopup(true)
                          }}
                          className="flex items-center justify-center gap-2 py-2 bg-white border border-gray-200 rounded-lg text-xs font-medium text-gray-700 hover:bg-gray-50 hover:border-green-200 transition-all"
                        >
                          <User className="h-3.5 w-3.5" />
                          Profile
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Navigation Links */}
                  <nav className="flex flex-col gap-2">
                    <a
                      href="#technology"
                      onClick={(e) => {
                        e.preventDefault()
                        setShowUserMenu(false)
                        const element = document.getElementById('technology')
                        if (element) {
                          element.scrollIntoView({ behavior: 'smooth' })
                        } else {
                          window.location.href = '/#technology'
                        }
                      }}
                      className="flex items-center justify-between p-3 rounded-xl hover:bg-gray-50 text-gray-700 hover:text-green-600 transition-all font-medium"
                    >
                      Công nghệ
                      <ChevronRight className="h-4 w-4 text-gray-400" />
                    </a>
                    <a
                      href="#journey"
                      onClick={(e) => {
                        e.preventDefault()
                        setShowUserMenu(false)
                        const element = document.getElementById('journey')
                        if (element) {
                          element.scrollIntoView({ behavior: 'smooth' })
                        } else {
                          window.location.href = '/#journey'
                        }
                      }}
                      className="flex items-center justify-between p-3 rounded-xl hover:bg-gray-50 text-gray-700 hover:text-green-600 transition-all font-medium"
                    >
                      Câu Chuyện Cây Chè
                      <ChevronRight className="h-4 w-4 text-gray-400" />
                    </a>
                    <a
                      href="#transparency"
                      onClick={(e) => {
                        e.preventDefault()
                        setShowUserMenu(false)
                        const element = document.getElementById('transparency')
                        if (element) {
                          element.scrollIntoView({ behavior: 'smooth' })
                        } else {
                          window.location.href = '/#transparency'
                        }
                      }}
                      className="flex items-center justify-between p-3 rounded-xl hover:bg-gray-50 text-gray-700 hover:text-green-600 transition-all font-medium"
                    >
                      Minh bạch
                      <ChevronRight className="h-4 w-4 text-gray-400" />
                    </a>
                  </nav>

                  {/* Logout Button (at bottom) */}
                  <div className="mt-auto pt-6 border-t border-gray-100">
                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center justify-center gap-2 p-3 rounded-xl bg-red-50 text-red-600 font-medium hover:bg-red-100 transition-colors"
                    >
                      <LogOut className="h-5 w-5" />
                      Đăng xuất
                    </button>
                  </div>
                </div>
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </header>

      {/* Profile Popup */}
      <ProfilePopup
        isOpen={showProfilePopup}
        onClose={() => setShowProfilePopup(false)}
      />
    </>
  )
}

