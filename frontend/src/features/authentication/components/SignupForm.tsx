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
import { Mail, Lock, User, Eye, EyeOff } from 'lucide-react'
import toast from 'react-hot-toast'

const signupSchema = z.object({
  name: z.string().min(2, 'Tên phải có ít nhất 2 ký tự'),
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(8, 'Mật khẩu phải có ít nhất 8 ký tự').regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, 'Mật khẩu phải chứa chữ hoa, chữ thường và số'),
  confirmPassword: z.string().min(8, 'Xác nhận mật khẩu phải có ít nhất 8 ký tự'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Mật khẩu không khớp',
  path: ['confirmPassword'],
})

type SignupFormData = z.infer<typeof signupSchema>

export const SignupForm = () => {
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignupFormData>({
    resolver: zodResolver(signupSchema),
  })

  const onSubmit = async (data: SignupFormData) => {
    try {
      // TODO: Implement signup API call
      console.log('Signup:', data)
      toast.success('Đăng ký thành công! Vui lòng đăng nhập.')
    } catch (error) {
      const axiosError = error as { response?: { data?: { message?: string } } }
      toast.error(
        axiosError.response?.data?.message || 'Đăng ký thất bại. Vui lòng thử lại.'
      )
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
      {/* Name Input */}
      <div className="relative">
        <div className="absolute left-5 top-4 text-white/40">
          <User size={24} />
        </div>
        <input
          type="text"
          placeholder="Full name"
          {...register('name')}
          className="w-full pl-16 pr-5 py-4 bg-white/5 border border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md font-mono"
          required
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-400">{errors.name.message}</p>
        )}
      </div>

      {/* Email Input */}
      <div className="relative">
        <div className="absolute left-5 top-4 text-white/40">
          <Mail size={24} />
        </div>
        <input
          type="email"
          placeholder="Email address"
          {...register('email')}
          className="w-full pl-16 pr-5 py-4 bg-white/5 border border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md font-mono"
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
          className="w-full pl-16 pr-14 py-4 bg-white/5 border border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md font-mono"
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

      {/* Confirm Password Input */}
      <div className="relative">
        <div className="absolute left-5 top-4 text-white/40">
          <Lock size={24} />
        </div>
        <input
          type={showConfirmPassword ? 'text' : 'password'}
          placeholder="Confirm password"
          {...register('confirmPassword')}
          className="w-full pl-16 pr-14 py-4 bg-white/5 border border-white/15 rounded-2xl text-white text-lg placeholder-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition-all duration-300 backdrop-blur-md font-mono"
          required
        />
        <button
          type="button"
          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
          className="absolute right-5 top-4 text-white/40 hover:text-white/60 transition-colors"
        >
          {showConfirmPassword ? <EyeOff size={24} /> : <Eye size={24} />}
        </button>
        {errors.confirmPassword && (
          <p className="mt-1 text-sm text-red-400">{errors.confirmPassword.message}</p>
        )}
      </div>

      {/* Terms Checkbox */}
      <label className="flex items-center gap-2 cursor-pointer">
        <input
          type="checkbox"
          className="w-5 h-5 rounded bg-white/5 border border-white/15 accent-white/50"
          required
        />
        <span className="text-white/60 text-base font-mono">
          I agree to the <a href="#" className="text-white/80 hover:text-white">terms</a>
        </span>
      </label>

      {/* Submit Button */}
      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full py-4 bg-white/10 hover:bg-white/15 border border-white/20 rounded-2xl text-white font-semibold text-lg transition-all duration-300 hover:shadow-xl active:scale-95 font-mono disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isSubmitting ? 'Đang tạo tài khoản...' : 'Create Account'}
      </button>
    </form>
  )
}
