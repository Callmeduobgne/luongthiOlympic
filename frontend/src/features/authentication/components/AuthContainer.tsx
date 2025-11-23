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
import { LoginForm } from './LoginForm'
import { SignupForm } from './SignupForm'

export const AuthContainer = () => {
  const [isLogin, setIsLogin] = useState(true)

  return (
    <div className="relative w-full">
      {/* Khung đăng nhập với liquid glass effect - trong suốt tối đa */}
      <div 
        className="relative z-20 rounded-3xl p-8 sm:p-10 transition-all duration-300"
        style={{
          background: 'rgba(255, 255, 255, 0.03)',
          backdropFilter: 'blur(40px) saturate(150%)',
          WebkitBackdropFilter: 'blur(40px) saturate(150%)',
          border: '1px solid rgba(255, 255, 255, 0.18)',
          boxShadow: `
            0 8px 32px 0 rgba(0, 0, 0, 0.2),
            inset 0 1px 0 0 rgba(255, 255, 255, 0.15),
            inset 0 -1px 0 0 rgba(255, 255, 255, 0.03)
          `,
        }}
      >
        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-4xl sm:text-5xl font-bold text-white mb-3 font-mono">
            {isLogin ? 'Welcome Back' : 'Join Us'}
          </h1>
          <p className="text-white/80 text-base sm:text-lg font-mono">
            {isLogin ? 'Sign in to your account' : 'Create a new account'}
          </p>
        </div>

        {/* Forms */}
        <div className="mb-6">
          {isLogin ? (
            <LoginForm />
          ) : (
            <SignupForm />
          )}
        </div>

        {/* Toggle */}
        <div className="text-center border-t border-white/30 pt-6">
          <p className="text-white/70 text-sm sm:text-base mb-4 font-mono">
            {isLogin ? "Don't have an account?" : 'Already have an account?'}
          </p>
          <button
            onClick={() => setIsLogin(!isLogin)}
            className="px-6 py-2.5 sm:px-8 sm:py-3 bg-white/10 hover:bg-white/15 border border-white/30 rounded-xl text-white/90 font-medium text-base sm:text-lg transition-all duration-300 hover:shadow-lg font-mono"
          >
            {isLogin ? 'Sign Up' : 'Sign In'}
          </button>
        </div>
      </div>

      {/* Decorative blur elements */}
      <div className="absolute -top-20 -right-20 w-40 h-40 bg-white/5 rounded-full blur-3xl pointer-events-none z-0" />
      <div className="absolute -bottom-20 -left-20 w-32 h-32 bg-white/5 rounded-full blur-3xl pointer-events-none z-0" />
    </div>
  )
}
