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

import { useEffect, useState, useRef } from 'react'
import { createPortal } from 'react-dom'
import { User, Mail, Shield, Calendar, X, Camera, Loader } from 'lucide-react'
import { authService } from '@features/authentication/services/authService'
import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import toast from 'react-hot-toast'

interface UserProfile {
  id: string
  email: string
  name?: string
  role?: string
  avatarUrl?: string
  createdAt?: string
  lastLogin?: string
}

interface ProfilePopupProps {
  isOpen: boolean
  onClose: () => void
}

export const ProfilePopup = ({ isOpen, onClose }: ProfilePopupProps) => {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isUploading, setIsUploading] = useState(false)
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null)
  const popupRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (isOpen) {
      fetchProfile()
    }
  }, [isOpen])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (popupRef.current && !popupRef.current.contains(event.target as Node)) {
        onClose()
      }
    }

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      document.addEventListener('keydown', handleEscape)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [isOpen, onClose])

  const fetchProfile = async () => {
    setIsLoading(true)
    try {
      // Check localStorage for avatar first
      const savedAvatar = localStorage.getItem('user_avatar')
      if (savedAvatar) {
        setAvatarPreview(savedAvatar)
      }

      // Try to get profile from API
      const response = await api.get<{ success: boolean; data: UserProfile & { avatar_url?: string } }>(
        API_ENDPOINTS.AUTH.PROFILE
      )
      if (response.data.success && response.data.data) {
        const profileData = response.data.data
        setProfile(profileData)
        // Use avatar from API (from DB) - check both avatarUrl and avatar_url (snake_case from backend)
        const avatarUrl = profileData.avatarUrl || profileData.avatar_url
        if (avatarUrl) {
          setAvatarPreview(avatarUrl)
          localStorage.setItem('user_avatar', avatarUrl)
          // Dispatch event to notify Header component
          window.dispatchEvent(new CustomEvent('avatarUpdated'))
        } else if (savedAvatar) {
          setAvatarPreview(savedAvatar)
        }
      } else {
        // Fallback: Get from token or localStorage
        const token = authService.getAccessToken()
        if (token) {
          // Decode token to get user info (if available)
          try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            setProfile({
              id: payload.sub || payload.id || 'N/A',
              email: payload.email || 'N/A',
              name: payload.name || payload.username || 'User',
              role: payload.role || 'user',
            })
          } catch {
            // If can't decode, use default
            setProfile({
              id: 'N/A',
              email: 'user@ibn.vn',
              name: 'User',
              role: 'user',
            })
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch profile:', error)
      // Fallback to default
      setProfile({
        id: 'N/A',
        email: 'user@ibn.vn',
        name: 'User',
        role: 'user',
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleAvatarUpload = async (file: File) => {
    // Validate file
    const maxSize = 5 * 1024 * 1024 // 5MB
    if (file.size > maxSize) {
      toast.error('Ảnh không được vượt quá 5MB')
      return
    }

    const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
    if (!validTypes.includes(file.type)) {
      toast.error('Chỉ chấp nhận file ảnh: JPG, PNG, WebP')
      return
    }

    setIsUploading(true)
    try {
      // Create preview
      const reader = new FileReader()
      reader.onloadend = () => {
        const result = reader.result as string
        setAvatarPreview(result)
        // Save to localStorage immediately for instant preview
        localStorage.setItem('user_avatar', result)
      }
      reader.readAsDataURL(file)

      // Upload to server
      const formData = new FormData()
      formData.append('avatar', file)

      const response = await api.post<{ success: boolean; data: { avatarUrl: string } }>(
        API_ENDPOINTS.AUTH.UPLOAD_AVATAR,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        }
      )

      if (response.data.success && response.data.data) {
        const avatarUrl = response.data.data.avatarUrl
        setProfile((prev) => prev ? { ...prev, avatarUrl } : null)
        localStorage.setItem('user_avatar', avatarUrl)
        
        // Dispatch event to notify Header component
        window.dispatchEvent(new CustomEvent('avatarUpdated'))
        
        toast.success('Cập nhật avatar thành công!')
      }
    } catch (error) {
      console.error('Failed to upload avatar:', error)
      toast.error('Không thể upload avatar. Vui lòng thử lại.')
      // Keep preview even if upload fails
    } finally {
      setIsUploading(false)
    }
  }

  const handleAvatarClick = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      handleAvatarUpload(file)
    }
    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  if (!isOpen) return null

  // Sử dụng Portal để render ở root level, đảm bảo hiển thị đúng
  return createPortal(
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999] transition-opacity"
        onClick={onClose}
      />

      {/* Popup - Căn giữa trang */}
      <div
        ref={popupRef}
        className="fixed inset-0 z-[9999] flex items-center justify-center p-4 pointer-events-none"
      >
        <div
          className="relative w-full max-w-md rounded-3xl p-8 transition-all duration-300 pointer-events-auto"
          onClick={(e) => e.stopPropagation()}
          style={{
            background: 'linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.05) 100%)',
            backdropFilter: 'blur(30px) saturate(180%)',
            WebkitBackdropFilter: 'blur(30px) saturate(180%)',
            border: '1px solid rgba(255, 255, 255, 0.3)',
            boxShadow: `
              0 20px 60px 0 rgba(0, 0, 0, 0.5),
              inset 0 1px 0 0 rgba(255, 255, 255, 0.2),
              inset 0 -1px 0 0 rgba(255, 255, 255, 0.1)
            `,
          }}
        >
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-2 rounded-full hover:bg-white/10 transition-colors"
            aria-label="Close"
          >
            <X className="h-5 w-5 text-white/70" />
          </button>

          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-2 border-white/20 border-t-white/60"></div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Header with Avatar */}
              <div className="text-center">
                {/* Avatar with upload button */}
                <div className="relative inline-block mb-4">
                  <div
                    onClick={handleAvatarClick}
                    className="relative w-24 h-24 rounded-full border-2 border-white/30 overflow-hidden cursor-pointer hover:border-white/50 transition-all group bg-white/10"
                  >
                    {avatarPreview ? (
                      <img
                        src={avatarPreview}
                        alt="Avatar"
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <User className="h-12 w-12 text-white/80" />
                      </div>
                    )}
                    {/* Upload overlay */}
                    <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                      {isUploading ? (
                        <Loader className="h-6 w-6 text-white animate-spin" />
                      ) : (
                        <Camera className="h-6 w-6 text-white" />
                      )}
                    </div>
                  </div>
                  {/* Upload hint */}
                  <p className="text-xs text-white/50 mt-2 font-mono">
                    Click để đổi avatar
                  </p>
                </div>

                {/* Hidden file input */}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/jpg,image/png,image/webp"
                  onChange={handleFileChange}
                  className="hidden"
                />

                <h2 className="text-2xl font-bold text-white mb-1 font-mono">
                  {profile?.name || 'User'}
                </h2>
                <p className="text-white/60 text-sm font-mono">
                  User Profile
                </p>
              </div>

              {/* Profile Info */}
              <div className="space-y-4">
                {/* Email */}
                <div className="flex items-start gap-3 p-4 rounded-2xl bg-white/5 border border-white/10">
                  <Mail className="h-5 w-5 text-white/60 mt-0.5 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-white/50 font-mono mb-1">Email</p>
                    <p className="text-white font-medium font-mono break-all">
                      {profile?.email || 'N/A'}
                    </p>
                  </div>
                </div>

                {/* Role */}
                {profile?.role && (
                  <div className="flex items-start gap-3 p-4 rounded-2xl bg-white/5 border border-white/10">
                    <Shield className="h-5 w-5 text-white/60 mt-0.5 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-white/50 font-mono mb-1">Role</p>
                      <p className="text-white font-medium font-mono capitalize">
                        {profile.role}
                      </p>
                    </div>
                  </div>
                )}

                {/* User ID */}
                <div className="flex items-start gap-3 p-4 rounded-2xl bg-white/5 border border-white/10">
                  <User className="h-5 w-5 text-white/60 mt-0.5 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-white/50 font-mono mb-1">User ID</p>
                    <p className="text-white font-medium font-mono text-sm break-all">
                      {profile?.id || 'N/A'}
                    </p>
                  </div>
                </div>

                {/* Created At */}
                {profile?.createdAt && (
                  <div className="flex items-start gap-3 p-4 rounded-2xl bg-white/5 border border-white/10">
                    <Calendar className="h-5 w-5 text-white/60 mt-0.5 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-white/50 font-mono mb-1">Member Since</p>
                      <p className="text-white font-medium font-mono text-sm">
                        {new Date(profile.createdAt).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric',
                        })}
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </>,
    document.body
  )
}

