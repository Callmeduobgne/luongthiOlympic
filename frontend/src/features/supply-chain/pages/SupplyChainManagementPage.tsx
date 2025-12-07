import { useState } from 'react'
import { Package, QrCode, Radio } from 'lucide-react'
import { BatchListPage } from './BatchListPage'
import { QRCodeGeneratorPage } from '@features/dashboard/pages/QRCodeGeneratorPage'
import { NFCManagerPage } from '@features/dashboard/pages/NFCManagerPage'

type TabType = 'batches' | 'qr-code' | 'nfc'

export const SupplyChainManagementPage = () => {
    const [activeTab, setActiveTab] = useState<TabType>('batches')

    const tabs = [
        {
            id: 'batches' as TabType,
            label: 'Batches',
            icon: Package,
            description: 'Quản lý lô trà',
        },
        {
            id: 'qr-code' as TabType,
            label: 'QR Code',
            icon: QrCode,
            description: 'Sinh mã QR',
        },
        {
            id: 'nfc' as TabType,
            label: 'NFC Tags',
            icon: Radio,
            description: 'Gán thẻ NFC',
        },
    ]

    return (
        <div className="space-y-6 text-white">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold">Supply Chain Management</h1>
                <p className="mt-1 text-sm text-gray-400">
                    Quản lý truy xuất nguồn gốc sản phẩm trà
                </p>
            </div>

            {/* Tabs */}
            <div className="border-b border-white/10">
                <nav className="flex space-x-4">
                    {tabs.map((tab) => {
                        const Icon = tab.icon
                        const isActive = activeTab === tab.id

                        return (
                            <button
                                key={tab.id}
                                onClick={() => setActiveTab(tab.id)}
                                className={`py-3 px-5 rounded-t-2xl text-sm font-semibold transition-all ${isActive
                                        ? 'bg-white text-black shadow-lg'
                                        : 'bg-white/5 text-gray-300 border border-white/10 border-b-0 hover:bg-white/10'
                                    }`}
                            >
                                <div className="flex items-center gap-2">
                                    <Icon className="w-4 h-4" />
                                    <span>{tab.label}</span>
                                </div>
                            </button>
                        )
                    })}
                </nav>
            </div>

            {/* Tab Content */}
            <div className="mt-6">
                {activeTab === 'batches' && <BatchListPage />}
                {activeTab === 'qr-code' && <QRCodeGeneratorPage />}
                {activeTab === 'nfc' && <NFCManagerPage />}
            </div>
        </div>
    )
}
