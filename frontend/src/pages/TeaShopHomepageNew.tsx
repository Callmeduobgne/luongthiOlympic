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

import { Link, useNavigate } from 'react-router-dom'
import { Shield, Leaf, QrCode, ChevronRight, CheckCircle, Hash, Radio, ArrowLeft, Camera, X, Lock, Eye, Sparkles, TrendingUp, Box, Layers, Hexagon, Menu } from 'lucide-react'
import { motion } from 'framer-motion'
import { HomeHeader } from '@shared/components/layout/HomeHeader'
import { authService } from '@features/authentication/services/authService'
import { useEffect, useState, useRef } from 'react'
import { Html5Qrcode } from 'html5-qrcode'
import toast from 'react-hot-toast'

type VerificationMethod = 'qr' | 'hash' | 'nfc' | null

export function TeaShopHomepage() {
    const isAuthenticated = authService.isAuthenticated()
    const navigate = useNavigate()
    const [showVerificationModal, setShowVerificationModal] = useState(false)
    const [selectedMethod, setSelectedMethod] = useState<VerificationMethod>(null)
    const [showMobileMenu, setShowMobileMenu] = useState(false)

    const [hash, setHash] = useState('')
    const [isScanning, setIsScanning] = useState(false)
    const [cameraError, setCameraError] = useState<string | null>(null)
    const qrCodeScannerRef = useRef<Html5Qrcode | null>(null)

    // QR Code Scanner functions
    const startQRScanner = async () => {
        const containerId = 'qr-reader-container'
        await new Promise(resolve => setTimeout(resolve, 100))
        const container = document.getElementById(containerId)
        if (!container) {
            console.error('Scanner container not found')
            setCameraError('Không tìm thấy container scanner')
            return
        }

        try {
            setCameraError(null)
            const scanner = new Html5Qrcode(containerId)
            qrCodeScannerRef.current = scanner

            await scanner.start(
                { facingMode: 'environment' },
                {
                    fps: 10,
                    qrbox: { width: 250, height: 250 },
                },
                (decodedText) => {
                    let packageId = decodedText
                    if (decodedText.includes('/verify/packages/')) {
                        packageId = decodedText.split('/verify/packages/')[1]?.split('?')[0] || decodedText
                    } else if (decodedText.includes('hash=')) {
                        packageId = decodedText.split('hash=')[1]?.split('&')[0] || decodedText
                    }

                    stopQRScanner()
                    setTimeout(() => {
                        if (packageId.trim()) {
                            navigate(`/verify/packages/${packageId.trim()}`)
                            setShowVerificationModal(false)
                        }
                    }, 500)
                },
                (_errorMessage) => { }
            )
            setIsScanning(true)
        } catch (error) {
            console.error('Failed to start camera:', error)
            const errorMsg = error instanceof Error ? error.message : 'Unknown error'
            if (errorMsg.includes('NotAllowedError') || errorMsg.includes('Permission denied')) {
                setCameraError('Quyền truy cập camera bị từ chối. Vui lòng cho phép truy cập camera trong cài đặt trình duyệt.')
            } else if (errorMsg.includes('NotFoundError') || errorMsg.includes('No camera')) {
                setCameraError('Không tìm thấy camera. Vui lòng kiểm tra thiết bị của bạn.')
            } else {
                setCameraError('Không thể truy cập camera. Vui lòng thử lại.')
            }
            setIsScanning(false)
        }
    }

    const stopQRScanner = async () => {
        if (qrCodeScannerRef.current) {
            try {
                await qrCodeScannerRef.current.stop()
                await qrCodeScannerRef.current.clear()
            } catch (error) {
                console.error('Error stopping scanner:', error)
            }
            qrCodeScannerRef.current = null
        }
        setIsScanning(false)
        setCameraError(null)
    }

    const startNfcScan = async () => {
        if (!window.isSecureContext) {
            toast.error('NFC chỉ hoạt động trên HTTPS hoặc localhost.')
            return
        }

        if (!('NDEFReader' in window)) {
            toast.error('Trình duyệt không hỗ trợ NFC. Vui lòng dùng Chrome trên Android.')
            return
        }

        try {
            // @ts-ignore
            const ndef = new window.NDEFReader()
            await ndef.scan()
            toast.loading('Đang chờ thẻ NFC... Vui lòng chạm thẻ vào mặt sau điện thoại.', { id: 'nfc-scan' })

            ndef.onreading = (event: any) => {
                const serialNumber = event.serialNumber
                if (serialNumber) {
                    toast.success('Đã đọc thẻ NFC thành công!', { id: 'nfc-scan' })

                    setTimeout(() => {
                        navigate(`/verify/nfc?tag=${encodeURIComponent(serialNumber)}`)
                        setShowVerificationModal(false)
                    }, 500)
                } else {
                    toast.error('Không đọc được mã thẻ', { id: 'nfc-scan' })
                }
            }
        } catch (error) {
            console.error('NFC Error:', error)
            toast.error('Không thể kích hoạt NFC. Vui lòng kiểm tra cài đặt.', { id: 'nfc-scan' })
        }
    }

    useEffect(() => {
        return () => {
            if (qrCodeScannerRef.current) {
                stopQRScanner()
            }
        }
    }, [])

    useEffect(() => {
        if (!showVerificationModal || selectedMethod !== 'qr') {
            stopQRScanner()
        }
    }, [showVerificationModal, selectedMethod])

    useEffect(() => {
        const handleScroll = () => {
            const scrolled = window.pageYOffset
            const parallaxElement = document.getElementById('parallax-background')
            if (parallaxElement) {
                const yPos = -(scrolled * 0.5)
                parallaxElement.style.transform = `translate3d(0, ${yPos}px, 0)`
            }
        }

        const updateBackgroundHeight = () => {
            const parallaxElement = document.getElementById('parallax-background')
            const contentElement = document.querySelector('[data-content]')
            if (parallaxElement && contentElement) {
                const contentHeight = contentElement.scrollHeight
                parallaxElement.style.height = `${Math.max(contentHeight, window.innerHeight)}px`
            }
        }

        window.addEventListener('scroll', handleScroll, { passive: true })
        window.addEventListener('resize', updateBackgroundHeight, { passive: true })
        setTimeout(updateBackgroundHeight, 100)

        return () => {
            window.removeEventListener('scroll', handleScroll)
            window.removeEventListener('resize', updateBackgroundHeight)
        }
    }, [])

    return (
        <div className="min-h-screen relative">
            {/* Background Image với Parallax Effect */}
            <div
                className="fixed inset-0 z-0"
                style={{
                    backgroundImage: 'url(/images2/BeautyPlus-Image-Enhancer-1764055417984.jpg)',
                    backgroundSize: 'cover',
                    backgroundPosition: 'center top',
                    backgroundRepeat: 'no-repeat',
                    backgroundAttachment: 'scroll',
                    willChange: 'transform',
                    minHeight: '100vh',
                }}
                id="parallax-background"
            />

            {/* Content */}
            <div className="relative z-10" data-content>
                {/* Header */}
                {isAuthenticated ? (
                    <HomeHeader />
                ) : (
                    <nav className="bg-white/80 backdrop-blur-md shadow-sm sticky top-0 z-50">
                        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                            <div className="flex justify-between items-center h-16">
                                <div className="flex items-center gap-3">
                                    <img
                                        src="/images2/image copy.png"
                                        alt="IBN Tea Logo"
                                        className="h-12 w-auto object-contain"
                                    />
                                    <span className="text-2xl font-bold" style={{ color: '#22C55E' }}>
                                        Ibn tea
                                    </span>
                                </div>

                                {/* Desktop Menu */}
                                <div className="hidden md:flex items-center gap-8">
                                    <a href="#technology" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Công nghệ</a>
                                    <a href="#journey" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Câu Chuyện Cây Chè</a>
                                    <a href="#transparency" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Minh bạch</a>
                                    <Link to="/login" className="px-6 py-2 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full hover:shadow-lg transition-all">
                                        Đăng Nhập
                                    </Link>
                                </div>

                                {/* Mobile Menu Button */}
                                <button
                                    onClick={() => setShowMobileMenu(!showMobileMenu)}
                                    className="md:hidden p-2 rounded-lg hover:bg-gray-100 transition-colors"
                                >
                                    {showMobileMenu ? (
                                        <X className="h-6 w-6 text-gray-600" />
                                    ) : (
                                        <Menu className="h-6 w-6 text-gray-600" />
                                    )}
                                </button>
                            </div>
                        </div>

                        {/* Mobile Menu Dropdown */}
                        {showMobileMenu && (
                            <div className="md:hidden border-t border-gray-100 bg-white px-4 py-4 shadow-lg">
                                <div className="flex flex-col gap-4">
                                    <a
                                        href="#technology"
                                        onClick={() => setShowMobileMenu(false)}
                                        className="text-gray-700 hover:text-green-600 font-medium py-2"
                                    >
                                        Công nghệ
                                    </a>
                                    <a
                                        href="#journey"
                                        onClick={(e) => {
                                            e.preventDefault();
                                            setShowMobileMenu(false);
                                            document.getElementById('journey')?.scrollIntoView({ behavior: 'smooth' });
                                        }}
                                        className="text-gray-700 hover:text-green-600 font-medium py-2"
                                    >
                                        Câu Chuyện Cây Chè
                                    </a>
                                    <a
                                        href="#transparency"
                                        onClick={() => setShowMobileMenu(false)}
                                        className="text-gray-700 hover:text-green-600 font-medium py-2"
                                    >
                                        Minh bạch
                                    </a>
                                    <Link
                                        to="/login"
                                        onClick={() => setShowMobileMenu(false)}
                                        className="px-6 py-2 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full hover:shadow-lg transition-all text-center"
                                    >
                                        Đăng Nhập
                                    </Link>
                                </div>
                            </div>
                        )}
                    </nav>
                )}

                {/* Hero Section - Redesigned */}
                <section className="relative overflow-hidden py-20 md:py-32">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="text-center max-w-4xl mx-auto space-y-8">
                            {/* Badge */}
                            <div className="inline-flex items-center gap-2 px-6 py-3 bg-white/90 backdrop-blur-sm rounded-full shadow-lg">
                                <Sparkles className="w-5 h-5 text-green-600" />
                                <span className="text-sm font-bold text-gray-900">Công nghệ Blockchain kiến tạo niềm tin</span>
                            </div>

                            {/* Main Heading */}
                            <h1 className="text-4xl md:text-7xl font-bold leading-tight">
                                <span className="text-gray-900 drop-shadow-lg">Minh Bạch</span>
                                <br />
                                <span className="bg-gradient-to-r from-green-600 to-emerald-700 bg-clip-text text-transparent drop-shadow-sm">
                                    Từng Búp Trà
                                </span>
                            </h1>

                            {/* Subtitle */}
                            <p className="text-xl md:text-2xl text-gray-800 drop-shadow-sm max-w-3xl mx-auto leading-relaxed font-medium">
                                Từ đồi chè hữu cơ đến tách trà trên tay bạn, mọi hành trình đều được khắc ghi vĩnh cửu trên Blockchain. Thưởng thức hương vị thuần khiết với sự an tâm tuyệt đối
                            </p>

                            {/* CTA Buttons */}
                            <div className="flex flex-wrap justify-center gap-4 pt-4">
                                <button
                                    onClick={() => setShowVerificationModal(true)}
                                    className="group px-8 py-4 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full font-bold hover:shadow-2xl transition-all flex items-center gap-3 text-lg"
                                >
                                    <QrCode className="w-6 h-6" />
                                    Quét Mã Xác Thực
                                    <ChevronRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                                </button>
                                <a
                                    href="#journey"
                                    onClick={(e) => {
                                        e.preventDefault();
                                        document.getElementById('journey')?.scrollIntoView({ behavior: 'smooth' });
                                    }}
                                    className="px-8 py-4 bg-white/90 backdrop-blur-sm border-2 border-gray-300 text-gray-900 rounded-full font-bold hover:bg-white hover:border-green-600 transition-all flex items-center gap-2 text-lg"
                                >
                                    Câu Chuyện Cây Chè
                                </a>
                            </div>

                            {/* Trust Indicators */}
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 pt-12 max-w-2xl mx-auto">
                                <div className="text-center">
                                    <div className="text-4xl font-bold text-gray-900 drop-shadow-sm mb-2">100%</div>
                                    <div className="text-sm text-gray-700 font-medium drop-shadow-sm">Minh Bạch</div>
                                </div>
                                <div className="text-center">
                                    <div className="text-4xl font-bold text-gray-900 drop-shadow-sm mb-2">24/7</div>
                                    <div className="text-sm text-gray-700 font-medium drop-shadow-sm">Đổi thành Truy Xuất 1 Chạm</div>
                                </div>
                                <div className="text-center">
                                    <div className="text-4xl font-bold text-gray-900 drop-shadow-sm mb-2">∞</div>
                                    <div className="text-sm text-gray-700 font-medium drop-shadow-sm">Không Thể Giả Mạo</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Technology Showcase Section */}
                <section className="py-20 bg-transparent" id="technology">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="text-center mb-16">
                            <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
                                Công Nghệ Blockchain
                            </h2>
                            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
                                Nền tảng xác thực nguồn gốc không thể giả mạo
                            </p>
                        </div>

                        <div className="grid md:grid-cols-3 gap-8">
                            {/* Tech Feature 1 - Liquid Glass */}
                            <div className="group relative bg-white/10 backdrop-blur-xl rounded-3xl p-8 border border-white/20 shadow-xl hover:shadow-2xl hover:bg-white/15 transition-all duration-500">
                                {/* Gradient overlay on hover */}
                                <div className="absolute inset-0 bg-gradient-to-br from-green-400/10 to-emerald-500/10 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                {/* Shine effect */}
                                <div className="absolute inset-0 bg-gradient-to-tr from-white/0 via-white/5 to-white/0 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                <div className="relative">
                                    <div className="w-16 h-16 bg-white/5 border border-white/20 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 group-hover:rotate-3 transition-all duration-500 shadow-lg backdrop-blur-md">
                                        <Box className="w-8 h-8 text-white stroke-[1.5]" />
                                    </div>
                                    <h3 className="text-2xl font-bold text-gray-900 mb-4">Cam Kết Vàng</h3>
                                    <p className="text-gray-700 leading-relaxed font-['Roboto']">
                                        Lịch sử của búp chè được ghi lại vĩnh cửu. Một khi đã lưu trên Blockchain, không ai có thể sửa đổi hay tẩy xóa sự thật.
                                    </p>
                                </div>
                            </div>

                            {/* Tech Feature 2 - Liquid Glass */}
                            <div className="group relative bg-white/10 backdrop-blur-xl rounded-3xl p-8 border border-white/20 shadow-xl hover:shadow-2xl hover:bg-white/15 transition-all duration-500">
                                {/* Gradient overlay on hover */}
                                <div className="absolute inset-0 bg-gradient-to-br from-blue-400/10 to-indigo-500/10 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                {/* Shine effect */}
                                <div className="absolute inset-0 bg-gradient-to-tr from-white/0 via-white/5 to-white/0 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                <div className="relative">
                                    <div className="w-16 h-16 bg-white/5 border border-white/20 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 group-hover:rotate-3 transition-all duration-500 shadow-lg backdrop-blur-md">
                                        <Layers className="w-8 h-8 text-white stroke-[1.5]" />
                                    </div>
                                    <h3 className="text-2xl font-bold text-gray-900 mb-4">Truy Xuất Nguồn Gốc</h3>
                                    <p className="text-gray-700 leading-relaxed font-['Roboto']">
                                        Không còn bí mật. Bạn có quyền biết chính xác chè được hái ở đâu, khi nào và bởi ai, chỉ với một cú chạm.
                                    </p>
                                </div>
                            </div>

                            {/* Tech Feature 3 - Liquid Glass */}
                            <div className="group relative bg-white/10 backdrop-blur-xl rounded-3xl p-8 border border-white/20 shadow-xl hover:shadow-2xl hover:bg-white/15 transition-all duration-500">
                                {/* Gradient overlay on hover */}
                                <div className="absolute inset-0 bg-gradient-to-br from-purple-400/10 to-pink-500/10 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                {/* Shine effect */}
                                <div className="absolute inset-0 bg-gradient-to-tr from-white/0 via-white/5 to-white/0 rounded-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                                <div className="relative">
                                    <div className="w-16 h-16 bg-white/5 border border-white/20 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 group-hover:rotate-3 transition-all duration-500 shadow-lg backdrop-blur-md">
                                        <Hexagon className="w-8 h-8 text-white stroke-[1.5]" />
                                    </div>
                                    <h3 className="text-2xl font-bold text-gray-900 mb-4">Chống Giả Tuyệt Đối</h3>
                                    <p className="text-gray-700 leading-relaxed font-['Roboto']">
                                        Mỗi hộp trà là độc nhất. Công nghệ mã hóa bảo vệ bạn hoàn toàn trước nguy cơ hàng giả, hàng kém chất lượng.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Product Journey Section */}
                <section className="py-20 bg-transparent" id="journey">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="text-center mb-16">
                            <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
                                Hành Trình Sản Phẩm
                            </h2>
                            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
                                Theo dõi từng bước đi của sản phẩm từ vườn chè đến tay bạn
                            </p>
                        </div>

                        <div className="relative">
                            {/* Tree Trunk SVG Background */}
                            <div className="absolute left-1/2 transform -translate-x-1/2 top-0 bottom-0 w-full max-w-4xl pointer-events-none hidden md:block">
                                <svg className="w-full h-full" viewBox="0 0 400 1000" preserveAspectRatio="none">
                                    <defs>
                                        <linearGradient id="treeGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                                            <stop offset="0%" stopColor="#14532d" /> {/* dark green */}
                                            <stop offset="50%" stopColor="#166534" /> {/* green-700 */}
                                            <stop offset="100%" stopColor="#14532d" />
                                        </linearGradient>
                                        <filter id="woodTexture">
                                            <feTurbulence type="fractalNoise" baseFrequency="0.5" numOctaves="3" stitchTiles="stitch" />
                                            <feColorMatrix type="saturate" values="0" />
                                            <feComponentTransfer>
                                                <feFuncR type="linear" slope="0.3" intercept="0.3" />
                                                <feFuncG type="linear" slope="0.3" intercept="0.3" />
                                                <feFuncB type="linear" slope="0.3" intercept="0.3" />
                                            </feComponentTransfer>
                                            <feComposite operator="in" in2="SourceGraphic" />
                                        </filter>
                                    </defs>

                                    {/* Main Trunk with Gradient */}
                                    <path
                                        d="M200,0 C200,100 190,200 200,300 C210,400 195,500 200,600 C205,700 190,800 200,1000"
                                        stroke="url(#treeGradient)"
                                        strokeWidth="12"
                                        fill="none"
                                        strokeLinecap="round"
                                        className="drop-shadow-xl"
                                    />

                                    {/* Branches with Leaves */}
                                    {/* Branch 1 - Left */}
                                    <g>
                                        <path d="M200,100 Q150,100 100,120" stroke="url(#treeGradient)" strokeWidth="8" fill="none" strokeLinecap="round" />
                                        <path d="M100,120 Q80,130 60,120" stroke="url(#treeGradient)" strokeWidth="6" fill="none" strokeLinecap="round" />
                                        {/* Leaves */}
                                        <path d="M120,110 Q130,100 140,110 Q130,120 120,110" fill="#4ade80" opacity="0.8" />
                                        <path d="M80,125 Q90,115 100,125 Q90,135 80,125" fill="#22c55e" opacity="0.8" />
                                    </g>

                                    {/* Branch 2 - Right */}
                                    <g>
                                        <path d="M200,350 Q250,350 300,370" stroke="url(#treeGradient)" strokeWidth="8" fill="none" strokeLinecap="round" />
                                        <path d="M300,370 Q320,380 340,370" stroke="url(#treeGradient)" strokeWidth="6" fill="none" strokeLinecap="round" />
                                        {/* Leaves */}
                                        <path d="M280,360 Q270,350 260,360 Q270,370 280,360" fill="#4ade80" opacity="0.8" />
                                        <path d="M320,375 Q310,365 300,375 Q310,385 320,375" fill="#22c55e" opacity="0.8" />
                                    </g>

                                    {/* Branch 3 - Left */}
                                    <g>
                                        <path d="M200,600 Q150,600 100,620" stroke="url(#treeGradient)" strokeWidth="8" fill="none" strokeLinecap="round" />
                                        <path d="M100,620 Q80,630 60,620" stroke="url(#treeGradient)" strokeWidth="6" fill="none" strokeLinecap="round" />
                                        {/* Leaves */}
                                        <path d="M120,610 Q130,600 140,610 Q130,620 120,610" fill="#4ade80" opacity="0.8" />
                                        <path d="M80,625 Q90,615 100,625 Q90,635 80,625" fill="#22c55e" opacity="0.8" />
                                    </g>

                                    {/* Branch 4 - Right */}
                                    <g>
                                        <path d="M200,850 Q250,850 300,870" stroke="url(#treeGradient)" strokeWidth="8" fill="none" strokeLinecap="round" />
                                        <path d="M300,870 Q320,880 340,870" stroke="url(#treeGradient)" strokeWidth="6" fill="none" strokeLinecap="round" />
                                        {/* Leaves */}
                                        <path d="M280,860 Q270,850 260,860 Q270,870 280,860" fill="#4ade80" opacity="0.8" />
                                        <path d="M320,875 Q310,865 300,875 Q310,885 320,875" fill="#22c55e" opacity="0.8" />
                                    </g>
                                </svg>
                            </div>

                            {/* Mobile Vertical Line */}
                            <div className="absolute left-1/2 transform -translate-x-1/2 top-0 bottom-0 w-1 bg-green-200 md:hidden"></div>

                            {/* Journey Steps */}
                            <div className="space-y-12 md:space-y-24 relative z-10">
                                {[
                                    {
                                        step: '01',
                                        title: 'Thu Hoạch',
                                        description: 'Trên những đồi chè hữu cơ thanh lành, từng búp non được nâng niu hái tay, chắt chiu tinh hoa đất trời. Hành trình ấy được khắc ghi tỉ mỉ từng thời khắc và tọa độ, bảo chứng cho sự thuần khiết vẹn nguyên từ nguồn cội đến chén trà thơm.',
                                        icon: Leaf,
                                        color: 'from-green-600 to-emerald-700'
                                    },
                                    {
                                        step: '02',
                                        title: 'Chế Biến',
                                        description: 'Tiếp nối tinh hoa ấy, quy trình chế biến được giám sát nghiêm ngặt tựa như một nghi thức. Mọi công đoạn chuyển mình của lá chè đều được khắc ghi vĩnh cửu trên nền tảng Blockchain, dệt nên bức tường minh bạch tuyệt đối, khẳng định niềm tin và chất lượng không thể lay chuyển.',
                                        icon: TrendingUp,
                                        color: 'from-blue-600 to-indigo-700'
                                    },
                                    {
                                        step: '03',
                                        title: 'Đóng Gói',
                                        description: 'Gói trọn tâm tình người làm trà, mỗi sản phẩm trao tay là một tuyệt tác được định danh bằng mã QR/NFC độc bản. Chiếc chìa khóa công nghệ ấy mở ra cánh cửa kết nối trực tiếp với chuỗi dữ liệu Blockchain, tái hiện hành trình minh bạch từ đồi chè sương sớm đến tách trà đượm hương.',
                                        icon: QrCode,
                                        color: 'from-purple-600 to-pink-700'
                                    },
                                    {
                                        step: '04',
                                        title: 'Xác Thực',
                                        description: 'Chỉ với một thao tác quét, người thưởng trà như bước vào chuyến du hành ngược thời gian, chiêm ngưỡng trọn vẹn vòng đời của búp chè. Sự minh bạch ấy không chỉ xác thực nguồn gốc chuẩn mực mà còn gửi trao niềm an tâm tuyệt đối, để mỗi lần nâng ly là trọn vẹn tin yêu.',
                                        icon: CheckCircle,
                                        color: 'from-amber-600 to-orange-700'
                                    },
                                ].map((item, index) => (
                                    <div key={index} className={`flex flex-col-reverse items-center gap-8 ${index % 2 === 0 ? 'md:flex-row' : 'md:flex-row-reverse'}`}>
                                        {/* Content */}
                                        <div className={`flex-1 ${index % 2 === 0 ? 'md:text-right' : 'md:text-left'}`}>
                                            <div className="relative">
                                                <motion.div
                                                    initial={{ clipPath: index % 2 === 0 ? 'inset(0 0 0 100%)' : 'inset(0 100% 0 0)' }}
                                                    whileInView={{ clipPath: 'inset(0 0 0 0)' }}
                                                    viewport={{ once: true, margin: "-50px" }}
                                                    transition={{ duration: 1.2, ease: "easeInOut", delay: index * 0.2 }}
                                                    className="bg-white/90 backdrop-blur-md rounded-2xl p-8 shadow-lg hover:shadow-2xl transition-all border-2 border-green-100 relative group"
                                                >
                                                    {/* Leaf Decoration */}
                                                    <div className={`absolute top-1/2 transform -translate-y-1/2 ${index % 2 === 0 ? '-right-3' : '-left-3'} w-6 h-6 bg-green-600 rotate-45 hidden md:block`}></div>

                                                    <div className="text-sm font-bold text-green-600 mb-2">BƯỚC {item.step}</div>
                                                    <h3 className="text-2xl font-bold text-gray-900 mb-3">{item.title}</h3>
                                                    <p className="text-gray-600 leading-loose font-['Roboto'] text-center">{item.description}</p>
                                                </motion.div>

                                                {/* Roller Effect - The "carpet roll" */}
                                                <motion.div
                                                    initial={{
                                                        left: index % 2 === 0 ? '100%' : '0%',
                                                        right: index % 2 === 0 ? '0%' : '100%',
                                                        opacity: 1
                                                    }}
                                                    whileInView={{
                                                        left: index % 2 === 0 ? '0%' : '100%',
                                                        right: index % 2 === 0 ? '100%' : '0%',
                                                        opacity: 0
                                                    }}
                                                    viewport={{ once: true, margin: "-50px" }}
                                                    transition={{ duration: 1.2, ease: "easeInOut", delay: index * 0.2 }}
                                                    className={`absolute top-0 bottom-0 w-4 bg-gradient-to-r ${index % 2 === 0 ? 'from-green-600/50 to-transparent' : 'from-transparent to-green-600/50'} z-20 pointer-events-none rounded-full blur-sm`}
                                                    style={{
                                                        [index % 2 === 0 ? 'left' : 'right']: 'auto' // Reset conflicting property
                                                    }}
                                                />
                                            </div>
                                        </div>

                                        {/* Icon Node on Tree */}
                                        <div className="relative flex-shrink-0">
                                            <div className={`w-20 h-20 bg-gradient-to-br ${item.color} rounded-full flex items-center justify-center shadow-xl z-10 relative border-4 border-white`}>
                                                <item.icon className="w-10 h-10 text-white" />
                                            </div>
                                        </div>

                                        {/* Spacer for alignment */}
                                        <div className="flex-1 hidden md:block"></div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </section>

                {/* Transparency & Trust Section */}
                <section className="py-20 bg-transparent" id="transparency">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="grid md:grid-cols-2 gap-12 items-center">
                            {/* Left Content */}
                            <div className="space-y-6">
                                <h2 className="text-4xl md:text-5xl font-bold text-white drop-shadow-lg">
                                    Minh Bạch Là
                                    <span className="block bg-gradient-to-r from-green-400 to-emerald-300 bg-clip-text text-transparent drop-shadow-md">
                                        Cam Kết Của Chúng Tôi
                                    </span>
                                </h2>
                                <p className="text-xl text-white/95 leading-relaxed font-['Roboto'] drop-shadow-md">
                                    Chúng tôi tin rằng người tiêu dùng có quyền biết chính xác nguồn gốc và chất lượng của sản phẩm họ sử dụng.
                                </p>

                                <div className="space-y-4">
                                    {[
                                        'Thông tin nguồn gốc rõ ràng',
                                        'Quy trình sản xuất minh bạch',
                                        'Chứng nhận chất lượng đầy đủ',
                                        'Truy xuất nguồn gốc 24/7'
                                    ].map((item, i) => (
                                        <div key={i} className="flex items-center gap-3">
                                            <div className="w-6 h-6 bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center flex-shrink-0 border border-white/30">
                                                <CheckCircle className="w-4 h-4 text-white" />
                                            </div>
                                            <span className="text-white font-medium drop-shadow-sm">{item}</span>
                                        </div>
                                    ))}
                                </div>

                                <button
                                    onClick={() => setShowVerificationModal(true)}
                                    className="mt-8 px-8 py-4 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full font-bold hover:shadow-xl transition-all inline-flex items-center gap-2"
                                >
                                    Thử Ngay
                                    <ChevronRight className="w-5 h-5" />
                                </button>
                            </div>

                            {/* Right Visual */}
                            <div className="relative">
                                <div className="absolute inset-0 bg-gradient-to-br from-green-400/20 to-emerald-600/20 rounded-3xl blur-3xl"></div>
                                <div className="relative bg-white/10 backdrop-blur-md rounded-3xl p-12 shadow-2xl border border-white/20">
                                    <div className="grid grid-cols-2 gap-6">
                                        {[
                                            { icon: Shield, label: 'Bảo Mật', color: 'from-green-600 to-emerald-700' },
                                            { icon: Eye, label: 'Minh Bạch', color: 'from-blue-600 to-indigo-700' },
                                            { icon: Lock, label: 'An Toàn', color: 'from-purple-600 to-pink-700' },
                                            { icon: CheckCircle, label: 'Xác Thực', color: 'from-amber-600 to-orange-700' },
                                        ].map((item, i) => (
                                            <div key={i} className="bg-white/90 backdrop-blur-md rounded-2xl p-6 shadow-lg hover:shadow-xl transition-all text-center">
                                                <div className={`w-16 h-16 bg-gradient-to-br ${item.color} rounded-2xl flex items-center justify-center mx-auto mb-4`}>
                                                    <item.icon className="w-8 h-8 text-white" />
                                                </div>
                                                <div className="text-sm font-bold text-gray-900">{item.label}</div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Final CTA Section */}
                <section className="py-20 bg-transparent">
                    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
                        <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-6">
                            Sẵn Sàng Xác Thực?
                        </h2>
                        <p className="text-xl text-gray-800 mb-8 max-w-2xl mx-auto font-medium">
                            Quét mã QR trên sản phẩm hoặc nhập mã hash để xem toàn bộ thông tin nguồn gốc
                        </p>
                        <button
                            onClick={() => setShowVerificationModal(true)}
                            className="px-10 py-5 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full font-bold text-lg hover:shadow-2xl transition-all inline-flex items-center gap-3"
                        >
                            <QrCode className="w-6 h-6" />
                            Bắt Đầu Xác Thực
                            <ChevronRight className="w-5 h-5" />
                        </button>
                    </div>
                </section>

                {/* Footer */}
                <footer className="bg-gray-900 text-white py-12">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
                        <div className="flex items-center justify-center gap-3 mb-4">
                            <img src="/images2/image copy.png" alt="IBN Tea Logo" className="h-10 w-auto" />
                            <span className="text-2xl font-bold text-green-400">Ibn tea</span>
                        </div>
                        <p className="text-gray-400 mb-4">
                            Minh bạch từng bước chân - Blockchain xác thực nguồn gốc
                        </p>
                        <div className="text-sm text-gray-500">
                            © 2024 IBN Network. All rights reserved.
                        </div>
                    </div>
                </footer>
            </div>

            {/* Verification Modal */}
            {showVerificationModal && (
                <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                    <div className="bg-white rounded-3xl max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-2xl">
                        <div className="p-8">
                            {!selectedMethod ? (
                                <>
                                    <div className="flex justify-between items-center mb-8">
                                        <h2 className="text-3xl font-bold text-gray-900">Chọn Phương Thức</h2>
                                        <button
                                            onClick={() => setShowVerificationModal(false)}
                                            className="w-10 h-10 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors"
                                        >
                                            <X className="w-5 h-5 text-gray-600" />
                                        </button>
                                    </div>

                                    <div className="space-y-4">
                                        {/* QR Method */}
                                        <button
                                            onClick={() => setSelectedMethod('qr')}
                                            className="w-full p-6 rounded-2xl border-2 border-gray-200 hover:border-green-600 hover:bg-green-50 transition-all flex items-center gap-4 group text-left"
                                        >
                                            <div className="w-14 h-14 bg-green-100 rounded-xl flex items-center justify-center group-hover:bg-green-600 transition-colors">
                                                <QrCode className="w-7 h-7 text-green-600 group-hover:text-white transition-colors" />
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="font-bold text-gray-900 text-lg mb-1">Quét Mã QR</h3>
                                                <p className="text-sm text-gray-600">Sử dụng camera để quét mã QR trên sản phẩm</p>
                                            </div>
                                            <ChevronRight className="w-6 h-6 text-gray-400 group-hover:text-green-600 transition-colors" />
                                        </button>

                                        {/* Hash Method */}
                                        <button
                                            onClick={() => setSelectedMethod('hash')}
                                            className="w-full p-6 rounded-2xl border-2 border-gray-200 hover:border-blue-600 hover:bg-blue-50 transition-all flex items-center gap-4 group text-left"
                                        >
                                            <div className="w-14 h-14 bg-blue-100 rounded-xl flex items-center justify-center group-hover:bg-blue-600 transition-colors">
                                                <Hash className="w-7 h-7 text-blue-600 group-hover:text-white transition-colors" />
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="font-bold text-gray-900 text-lg mb-1">Nhập Mã Hash</h3>
                                                <p className="text-sm text-gray-600">Nhập mã hash từ sản phẩm để xác thực</p>
                                            </div>
                                            <ChevronRight className="w-6 h-6 text-gray-400 group-hover:text-blue-600 transition-colors" />
                                        </button>

                                        {/* NFC Method */}
                                        <button
                                            onClick={() => setSelectedMethod('nfc')}
                                            className="w-full p-6 rounded-2xl border-2 border-gray-200 hover:border-purple-600 hover:bg-purple-50 transition-all flex items-center gap-4 group text-left"
                                        >
                                            <div className="w-14 h-14 bg-purple-100 rounded-xl flex items-center justify-center group-hover:bg-purple-600 transition-colors">
                                                <Radio className="w-7 h-7 text-purple-600 group-hover:text-white transition-colors" />
                                            </div>
                                            <div className="flex-1">
                                                <h3 className="font-bold text-gray-900 text-lg mb-1">Quét Thẻ NFC</h3>
                                                <p className="text-sm text-gray-600">Chạm thẻ NFC vào điện thoại để xác thực</p>
                                            </div>
                                            <ChevronRight className="w-6 h-6 text-gray-400 group-hover:text-purple-600 transition-colors" />
                                        </button>
                                    </div>
                                </>
                            ) : (
                                <>
                                    <button
                                        onClick={() => {
                                            setSelectedMethod(null)
                                            stopQRScanner()
                                        }}
                                        className="mb-6 flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
                                    >
                                        <ArrowLeft className="w-5 h-5" />
                                        <span className="font-medium">Quay lại</span>
                                    </button>

                                    {selectedMethod === 'qr' && (
                                        <div className="space-y-6">
                                            <div className="text-center">
                                                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                    <QrCode className="w-8 h-8 text-green-600" />
                                                </div>
                                                <h2 className="text-2xl font-bold text-gray-900 mb-2">Quét Mã QR</h2>
                                                <p className="text-gray-600">Đưa camera vào mã QR trên sản phẩm</p>
                                            </div>

                                            {cameraError ? (
                                                <div className="p-6 bg-red-50 border border-red-200 rounded-xl">
                                                    <p className="text-red-600 text-center">{cameraError}</p>
                                                </div>
                                            ) : (
                                                <div id="qr-reader-container" className="rounded-xl overflow-hidden"></div>
                                            )}

                                            <div className="flex gap-3">
                                                <button
                                                    onClick={() => {
                                                        if (isScanning) {
                                                            stopQRScanner()
                                                        } else {
                                                            startQRScanner()
                                                        }
                                                    }}
                                                    className={`flex-1 px-6 py-3 rounded-xl font-semibold transition-all flex items-center justify-center gap-2 ${isScanning
                                                        ? 'bg-red-500 text-white hover:bg-red-600'
                                                        : 'bg-green-500 text-white hover:bg-green-600'
                                                        }`}
                                                >
                                                    {isScanning ? (
                                                        <>
                                                            <X className="w-5 h-5" />
                                                            Tắt
                                                        </>
                                                    ) : (
                                                        <>
                                                            <Camera className="w-5 h-5" />
                                                            Quét
                                                        </>
                                                    )}
                                                </button>
                                            </div>
                                        </div>
                                    )}

                                    {selectedMethod === 'hash' && (
                                        <div className="space-y-6">
                                            <div className="text-center">
                                                <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                    <Hash className="w-8 h-8 text-blue-600" />
                                                </div>
                                                <h2 className="text-2xl font-bold text-gray-900 mb-2">Nhập Mã Hash</h2>
                                                <p className="text-gray-600">Nhập mã hash từ sản phẩm</p>
                                            </div>

                                            <input
                                                type="text"
                                                value={hash}
                                                onChange={(e) => setHash(e.target.value)}
                                                placeholder="Nhập mã hash..."
                                                className="w-full px-6 py-4 border-2 border-gray-200 rounded-xl focus:border-blue-500 focus:outline-none text-lg"
                                            />

                                            <button
                                                onClick={() => {
                                                    if (hash.trim()) {
                                                        navigate(`/verify/hash?hash=${encodeURIComponent(hash.trim())}`)
                                                        setShowVerificationModal(false)
                                                    }
                                                }}
                                                disabled={!hash.trim()}
                                                className="w-full px-6 py-4 bg-gradient-to-r from-blue-600 to-indigo-700 text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                                            >
                                                Xác Thực
                                            </button>
                                        </div>
                                    )}

                                    {selectedMethod === 'nfc' && (
                                        <div className="space-y-6">
                                            <div className="text-center">
                                                <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                    <Radio className="w-8 h-8 text-purple-600" />
                                                </div>
                                                <h2 className="text-2xl font-bold text-gray-900 mb-2">Quét Thẻ NFC</h2>
                                                <p className="text-gray-600">Chạm thẻ NFC vào mặt sau điện thoại</p>
                                            </div>

                                            <button
                                                onClick={startNfcScan}
                                                className="w-full px-6 py-4 bg-gradient-to-r from-purple-600 to-pink-700 text-white rounded-xl font-semibold hover:shadow-lg transition-all"
                                            >
                                                Bắt Đầu Quét
                                            </button>
                                        </div>
                                    )}
                                </>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
