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
import { Shield, Leaf, Award, QrCode, ChevronRight, CheckCircle, Hash, Radio, ArrowLeft, Camera, X } from 'lucide-react'
import { HomeHeader } from '@shared/components/layout/HomeHeader'
import { authService } from '@features/authentication/services/authService'
import { useEffect, useState, useRef } from 'react'
import { Html5Qrcode } from 'html5-qrcode'
// NOTE: Removed transactionExplorerService import - hash verification now queries from blockchain network, not database

type VerificationMethod = 'qr' | 'hash' | 'nfc' | null

export function TeaShopHomepage() {
    const isAuthenticated = authService.isAuthenticated()
    const navigate = useNavigate()
    const [showVerificationModal, setShowVerificationModal] = useState(false)
    const [selectedMethod, setSelectedMethod] = useState<VerificationMethod>(null)
    const [qrCode, setQrCode] = useState('')
    const [hash, setHash] = useState('')
    const [nfcTag, setNfcTag] = useState('')
    const [isScanning, setIsScanning] = useState(false)
    const [cameraError, setCameraError] = useState<string | null>(null)
    const qrCodeScannerRef = useRef<Html5Qrcode | null>(null)
    const scannerContainerRef = useRef<HTMLDivElement>(null)

    // NOTE: Removed fetching transactions from database
    // Hash verification now queries directly from blockchain network, not from database

    // QR Code Scanner functions
    const startQRScanner = async () => {
        const containerId = 'qr-reader-container'
        
        // Wait for container to be rendered
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
                { facingMode: 'environment' }, // Use back camera on mobile
                {
                    fps: 10,
                    qrbox: { width: 250, height: 250 },
                },
                (decodedText) => {
                    // QR code scanned successfully
                    // Extract package ID from URL if it's a verification URL
                    let packageId = decodedText
                    if (decodedText.includes('/verify/packages/')) {
                        packageId = decodedText.split('/verify/packages/')[1]?.split('?')[0] || decodedText
                    } else if (decodedText.includes('hash=')) {
                        packageId = decodedText.split('hash=')[1]?.split('&')[0] || decodedText
                    }
                    setQrCode(packageId)
                    stopQRScanner()
                    // Auto-submit after a short delay
                    setTimeout(() => {
                        if (packageId.trim()) {
                            navigate(`/verify/packages/${packageId.trim()}`)
                            setShowVerificationModal(false)
                        }
                    }, 500)
                },
                (_errorMessage) => {
                    // Ignore scanning errors (they're frequent during scanning)
                }
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

    // Cleanup scanner when component unmounts or method changes
    useEffect(() => {
        return () => {
            if (qrCodeScannerRef.current) {
                stopQRScanner()
            }
        }
    }, [])

    // Stop scanner when method changes or modal closes
    useEffect(() => {
        if (!showVerificationModal || selectedMethod !== 'qr') {
            stopQRScanner()
        }
    }, [showVerificationModal, selectedMethod])

    // Parallax scroll effect - ảnh di chuyển từ trên xuống khi scroll
    useEffect(() => {
        const handleScroll = () => {
            const scrolled = window.pageYOffset
            const parallaxElement = document.getElementById('parallax-background')
            if (parallaxElement) {
                // Di chuyển ảnh với tốc độ chậm hơn scroll (parallax effect)
                const yPos = -(scrolled * 0.5)
                parallaxElement.style.transform = `translate3d(0, ${yPos}px, 0)`
            }
        }

        // Update background height khi content thay đổi
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
        
        // Update khi component mount
        setTimeout(updateBackgroundHeight, 100)
        
        return () => {
            window.removeEventListener('scroll', handleScroll)
            window.removeEventListener('resize', updateBackgroundHeight)
        }
    }, [])

    return (
        <div className="min-h-screen relative">
            {/* Background Image với Parallax Effect - kéo dài đến hết trang */}
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
            {/* Use HomeHeader when authenticated, otherwise use simple navigation */}
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
                            <div className="hidden md:flex items-center gap-8">
                                <a href="#products" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Sản phẩm</a>
                                <a href="#process" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Quy trình</a>
                                <a href="#uses" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Công dụng</a>
                                <a href="#about" className="text-gray-700 hover:text-green-600 transition-colors font-medium">Giới thiệu</a>
                                <Link to="/login" className="px-6 py-2 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full hover:shadow-lg transition-all">
                                    Đăng Nhập
                                </Link>
                            </div>
                        </div>
                    </div>
                </nav>
            )}

            {/* Hero Section */}
            <section className="relative overflow-hidden">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 md:py-32">
                    <div className="grid md:grid-cols-2 gap-12 items-center">
                        {/* Left Content */}
                        <div className="space-y-8">
                            <div className="inline-flex items-center gap-2 px-4 py-2 bg-green-100 rounded-full text-green-800 text-sm font-semibold">
                                <Shield className="w-4 h-4" />
                                Xác Thực Blockchain
                            </div>

                            <h1 className="text-5xl md:text-6xl font-bold text-gray-900 leading-tight">
                                Chè Cao Cấp
                                <span className="block bg-gradient-to-r from-green-600 to-emerald-700 bg-clip-text text-transparent">
                                    Nguồn Gốc Rõ Ràng
                                </span>
                            </h1>

                            <p className="text-xl text-gray-600 leading-relaxed">
                                Mỗi sản phẩm đều được xác thực bằng công nghệ blockchain,
                                đảm bảo chất lượng và nguồn gốc từ vườn chè đến tay bạn.
                            </p>

                            <div className="flex flex-wrap gap-4">
                                <a
                                    href="#products"
                                    className="px-8 py-4 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-full font-semibold hover:shadow-xl transition-all flex items-center gap-2 group"
                                >
                                    Khám Phá Sản Phẩm
                                    <ChevronRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                                </a>
                                <button
                                    onClick={() => setShowVerificationModal(true)}
                                    className="px-8 py-4 bg-white border-2 border-green-600 text-green-600 rounded-full font-semibold hover:bg-green-50 transition-all flex items-center gap-2"
                                >
                                    <QrCode className="w-5 h-5" />
                                    Xác Thực Ngay
                                </button>
                            </div>

                            {/* Stats */}
                            <div className="grid grid-cols-3 gap-6 pt-8 border-t border-gray-200">
                                <div>
                                    <div className="text-3xl font-bold text-green-600">100%</div>
                                    <div className="text-sm text-gray-600">Hữu Cơ</div>
                                </div>
                                <div>
                                    <div className="text-3xl font-bold text-green-600">50+</div>
                                    <div className="text-sm text-gray-600">Sản Phẩm</div>
                                </div>
                                <div>
                                    <div className="text-3xl font-bold text-green-600">10K+</div>
                                    <div className="text-sm text-gray-600">Khách Hàng</div>
                                </div>
                            </div>
                        </div>

                        {/* Right Image */}
                        <div className="relative">
                            <div className="absolute inset-0 bg-gradient-to-br from-green-400/20 to-emerald-600/20 rounded-3xl blur-3xl"></div>
                            <div className="relative bg-gradient-to-br from-green-100 to-emerald-100 rounded-3xl p-8 shadow-2xl">
                                <div className="aspect-square bg-white rounded-2xl shadow-lg flex items-center justify-center">
                                    <Leaf className="w-32 h-32 text-green-600" />
                                </div>
                                {/* Floating Badge */}
                                <div className="absolute -top-4 -right-4 bg-white rounded-2xl shadow-xl p-4 flex items-center gap-3">
                                    <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
                                        <Shield className="w-6 h-6 text-green-600" />
                                    </div>
                                    <div>
                                        <div className="text-sm font-semibold text-gray-900">Blockchain</div>
                                        <div className="text-xs text-gray-600">Verified</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Features Section */}
            <section className="py-20 bg-transparent" id="verification">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl font-bold text-white mb-4 drop-shadow-lg">
                            Xác Thực Blockchain
                        </h2>
                        <p className="text-xl text-white/90 max-w-2xl mx-auto drop-shadow-md">
                            Công nghệ blockchain đảm bảo mỗi sản phẩm đều có nguồn gốc rõ ràng
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        {/* Feature 1 */}
                        <div className="group p-8 rounded-2xl transition-all">
                            <div className="w-16 h-16 bg-gradient-to-br from-green-600 to-emerald-700 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
                                <Shield className="w-8 h-8 text-white" />
                            </div>
                            <h3 className="text-2xl font-bold text-white mb-4 drop-shadow-lg">Xác Thực Blockchain</h3>
                            <p className="text-white/90 leading-relaxed drop-shadow-md">
                                Mỗi sản phẩm được ghi nhận trên blockchain, không thể giả mạo hay thay đổi.
                            </p>
                        </div>

                        {/* Feature 2 */}
                        <div className="group p-8 rounded-2xl transition-all">
                            <div className="w-16 h-16 bg-gradient-to-br from-amber-600 to-orange-700 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
                                <QrCode className="w-8 h-8 text-white" />
                            </div>
                            <h3 className="text-2xl font-bold text-white mb-4 drop-shadow-lg">Quét QR Dễ Dàng</h3>
                            <p className="text-white/90 leading-relaxed drop-shadow-md">
                                Chỉ cần quét mã QR trên bao bì để xem toàn bộ hành trình của sản phẩm.
                            </p>
                        </div>

                        {/* Feature 3 */}
                        <div className="group p-8 rounded-2xl transition-all">
                            <div className="w-16 h-16 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
                                <Award className="w-8 h-8 text-white" />
                            </div>
                            <h3 className="text-2xl font-bold text-white mb-4 drop-shadow-lg">Chất Lượng Đảm Bảo</h3>
                            <p className="text-white/90 leading-relaxed drop-shadow-md">
                                Chứng nhận hữu cơ, kiểm định chất lượng nghiêm ngặt từ vườn chè.
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            {/* Products Section */}
            <section className="py-20 bg-transparent" id="products">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl font-bold text-white mb-4 drop-shadow-lg">
                            Sản Phẩm Nổi Bật
                        </h2>
                        <p className="text-xl text-white/90 drop-shadow-md">
                            Khám phá bộ sưu tập chè cao cấp của chúng tôi
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        {[
                            { name: 'Chè Xanh Thái Nguyên', price: '250.000đ', badge: 'Bán Chạy' },
                            { name: 'Chè Shan Tuyết Cổ Thụ', price: '850.000đ', badge: 'Premium' },
                            { name: 'Chè Ô Long Đài Loan', price: '450.000đ', badge: 'Mới' },
                        ].map((product, i) => (
                            <div key={i} className="group bg-white rounded-2xl overflow-hidden shadow-lg hover:shadow-2xl transition-all">
                                <div className="relative aspect-square bg-gradient-to-br from-green-100 to-emerald-100 flex items-center justify-center">
                                    <Leaf className="w-24 h-24 text-green-600" />
                                    <div className="absolute top-4 right-4 px-3 py-1 bg-green-600 text-white text-sm font-semibold rounded-full">
                                        {product.badge}
                                    </div>
                                </div>
                                <div className="p-6">
                                    <h3 className="text-xl font-bold text-gray-900 mb-2">{product.name}</h3>
                                    <div className="flex items-center justify-between mb-4">
                                        <span className="text-2xl font-bold text-green-600">{product.price}</span>
                                        <div className="flex items-center gap-1 text-sm text-gray-600">
                                            <Shield className="w-4 h-4" />
                                            Verified
                                        </div>
                                    </div>
                                    <button className="w-full py-3 bg-gradient-to-r from-green-600 to-emerald-700 text-white rounded-xl font-semibold hover:shadow-lg transition-all">
                                        Xem Chi Tiết
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* Uses Section */}
            <section className="py-20 bg-transparent" id="uses">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl font-bold text-white mb-4 drop-shadow-lg">
                            Công Dụng Của Chè
                        </h2>
                        <p className="text-xl text-white/90 drop-shadow-md">
                            Những lợi ích tuyệt vời cho sức khỏe
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        {[
                            { title: 'Chống Oxy Hóa', desc: 'Chứa nhiều chất chống oxy hóa giúp bảo vệ tế bào khỏi tổn thương', icon: Shield },
                            { title: 'Tăng Cường Miễn Dịch', desc: 'Hỗ trợ hệ miễn dịch, giúp cơ thể khỏe mạnh hơn', icon: Award },
                            { title: 'Giảm Căng Thẳng', desc: 'Giúp thư giãn tinh thần, giảm stress hiệu quả', icon: Leaf },
                        ].map((use, i) => (
                            <div key={i} className="p-8 rounded-2xl transition-all">
                                <div className="w-16 h-16 bg-gradient-to-br from-green-600 to-emerald-700 rounded-2xl flex items-center justify-center mb-6">
                                    <use.icon className="w-8 h-8 text-white" />
                                </div>
                                <h3 className="text-2xl font-bold text-white mb-4 drop-shadow-lg">{use.title}</h3>
                                <p className="text-white/90 leading-relaxed drop-shadow-md">{use.desc}</p>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* How It Works - Quy Trình */}
            <section className="py-20 bg-transparent" id="process">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl font-bold text-white mb-4 drop-shadow-lg">
                            Quy Trình Minh Bạch
                        </h2>
                        <p className="text-xl text-white/90 drop-shadow-md">
                            Từ vườn chè đến tay bạn - mọi bước đều được ghi nhận
                        </p>
                    </div>

                    <div className="grid md:grid-cols-4 gap-8">
                        {[
                            { icon: Leaf, title: 'Thu Hoạch', desc: 'Từ vườn chè hữu cơ' },
                            { icon: CheckCircle, title: 'Kiểm Định', desc: 'Kiểm tra chất lượng' },
                            { icon: Shield, title: 'Blockchain', desc: 'Ghi nhận trên blockchain' },
                            { icon: QrCode, title: 'Xác Thực', desc: 'Quét QR để kiểm tra' },
                        ].map((step, i) => (
                            <div key={i} className="text-center">
                                <div className="relative mb-6">
                                    <div className="w-20 h-20 bg-gradient-to-br from-green-600 to-emerald-700 rounded-full flex items-center justify-center mx-auto">
                                        <step.icon className="w-10 h-10 text-white" />
                                    </div>
                                    {i < 3 && (
                                        <div className="hidden md:block absolute top-10 left-[60%] w-full h-0.5 bg-gradient-to-r from-green-600 to-emerald-700"></div>
                                    )}
                                </div>
                                <h3 className="text-xl font-bold text-white mb-2 drop-shadow-lg">{step.title}</h3>
                                <p className="text-white/90 drop-shadow-md">{step.desc}</p>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* About Section - Giới thiệu */}
            <section className="py-20 bg-transparent" id="about">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl font-bold text-white mb-4 drop-shadow-lg">
                            Về Chúng Tôi
                        </h2>
                        <p className="text-xl text-white/90 max-w-3xl mx-auto drop-shadow-md">
                            IBN Tea - Đối tác tin cậy của bạn trong việc cung cấp chè cao cấp với công nghệ blockchain tiên tiến
                        </p>
                    </div>

                    <div className="grid md:grid-cols-2 gap-12 items-center">
                        <div className="space-y-6">
                            <div className="space-y-4">
                                <h3 className="text-2xl font-bold text-white drop-shadow-lg">Sứ Mệnh Của Chúng Tôi</h3>
                                <p className="text-white/90 leading-relaxed drop-shadow-md">
                                    Chúng tôi cam kết mang đến những sản phẩm chè chất lượng cao với nguồn gốc rõ ràng, 
                                    được xác thực bằng công nghệ blockchain Hyperledger Fabric. Mỗi sản phẩm đều được 
                                    truy xuất từ vườn chè đến tay người tiêu dùng.
                                </p>
                            </div>
                            <div className="space-y-4">
                                <h3 className="text-2xl font-bold text-white drop-shadow-lg">Giá Trị Cốt Lõi</h3>
                                <ul className="space-y-3 text-white/90 drop-shadow-md">
                                    <li className="flex items-start gap-3">
                                        <CheckCircle className="w-6 h-6 text-green-400 mt-0.5 flex-shrink-0" />
                                        <span>Minh bạch và trung thực trong mọi giao dịch</span>
                                    </li>
                                    <li className="flex items-start gap-3">
                                        <CheckCircle className="w-6 h-6 text-green-400 mt-0.5 flex-shrink-0" />
                                        <span>Chất lượng sản phẩm được đảm bảo 100%</span>
                                    </li>
                                    <li className="flex items-start gap-3">
                                        <CheckCircle className="w-6 h-6 text-green-400 mt-0.5 flex-shrink-0" />
                                        <span>Ứng dụng công nghệ blockchain tiên tiến</span>
                                    </li>
                                    <li className="flex items-start gap-3">
                                        <CheckCircle className="w-6 h-6 text-green-400 mt-0.5 flex-shrink-0" />
                                        <span>Cam kết phát triển bền vững</span>
                                    </li>
                                </ul>
                            </div>
                        </div>
                        <div className="relative">
                            <div className="absolute inset-0 bg-gradient-to-br from-green-400/20 to-emerald-600/20 rounded-3xl blur-3xl"></div>
                            <div className="relative bg-gradient-to-br from-green-100/50 to-emerald-100/50 backdrop-blur-md rounded-3xl p-12 shadow-2xl border border-white/30">
                                <div className="aspect-square bg-white/20 backdrop-blur-md rounded-2xl shadow-lg flex items-center justify-center border border-white/30">
                                    <div className="text-center space-y-4">
                                        <Leaf className="w-24 h-24 text-green-400 mx-auto drop-shadow-lg" />
                                        <div className="space-y-2">
                                            <div className="text-3xl font-bold text-white drop-shadow-lg">IBN Tea</div>
                                            <div className="text-lg text-white/90 drop-shadow-md">Blockchain Verified</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* CTA Section */}
            <section className="py-20 bg-gradient-to-r from-green-600/90 to-emerald-700/90 backdrop-blur-md">
                <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
                    <h2 className="text-4xl font-bold text-white mb-6">
                        Sẵn Sàng Trải Nghiệm?
                    </h2>
                    <p className="text-xl text-white/90 mb-8">
                        Quét mã QR trên sản phẩm để xác thực nguồn gốc ngay bây giờ
                    </p>
                           <div className="flex flex-wrap justify-center gap-4">
                               <button
                                   onClick={() => setShowVerificationModal(true)}
                                   className="px-8 py-4 bg-white text-green-600 rounded-full font-semibold hover:shadow-xl transition-all flex items-center gap-2"
                               >
                                   <QrCode className="w-5 h-5" />
                                   Xác Thực Sản Phẩm
                               </button>
                               <a
                                   href="#products"
                                   className="px-8 py-4 bg-white/10 backdrop-blur-sm border-2 border-white text-white rounded-full font-semibold hover:bg-white/20 transition-all"
                               >
                                   Xem Sản Phẩm
                               </a>
                           </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="bg-gray-900/80 backdrop-blur-md text-white py-12">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="grid md:grid-cols-4 gap-8 mb-8">
                        <div>
                            <div className="flex items-center gap-2 mb-4">
                                <Leaf className="w-6 h-6 text-green-400" />
                                <span className="text-xl font-bold">IBN Tea</span>
                            </div>
                            <p className="text-gray-400">
                                Chè cao cấp với công nghệ blockchain
                            </p>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Sản Phẩm</h4>
                            <ul className="space-y-2 text-gray-400">
                                <li><a href="#" className="hover:text-white transition-colors">Chè Xanh</a></li>
                                <li><a href="#" className="hover:text-white transition-colors">Chè Ô Long</a></li>
                                <li><a href="#" className="hover:text-white transition-colors">Chè Đen</a></li>
                            </ul>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Công Ty</h4>
                            <ul className="space-y-2 text-gray-400">
                                <li><a href="#" className="hover:text-white transition-colors">Về Chúng Tôi</a></li>
                                <li><a href="#" className="hover:text-white transition-colors">Liên Hệ</a></li>
                                <li><Link to="/login" className="hover:text-white transition-colors">Đăng Nhập</Link></li>
                            </ul>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Hỗ Trợ</h4>
                            <ul className="space-y-2 text-gray-400">
                                <li><a href="#" className="hover:text-white transition-colors">FAQ</a></li>
                                <li><a href="#" className="hover:text-white transition-colors">Chính Sách</a></li>
                                <li><a href="#" className="hover:text-white transition-colors">Bảo Mật</a></li>
                            </ul>
                        </div>
                    </div>
                    <div className="border-t border-gray-800 pt-8 text-center text-gray-400">
                        <p>&copy; 2024 IBN Tea. Powered by Blockchain Technology.</p>
                    </div>
                </div>
            </footer>

            {/* Verification Modal */}
            {showVerificationModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
                    <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-8 relative">
                        {/* Close button */}
                        <button
                            onClick={() => {
                                setShowVerificationModal(false)
                                setSelectedMethod(null)
                                setQrCode('')
                                setHash('')
                                setNfcTag('')
                            }}
                            className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors"
                        >
                            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>

                        {!selectedMethod ? (
                            <>
                                {/* Header */}
                                <div className="text-center mb-8">
                                    <div className="w-16 h-16 bg-gradient-to-br from-green-600 to-emerald-700 rounded-full flex items-center justify-center mx-auto mb-4">
                                        <Shield className="w-8 h-8 text-white" />
                                    </div>
                                    <h2 className="text-2xl font-bold text-gray-900 mb-2">Chọn Phương Thức Xác Thực</h2>
                                    <p className="text-gray-600">Vui lòng chọn cách bạn muốn xác thực sản phẩm</p>
                                </div>

                                {/* Options */}
                                <div className="space-y-4">
                                    {/* Option 1: QR Code */}
                                    <button
                                        onClick={() => setSelectedMethod('qr')}
                                        className="w-full p-6 rounded-xl border-2 border-gray-200 hover:border-green-600 hover:bg-green-50 transition-all flex items-center gap-4 group"
                                    >
                                        <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center group-hover:bg-green-600 transition-colors">
                                            <QrCode className="w-6 h-6 text-green-600 group-hover:text-white transition-colors" />
                                        </div>
                                        <div className="flex-1 text-left">
                                            <h3 className="font-semibold text-gray-900 mb-1">Mã QR</h3>
                                            <p className="text-sm text-gray-600">Quét mã QR trên bao bì sản phẩm</p>
                                        </div>
                                        <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-green-600 transition-colors" />
                                    </button>

                                    {/* Option 2: Hash */}
                                    <button
                                        onClick={() => setSelectedMethod('hash')}
                                        className="w-full p-6 rounded-xl border-2 border-gray-200 hover:border-blue-600 hover:bg-blue-50 transition-all flex items-center gap-4 group"
                                    >
                                        <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center group-hover:bg-blue-600 transition-colors">
                                            <Hash className="w-6 h-6 text-blue-600 group-hover:text-white transition-colors" />
                                        </div>
                                        <div className="flex-1 text-left">
                                            <h3 className="font-semibold text-gray-900 mb-1">Hash</h3>
                                            <p className="text-sm text-gray-600">Nhập hash hoặc transaction ID để xác thực</p>
                                        </div>
                                        <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 transition-colors" />
                                    </button>

                                    {/* Option 3: NFC */}
                                    <button
                                        onClick={() => setSelectedMethod('nfc')}
                                        className="w-full p-6 rounded-xl border-2 border-gray-200 hover:border-purple-600 hover:bg-purple-50 transition-all flex items-center gap-4 group"
                                    >
                                        <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center group-hover:bg-purple-600 transition-colors">
                                            <Radio className="w-6 h-6 text-purple-600 group-hover:text-white transition-colors" />
                                        </div>
                                        <div className="flex-1 text-left">
                                            <h3 className="font-semibold text-gray-900 mb-1">Xác Thực NFC</h3>
                                            <p className="text-sm text-gray-600">Quét thẻ NFC trên sản phẩm</p>
                                        </div>
                                        <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-purple-600 transition-colors" />
                                    </button>
                                </div>

                                {/* Cancel button */}
                                <button
                                    onClick={() => {
                                        setShowVerificationModal(false)
                                        setSelectedMethod(null)
                                        setQrCode('')
                                        setHash('')
                                        setNfcTag('')
                                    }}
                                    className="w-full mt-6 px-6 py-3 rounded-xl border-2 border-gray-300 text-gray-700 font-semibold hover:bg-gray-50 transition-colors"
                                >
                                    Hủy
                                </button>
                            </>
                        ) : (
                            <>
                                {/* Back button */}
                                <button
                                    onClick={() => {
                                        setSelectedMethod(null)
                                        setQrCode('')
                                        setHash('')
                                        setNfcTag('')
                                    }}
                                    className="mb-6 flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
                                >
                                    <ArrowLeft className="w-5 h-5" />
                                    <span className="font-medium">Quay lại</span>
                                </button>

                                {/* Form Header */}
                                <div className="text-center mb-6">
                                    {selectedMethod === 'qr' && (
                                        <>
                                            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                <QrCode className="w-8 h-8 text-green-600" />
                                            </div>
                                            <h2 className="text-2xl font-bold text-gray-900 mb-2">Xác Thực Bằng Mã QR</h2>
                                            <p className="text-gray-600">Nhập mã QR hoặc Package ID</p>
                                        </>
                                    )}
                                    {selectedMethod === 'hash' && (
                                        <>
                                            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                <Hash className="w-8 h-8 text-blue-600" />
                                            </div>
                                            <h2 className="text-2xl font-bold text-gray-900 mb-2">Xác Thực Bằng Hash</h2>
                                            <p className="text-gray-600">Nhập hash hoặc transaction ID</p>
                                        </>
                                    )}
                                    {selectedMethod === 'nfc' && (
                                        <>
                                            <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                                <Radio className="w-8 h-8 text-purple-600" />
                                            </div>
                                            <h2 className="text-2xl font-bold text-gray-900 mb-2">Xác Thực Bằng NFC</h2>
                                            <p className="text-gray-600">Nhập mã NFC tag</p>
                                        </>
                                    )}
                                </div>

                                {/* Input Form */}
                                <div className="space-y-4">
                                    {selectedMethod === 'qr' && (
                                        <div className="space-y-3">
                                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                                Mã QR / Package ID
                                            </label>
                                            <div className="flex gap-2">
                                                <input
                                                    type="text"
                                                    value={qrCode}
                                                    onChange={(e) => setQrCode(e.target.value)}
                                                    placeholder="Nhập mã QR hoặc Package ID"
                                                    className="flex-1 px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-green-600 focus:outline-none transition-colors text-gray-900"
                                                    autoFocus
                                                />
                                                <button
                                                    type="button"
                                                    onClick={() => {
                                                        if (isScanning) {
                                                            stopQRScanner()
                                                        } else {
                                                            startQRScanner()
                                                        }
                                                    }}
                                                    className={`px-4 py-3 rounded-xl border-2 transition-colors flex items-center gap-2 ${
                                                        isScanning
                                                            ? 'bg-red-500 text-white border-red-600 hover:bg-red-600'
                                                            : 'bg-green-500 text-white border-green-600 hover:bg-green-600'
                                                    }`}
                                                >
                                                    {isScanning ? (
                                                        <>
                                                            <X className="w-5 h-5" />
                                                            <span className="hidden sm:inline">Tắt</span>
                                                        </>
                                                    ) : (
                                                        <>
                                                            <Camera className="w-5 h-5" />
                                                            <span className="hidden sm:inline">Quét</span>
                                                        </>
                                                    )}
                                                </button>
                                            </div>

                                            {/* Camera Scanner */}
                                            {isScanning && (
                                                <div className="relative">
                                                    <div
                                                        id="qr-reader-container"
                                                        ref={scannerContainerRef}
                                                        className="w-full rounded-xl overflow-hidden border-2 border-green-500"
                                                        style={{ minHeight: '300px' }}
                                                    />
                                                    <div className="absolute top-2 right-2 bg-black/70 text-white px-3 py-1 rounded-lg text-sm z-10">
                                                        Đưa QR code vào khung
                                                    </div>
                                                </div>
                                            )}

                                            {cameraError && (
                                                <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-700">
                                                    {cameraError}
                                                </div>
                                            )}

                                            {!isScanning && (
                                                <p className="text-xs text-gray-500">
                                                    Nhấn nút "Quét" để bật camera và quét QR code tự động
                                                </p>
                                            )}
                                        </div>
                                    )}

                                    {selectedMethod === 'hash' && (
                                        <div className="space-y-3">
                                            <div>
                                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                                    Hash / Transaction ID
                                                </label>
                                                <input
                                                    type="text"
                                                    value={hash}
                                                    onChange={(e) => setHash(e.target.value)}
                                                    placeholder="Nhập hash hoặc transaction ID"
                                                    className="w-full px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-blue-600 focus:outline-none transition-colors text-gray-900"
                                                    autoFocus
                                                />
                                            </div>
                                            
                                            <p className="text-xs text-gray-500 mt-2">
                                                Nhập transaction ID (hash) từ blockchain network để xác thực. Hệ thống sẽ query trực tiếp từ blockchain, không từ database.
                                            </p>
                                        </div>
                                    )}

                                    {selectedMethod === 'nfc' && (
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                                Mã NFC Tag
                                            </label>
                                            <input
                                                type="text"
                                                value={nfcTag}
                                                onChange={(e) => setNfcTag(e.target.value)}
                                                placeholder="Nhập mã NFC tag"
                                                className="w-full px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-purple-600 focus:outline-none transition-colors text-gray-900"
                                                autoFocus
                                            />
                                        </div>
                                    )}

                                    {/* Submit button */}
                                    <button
                                        onClick={() => {
                                            if (selectedMethod === 'qr' && qrCode.trim()) {
                                                navigate(`/verify/packages/${qrCode.trim()}`)
                                                setShowVerificationModal(false)
                                            } else if (selectedMethod === 'hash' && hash.trim()) {
                                                navigate(`/verify/hash?hash=${encodeURIComponent(hash.trim())}`)
                                                setShowVerificationModal(false)
                                            } else if (selectedMethod === 'nfc' && nfcTag.trim()) {
                                                navigate(`/verify/nfc?tag=${encodeURIComponent(nfcTag.trim())}`)
                                                setShowVerificationModal(false)
                                            }
                                        }}
                                        disabled={
                                            (selectedMethod === 'qr' && !qrCode.trim()) ||
                                            (selectedMethod === 'hash' && !hash.trim()) ||
                                            (selectedMethod === 'nfc' && !nfcTag.trim())
                                        }
                                        className={`w-full px-6 py-3 rounded-xl font-semibold text-white transition-all ${
                                            selectedMethod === 'qr'
                                                ? 'bg-gradient-to-r from-green-600 to-emerald-700 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed'
                                                : selectedMethod === 'hash'
                                                ? 'bg-gradient-to-r from-blue-600 to-indigo-700 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed'
                                                : 'bg-gradient-to-r from-purple-600 to-purple-700 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed'
                                        }`}
                                    >
                                        Xác Thực
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                </div>
            )}
            </div>
        </div>
    )
}
