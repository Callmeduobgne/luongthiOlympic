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
