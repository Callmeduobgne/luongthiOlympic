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

import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Globe, Code, Users, BarChart3, User, LogIn, LayoutDashboard, Shield, Zap, Lock, ArrowRight, CheckCircle2, X } from 'lucide-react'
import { authService } from '@features/authentication/services/authService'
import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import { verifyProductByBlockhash, type VerifyProductByBlockhashResponse } from '../services/productVerificationService'

interface UserProfile {
  id: string
  email: string
  name?: string
  role?: string
  avatar_url?: string
}

export const Page = () => {
  const navigate = useNavigate()
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isVerificationModalOpen, setIsVerificationModalOpen] = useState(false)
  const [blockhash, setBlockhash] = useState('')
  const [verificationResult, setVerificationResult] = useState<VerifyProductByBlockhashResponse | null>(null)
  const [isVerifying, setIsVerifying] = useState(false)

  const handleVerify = async () => {
    if (!blockhash.trim()) {
      alert('Vui lòng nhập blockhash hoặc transaction ID')
      return
    }
    setIsVerifying(true)
    try {
      const result = await verifyProductByBlockhash(blockhash.trim())
      setVerificationResult(result)
    } catch (error) {
      setVerificationResult({
        isValid: false,
        message: 'Đã xảy ra lỗi khi xác thực. Vui lòng thử lại sau.',
      })
    } finally {
      setIsVerifying(false)
    }
  }

  useEffect(() => {
    const checkAuth = async () => {
      const authenticated = authService.isAuthenticated()
      setIsAuthenticated(authenticated)

      if (authenticated) {
        try {
          const response = await api.get<{ success: boolean; data: UserProfile }>(
            API_ENDPOINTS.AUTH.PROFILE
          )
          if (response.data.success && response.data.data) {
            setUserProfile(response.data.data)
          }
        } catch (error) {
          console.error('Failed to fetch user profile:', error)
        }
      }
      setIsLoading(false)
    }

    checkAuth()
  }, [])

  const features = [
    {
      icon: <Globe className="h-8 w-8" />,
      title: 'Blockchain Network',
      description: 'Mạng lưới blockchain phân tán với Hyperledger Fabric, đảm bảo tính minh bạch và bảo mật cao',
      gradient: 'from-blue-500/20 to-cyan-500/20',
      iconColor: 'text-blue-400',
    },
    {
      icon: <Code className="h-8 w-8" />,
      title: 'Chaincode Management',
      description: 'Quản lý và triển khai chaincode một cách dễ dàng với giao diện trực quan và mạnh mẽ',
      gradient: 'from-green-500/20 to-emerald-500/20',
      iconColor: 'text-green-400',
    },
    {
      icon: <Shield className="h-8 w-8" />,
      title: 'Traceability',
      description: 'Truy xuất nguồn gốc chè từ khâu trồng trọt đến tiêu thụ, đảm bảo chất lượng và minh bạch',
      gradient: 'from-purple-500/20 to-pink-500/20',
      iconColor: 'text-purple-400',
    },
    {
      icon: <BarChart3 className="h-8 w-8" />,
      title: 'Real-time Analytics',
      description: 'Phân tích dữ liệu blockchain và metrics real-time để theo dõi hiệu suất hệ thống',
      gradient: 'from-amber-500/20 to-orange-500/20',
      iconColor: 'text-amber-400',
    },
    {
      icon: <Users className="h-8 w-8" />,
      title: 'User Management',
      description: 'Quản lý người dùng, phân quyền và xác thực với JWT và API keys một cách an toàn',
      gradient: 'from-red-500/20 to-rose-500/20',
      iconColor: 'text-red-400',
    },
    {
      icon: <Lock className="h-8 w-8" />,
      title: 'Enterprise Security',
      description: 'Bảo mật cấp doanh nghiệp với mã hóa end-to-end và quản lý chứng chỉ số',
      gradient: 'from-indigo-500/20 to-violet-500/20',
      iconColor: 'text-indigo-400',
    },
  ]

  const techStack = [
    { name: 'React + TypeScript', category: 'Frontend', color: 'text-blue-400' },
    { name: 'Go + Chi Router', category: 'Backend', color: 'text-green-400' },
    { name: 'Hyperledger Fabric', category: 'Blockchain', color: 'text-purple-400' },
    { name: 'PostgreSQL + Redis', category: 'Database', color: 'text-amber-400' },
  ]

  const benefits = [
    'Truy xuất nguồn gốc minh bạch',
    'Bảo mật cấp doanh nghiệp',
    'Hiệu suất cao và mở rộng dễ dàng',
    'Giao diện trực quan và dễ sử dụng',
  ]

  return (
    <div className="min-h-screen relative overflow-hidden">
      {/* Fixed Header */}
      <header className="fixed top-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-md border-b border-gray-200 shadow-sm">
        <nav className="container mx-auto max-w-7xl px-4 py-4">
          <div className="flex items-center justify-between">
            {/* Left side - empty for now */}
            <div className="flex-1"></div>
            
            {/* Center - Product verification box */}
            <div className="flex-1 flex justify-center">
              <button
                onClick={() => setIsVerificationModalOpen(true)}
                className="px-6 py-2 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-900 font-medium text-sm transition-colors border border-gray-300"
              >
                Xác thực sản phẩm của bạn
              </button>
            </div>
            
            {/* Right side - Navigation menu */}
            <div className="flex-1 flex items-center justify-end gap-8">
              {['Quy trình', 'Sản phẩm', 'Công dụng', 'Giới thiệu'].map((item) => (
                <button
                  key={item}
                  onClick={() => {
                    // Scroll to section or navigate
                    const element = document.getElementById(item.toLowerCase().replace(/\s+/g, '-'))
                    if (element) {
                      element.scrollIntoView({ behavior: 'smooth' })
                    }
                  }}
                  className="text-gray-900 font-medium text-sm hover:text-gray-700 transition-colors"
                >
                  {item}
                </button>
              ))}
            </div>
          </div>
        </nav>
      </header>

      {/* Background Image with Overlay */}
      <div
        className="absolute inset-0 z-0"
        style={{
          backgroundImage: 'url(/images/image%20copy.png)',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          backgroundRepeat: 'no-repeat',
          backgroundAttachment: 'scroll',
        }}
      />

      {/* Content */}
      <div className="relative z-10">
        {/* Hero Section */}
        <section className="relative min-h-screen flex items-center justify-center px-4 py-20 pt-24">
          <div className="container mx-auto max-w-7xl">
            <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-12">
              {/* Left: Hero Content */}
              <div className="flex-1 space-y-8 text-center lg:text-left">
                {/* Badge */}
                <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-gray-300 bg-white/90 backdrop-blur-xl text-sm text-gray-900 mb-4 shadow-md">
                  <Zap className="h-4 w-4 text-amber-500" />
                  <span className="uppercase tracking-wider">ICTU Blockchain Network</span>
                </div>

                {/* Main Heading */}
                <h1 className="text-5xl md:text-6xl lg:text-7xl font-bold leading-tight text-gray-900">
                  <span>
                    IBN Network
                  </span>
                  <br />
                  <span className="text-4xl md:text-5xl lg:text-6xl text-gray-800 font-light">
                    Hệ thống quản lý nguồn gốc chè
                  </span>
                </h1>

                {/* Description */}
                <p className="text-xl md:text-2xl text-gray-700 max-w-2xl mx-auto lg:mx-0 leading-relaxed">
                  Giải pháp blockchain toàn diện cho việc truy xuất nguồn gốc và quản lý chuỗi cung ứng chè
                  với công nghệ Hyperledger Fabric tiên tiến
                </p>

                {/* Benefits List */}
                <div className="flex flex-wrap gap-4 justify-center lg:justify-start">
                  {benefits.map((benefit, index) => (
                    <div
                      key={index}
                      className="flex items-center gap-2 px-4 py-2 rounded-full border border-gray-300 bg-white/90 backdrop-blur-xl shadow-sm"
                    >
                      <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                      <span className="text-sm text-gray-900">{benefit}</span>
                    </div>
                  ))}
                </div>

                {/* CTA Buttons */}
                <div className="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start pt-4">
                  {!isAuthenticated ? (
                    <button
                      onClick={() => navigate('/login')}
                      className="group px-8 py-4 rounded-2xl bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-semibold text-lg transition-all duration-300 shadow-lg shadow-blue-500/50 hover:shadow-xl hover:shadow-blue-500/50 hover:scale-105 flex items-center justify-center gap-2"
                    >
                      <span>Bắt đầu ngay</span>
                      <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                    </button>
                  ) : (
                    <button
                      onClick={() => navigate('/dashboard')}
                      className="group px-8 py-4 rounded-2xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-semibold text-lg transition-all duration-300 shadow-lg shadow-emerald-500/50 hover:shadow-xl hover:shadow-emerald-500/50 hover:scale-105 flex items-center justify-center gap-2"
                    >
                      <LayoutDashboard className="h-5 w-5" />
                      <span>Vào Dashboard</span>
                      <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                    </button>
                  )}
                  <button
                    onClick={() => navigate('/page')}
                    className="px-8 py-4 rounded-2xl border-2 border-gray-300 bg-white/90 hover:bg-white text-gray-900 font-semibold text-lg transition-all duration-300 backdrop-blur-xl shadow-md"
                  >
                    Tìm hiểu thêm
                  </button>
                </div>
              </div>

              {/* Right: User Card */}
              {!isLoading && (
                <div className="w-full lg:w-auto lg:min-w-[320px]">
                  <div
                    className="rounded-3xl border border-gray-200 bg-white/95 px-8 py-6 backdrop-blur-xl shadow-2xl"
                    style={{
                      background: 'rgba(255, 255, 255, 0.95)',
                      backdropFilter: 'blur(16px) saturate(160%)',
                      WebkitBackdropFilter: 'blur(16px) saturate(160%)',
                      border: '1px solid rgba(0, 0, 0, 0.1)',
                      boxShadow: `
                        0 20px 45px rgba(0, 0, 0, 0.15),
                        inset 0 1px 0 0 rgba(255, 255, 255, 0.5),
                        inset 0 -1px 0 0 rgba(0, 0, 0, 0.03)
                      `,
                    }}
                  >
                    {isAuthenticated && userProfile ? (
                      <div className="space-y-6">
                        <div className="flex items-center gap-4">
                          <div className="h-16 w-16 rounded-2xl bg-gray-100 flex items-center justify-center overflow-hidden border-2 border-gray-200 shadow-lg">
                            {userProfile.avatar_url ? (
                              <img
                                src={userProfile.avatar_url}
                                alt={userProfile.name || userProfile.email}
                                className="w-full h-full object-cover"
                              />
                            ) : (
                              <User className="h-8 w-8 text-gray-500" />
                            )}
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-lg font-bold text-gray-900 truncate">
                              {userProfile.name || userProfile.email}
                            </div>
                            <div className="text-sm text-gray-600 truncate">
                              {userProfile.email}
                            </div>
                            {userProfile.role && (
                              <div className="text-xs text-gray-600 mt-1 px-2 py-1 rounded-full bg-gray-100 inline-block">
                                {userProfile.role}
                              </div>
                            )}
                          </div>
                        </div>
                        <button
                          onClick={() => navigate('/dashboard')}
                          className="w-full px-6 py-3 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-semibold transition-all duration-300 shadow-lg shadow-emerald-500/50 hover:shadow-xl hover:shadow-emerald-500/50 hover:scale-105 flex items-center justify-center gap-2"
                        >
                          <LayoutDashboard className="h-5 w-5" />
                          Vào Dashboard
                        </button>
                      </div>
                    ) : (
                      <div className="space-y-6">
                        <div className="flex items-center gap-4">
                          <div className="h-16 w-16 rounded-2xl bg-gray-100 flex items-center justify-center border-2 border-gray-200 shadow-lg">
                            <User className="h-8 w-8 text-gray-500" />
                          </div>
                          <div className="flex-1">
                            <div className="text-lg font-bold text-gray-900">
                              Chưa đăng nhập
                            </div>
                            <div className="text-sm text-gray-600">
                              Đăng nhập để truy cập đầy đủ tính năng
                            </div>
                          </div>
                        </div>
                        <button
                          onClick={() => navigate('/login')}
                          className="w-full px-6 py-3 rounded-xl bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-semibold transition-all duration-300 shadow-lg shadow-blue-500/50 hover:shadow-xl hover:shadow-blue-500/50 hover:scale-105 flex items-center justify-center gap-2"
                        >
                          <LogIn className="h-5 w-5" />
                          Đăng nhập ngay
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="relative py-20 px-4">
          <div className="container mx-auto max-w-7xl">
            <div className="text-center mb-16">
              <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
                Tính năng nổi bật
              </h2>
              <p className="text-xl text-gray-700 max-w-2xl mx-auto">
                Khám phá các tính năng mạnh mẽ giúp quản lý blockchain hiệu quả
              </p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
              {features.map((feature, index) => (
                <div
                  key={index}
                  className="group relative overflow-hidden rounded-3xl border border-gray-200 bg-white/90 p-8 backdrop-blur-xl hover:bg-white transition-all duration-300 hover:scale-105 hover:shadow-2xl"
                  style={{
                    background: 'rgba(255, 255, 255, 0.9)',
                    border: '1px solid rgba(0, 0, 0, 0.1)',
                    boxShadow: `
                      0 8px 32px 0 rgba(0, 0, 0, 0.1),
                      inset 0 1px 0 0 rgba(255, 255, 255, 0.5)
                    `,
                  }}
                >
                  {/* Gradient Background */}
                  <div className={`absolute inset-0 bg-gradient-to-br ${feature.gradient} opacity-0 group-hover:opacity-60 transition-opacity duration-300`} />
                  
                  {/* Content */}
                  <div className="relative z-10">
                    <div className={`h-16 w-16 rounded-2xl bg-white border border-gray-200 flex items-center justify-center mb-6 ${feature.iconColor} group-hover:scale-110 transition-transform duration-300 shadow-sm`}>
                      {feature.icon}
                    </div>
                    <h3 className="text-2xl font-bold text-gray-900 mb-3">
                      {feature.title}
                    </h3>
                    <p className="text-gray-700 leading-relaxed">
                      {feature.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Tech Stack Section */}
        <section className="relative py-20 px-4">
          <div className="container mx-auto max-w-7xl">
            <div
              className="rounded-3xl border border-gray-200 bg-white/85 p-12 backdrop-blur-2xl"
              style={{
                background: 'rgba(255, 255, 255, 0.85)',
                border: '1px solid rgba(0, 0, 0, 0.1)',
                boxShadow: `
                  0 8px 32px 0 rgba(0, 0, 0, 0.1),
                  inset 0 1px 0 0 rgba(255, 255, 255, 0.5)
                `,
              }}
            >
              <div className="text-center mb-12">
                <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
                  Tech Stack
                </h2>
                <p className="text-xl text-gray-700">
                  Công nghệ hiện đại và mạnh mẽ
                </p>
              </div>
              <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
                {techStack.map((tech, index) => (
                  <div
                    key={index}
                    className="text-center p-6 rounded-2xl bg-white border border-gray-200 hover:bg-gray-50 transition-all duration-300 hover:scale-105 shadow-sm"
                  >
                    <div className={`font-bold text-lg ${tech.color} mb-2`}>
                      {tech.name}
                    </div>
                    <div className="text-sm text-gray-600">
                      {tech.category}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="relative py-20 px-4">
          <div className="container mx-auto max-w-4xl text-center">
            <div
              className="rounded-3xl border border-gray-200 bg-gradient-to-r from-blue-50 to-cyan-50 p-12 backdrop-blur-2xl"
              style={{
                background: 'linear-gradient(135deg, rgba(239, 246, 255, 0.95), rgba(236, 254, 255, 0.95))',
                border: '1px solid rgba(0, 0, 0, 0.1)',
                boxShadow: `
                  0 8px 32px 0 rgba(0, 0, 0, 0.1),
                  inset 0 1px 0 0 rgba(255, 255, 255, 0.5)
                `,
              }}
            >
              <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-6">
                Sẵn sàng bắt đầu?
              </h2>
              <p className="text-xl text-gray-700 mb-8 max-w-2xl mx-auto">
                Tham gia ngay để trải nghiệm hệ thống blockchain quản lý nguồn gốc chè hiện đại nhất
              </p>
              {!isAuthenticated ? (
                <button
                  onClick={() => navigate('/login')}
                  className="px-10 py-5 rounded-2xl bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-bold text-lg transition-all duration-300 shadow-lg shadow-blue-500/50 hover:shadow-xl hover:shadow-blue-500/50 hover:scale-105 flex items-center justify-center gap-3 mx-auto"
                >
                  <span>Đăng nhập ngay</span>
                  <ArrowRight className="h-6 w-6" />
                </button>
              ) : (
                <button
                  onClick={() => navigate('/dashboard')}
                  className="px-10 py-5 rounded-2xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-lg transition-all duration-300 shadow-lg shadow-emerald-500/50 hover:shadow-xl hover:shadow-emerald-500/50 hover:scale-105 flex items-center justify-center gap-3 mx-auto"
                >
                  <LayoutDashboard className="h-6 w-6" />
                  <span>Vào Dashboard</span>
                  <ArrowRight className="h-6 w-6" />
                </button>
              )}
            </div>
          </div>
        </section>
      </div>

      {/* Product Verification Modal */}
      {isVerificationModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center"
          onClick={() => {
            setIsVerificationModalOpen(false)
            setBlockhash('')
            setVerificationResult(null)
          }}
        >
          {/* Backdrop */}
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" />

          {/* Modal Content */}
          <div
            className="relative z-50 w-full max-w-lg rounded-2xl border border-gray-200 bg-white shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">
                Xác thực sản phẩm
              </h2>
              <button
                onClick={() => {
                  setIsVerificationModalOpen(false)
                  setBlockhash('')
                  setVerificationResult(null)
                }}
                className="rounded-full p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300"
                aria-label="Close modal"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {/* Body */}
            <div className="p-6">
              <div className="space-y-6">
                {!verificationResult ? (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-gray-900 mb-2">
                        Nhập Blockhash hoặc Transaction ID
                      </label>
                      <input
                        type="text"
                        value={blockhash}
                        onChange={(e) => setBlockhash(e.target.value)}
                        placeholder="Nhập blockhash hoặc transaction ID..."
                        className="w-full px-4 py-3 rounded-lg bg-gray-50 border border-gray-300 text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        disabled={isVerifying}
                        onKeyPress={(e) => {
                          if (e.key === 'Enter' && blockhash.trim() && !isVerifying) {
                            handleVerify()
                          }
                        }}
                      />
                    </div>
                    <button
                      onClick={handleVerify}
                      disabled={isVerifying || !blockhash.trim()}
                      className="w-full px-6 py-3 rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {isVerifying ? 'Đang xác thực...' : 'Xác thực'}
                    </button>
                  </>
                ) : (
                  <div className="space-y-4">
                    <div
                      className={`p-6 rounded-lg border-2 ${
                        verificationResult.isValid
                          ? 'bg-green-50 border-green-500 text-green-900'
                          : 'bg-red-50 border-red-500 text-red-900'
                      }`}
                    >
                      <div className="flex items-center gap-3 mb-2">
                        {verificationResult.isValid ? (
                          <CheckCircle2 className="h-6 w-6 text-green-600" />
                        ) : (
                          <X className="h-6 w-6 text-red-600" />
                        )}
                        <h3 className="text-lg font-semibold">
                          {verificationResult.message}
                        </h3>
                      </div>
                      {verificationResult.transactionId && (
                        <p className="text-sm mt-2 text-gray-700">
                          Transaction ID: <span className="font-mono">{verificationResult.transactionId}</span>
                        </p>
                      )}
                      {verificationResult.batchId && (
                        <p className="text-sm mt-1 text-gray-700">
                          Batch ID: <span className="font-mono">{verificationResult.batchId}</span>
                        </p>
                      )}
                      {verificationResult.packageId && (
                        <p className="text-sm mt-1 text-gray-700">
                          Package ID: <span className="font-mono">{verificationResult.packageId}</span>
                        </p>
                      )}
                    </div>
                    <button
                      onClick={() => {
                        setBlockhash('')
                        setVerificationResult(null)
                      }}
                      className="w-full px-6 py-3 rounded-lg bg-gray-200 hover:bg-gray-300 text-gray-900 font-semibold transition-colors"
                    >
                      Xác thực lại
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
