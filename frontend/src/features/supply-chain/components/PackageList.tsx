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

import { Package, QrCode, Calendar, Weight } from 'lucide-react'
import { Badge } from '@shared/components/ui/Badge'
import { Button } from '@shared/components/ui/Button'
import { formatDate } from '@shared/utils/formatters'
import type { TeaPackage, PackageStatus } from '../types/package.types'
import type { BadgeVariant } from '@shared/components/ui/Badge'

interface PackageListProps {
    packages: TeaPackage[]
    onViewQRCode?: (packageId: string) => void
}

const getStatusVariant = (status: PackageStatus): BadgeVariant => {
    switch (status) {
        case 'CREATED':
            return 'created'
        case 'VERIFIED':
            return 'verified'
        case 'SOLD':
            return 'shipped'
        case 'EXPIRED':
            return 'failed'
        default:
            return 'default'
    }
}

export const PackageList = ({ packages, onViewQRCode }: PackageListProps) => {
    if (packages.length === 0) {
        return (
            <div className="text-center py-12">
                <Package className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-400">No packages created yet</p>
                <p className="text-sm text-gray-500 mt-2">
                    Click "Create Package" to add packages from this batch
                </p>
            </div>
        )
    }

    return (
        <div className="space-y-3">
            {packages.map((pkg) => (
                <div
                    key={pkg.packageId}
                    className="flex items-center justify-between p-4 rounded-2xl border border-white/10 bg-black/20 hover:bg-black/30 transition-colors"
                >
                    <div className="flex items-start gap-4 flex-1">
                        <div className="p-2 rounded-xl bg-white/5">
                            <Package className="h-5 w-5 text-white/60" />
                        </div>
                        <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                                <h4 className="text-sm font-medium text-white truncate">
                                    {pkg.packageId}
                                </h4>
                                <Badge variant={getStatusVariant(pkg.status)} size="sm">
                                    {pkg.status}
                                </Badge>
                            </div>
                            <div className="flex items-center gap-4 text-xs text-gray-400">
                                <div className="flex items-center gap-1">
                                    <Weight className="h-3 w-3" />
                                    <span>{pkg.weight}g</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Calendar className="h-3 w-3" />
                                    <span>{formatDate(pkg.productionDate, 'dd/MM/yyyy')}</span>
                                </div>
                                {pkg.expiryDate && (
                                    <div className="flex items-center gap-1">
                                        <span className="text-gray-500">Exp:</span>
                                        <span>{formatDate(pkg.expiryDate, 'dd/MM/yyyy')}</span>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                    {onViewQRCode && (
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => onViewQRCode(pkg.packageId)}
                            className="ml-4"
                        >
                            <QrCode className="h-4 w-4" />
                        </Button>
                    )}
                </div>
            ))}
        </div>
    )
}
