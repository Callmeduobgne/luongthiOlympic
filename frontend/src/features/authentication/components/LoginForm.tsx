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

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Mail, Lock, Eye, EyeOff } from 'lucide-react'
import { useAuth } from '../hooks/useAuth'
import toast from 'react-hot-toast'

const loginSchema = z.object({
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(8, 'Mật khẩu phải có ít nhất 8 ký tự'),
})

type LoginFormData = z.infer<typeof loginSchema>

export const LoginForm = () => {
  const { login, isLoading } = useAuth()
  const [showPassword, setShowPassword] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      // Dev mode: Log request data
      if (import.meta.env.DEV) {
        console.log('🔐 [DEV] Login attempt:', { email: data.email, password: '***' })
      }
      
      await login(data)
      
      if (import.meta.env.DEV) {
        console.log('✅ [DEV] Login successful')
        // Check if token is stored
        const token = localStorage.getItem('accessToken')
        console.log('🔑 [DEV] Token in localStorage:', token ? 'YES' : 'NO')
      }
      
      toast.success('Đăng nhập thành công!')
      
      // Wait a bit for token to be stored, then redirect
      // Use window.location to force a full page reload and ensure ProtectedRoute checks token
      setTimeout(() => {
        const token = localStorage.getItem('accessToken')
        if (import.meta.env.DEV) {
          console.log('🔍 [DEV] Before redirect, token exists:', !!token)
        }
        if (token) {
          window.location.href = '/'
        } else {
          console.error('❌ [DEV] No token found, cannot redirect')
          toast.error('Đăng nhập thành công nhưng không thể lưu token. Vui lòng thử lại.')
        }
      }, 300)
    } catch (error) {
      const axiosError = error as {
        message?: string
        response?: {
          data?: { message?: string; error?: { message?: string } }
          status?: number
          statusText?: string
        }
        config?: { url?: string; method?: string; headers?: unknown }
      }
      // Dev mode: Log detailed error
      if (import.meta.env.DEV) {
        console.error('❌ [DEV] Login error:', {
          message: axiosError.message,
          response: axiosError.response?.data,
          status: axiosError.response?.status,
          statusText: axiosError.response?.statusText,
          config: {
            url: axiosError.config?.url,
            method: axiosError.config?.method,
            headers: axiosError.config?.headers,
          },
        })
      }
      
      const errorMessage =
        axiosError.response?.data?.message ||
        axiosError.response?.data?.error?.message ||
        axiosError.message ||
        'Đăng nhập thất bại. Vui lòng thử lại.'
      toast.error(errorMessage)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
      {/* Email Input */}
      <div className="relative">
        <div className="absolute left-5 top-4 text-white/40">
          <Mail size={24} />
        </div>
        <input
          type="email"
          placeholder="Email address"
          {...register('email')}
          className="py-4 bg-white/5 border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md h-auto opacity-100 pl-16 pt-4 pr-16 font-mono border w-full"
          required
        />
        {errors.email && (
          <p className="mt-1 text-sm text-red-400">{errors.email.message}</p>
        )}
      </div>

      {/* Password Input */}
      <div className="relative">
        <div className="absolute left-5 top-4 text-white/40">
          <Lock size={24} />
        </div>
        <input
          type={showPassword ? 'text' : 'password'}
          placeholder="Password"
          {...register('password')}
          className="pl-16 pr-14 py-4 bg-white/5 border border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md font-mono w-full"
          required
        />
        <button
          type="button"
          onClick={() => setShowPassword(!showPassword)}
          className="absolute right-5 top-4 text-white/40 hover:text-white/60 transition-colors"
        >
          {showPassword ? <EyeOff size={24} /> : <Eye size={24} />}
        </button>
        {errors.password && (
          <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>
        )}
      </div>

      {/* Remember Me & Forgot Password */}
      <div className="flex items-center justify-between text-base">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            className="w-5 h-5 rounded bg-white/5 border accent-white/50 border-primary-foreground rounded-4xl opacity-80"
          />
          <span className="text-white/60 font-mono">Remember me</span>
        </label>
        <a href="#" className="text-white/50 hover:text-white/70 transition-colors font-mono">
          Forgot password?
        </a>
      </div>

      {/* Submit Button */}
      <button
        type="submit"
        disabled={isSubmitting || isLoading}
        className="py-4 bg-white/10 hover:bg-white/15 border border-white/20 rounded-2xl text-white font-semibold text-lg transition-all duration-300 hover:shadow-xl active:scale-95 font-mono w-full disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isSubmitting || isLoading ? 'Đang đăng nhập...' : 'Sign In'}
      </button>
    </form>
  )
}
