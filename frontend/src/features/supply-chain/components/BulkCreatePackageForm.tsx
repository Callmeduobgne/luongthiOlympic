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
import { Button } from '@shared/components/ui/Button'
import { useCreatePackage } from '../hooks/usePackages'
import type { CreatePackageRequest } from '../types/package.types'
import { AlertCircle, CheckCircle, Loader2 } from 'lucide-react'

interface BulkCreatePackageFormProps {
    batchId: string
    onSuccess?: () => void
    onCancel?: () => void
}

interface BulkProgress {
    total: number
    completed: number
    failed: number
    current: string
    errors: Array<{ packageId: string; error: string }>
}

export const BulkCreatePackageForm = ({ batchId, onSuccess, onCancel }: BulkCreatePackageFormProps) => {
    const createPackage = useCreatePackage()
    const [formData, setFormData] = useState({
        prefix: '',
        startNumber: '1',
        count: '10',
        weight: '100',
        productionDate: new Date().toISOString().split('T')[0],
        expiryDate: '',
    })
    const [isCreating, setIsCreating] = useState(false)
    const [progress, setProgress] = useState<BulkProgress | null>(null)

    const generatePackageIds = () => {
        const start = parseInt(formData.startNumber)
        const count = parseInt(formData.count)
        const ids: string[] = []

        for (let i = 0; i < count; i++) {
            const num = (start + i).toString().padStart(3, '0')
            ids.push(`${formData.prefix}${num}`)
        }

        return ids
    }

    const previewIds = formData.prefix && formData.count ? generatePackageIds().slice(0, 3) : []

    const handleBulkCreate = async (e: React.FormEvent) => {
        e.preventDefault()

        const packageIds = generatePackageIds()
        const total = packageIds.length

        setIsCreating(true)
        setProgress({
            total,
            completed: 0,
            failed: 0,
            current: packageIds[0],
            errors: []
        })

        let completed = 0
        let failed = 0
        const errors: Array<{ packageId: string; error: string }> = []

        // Create packages sequentially to avoid overwhelming the blockchain
        for (const packageId of packageIds) {
            setProgress(prev => prev ? { ...prev, current: packageId } : null)

            const request: CreatePackageRequest = {
                package_id: packageId,
                batch_id: batchId,
                weight: parseFloat(formData.weight),
                production_date: formData.productionDate,
                expiry_date: formData.expiryDate || undefined,
            }

            try {
                await createPackage.mutateAsync(request)
                completed++
            } catch (error) {
                failed++
                errors.push({
                    packageId,
                    error: error instanceof Error ? error.message : 'Unknown error'
                })
            }

            setProgress({
                total,
                completed,
                failed,
                current: packageId,
                errors
            })

            // Small delay to prevent overwhelming the system
            await new Promise(resolve => setTimeout(resolve, 500))
        }

        setIsCreating(false)

        // If all succeeded, close modal
        if (failed === 0) {
            setTimeout(() => {
                onSuccess?.()
            }, 1500)
        }
    }

    const estimatedTime = () => {
        const count = parseInt(formData.count) || 0
        const seconds = count * 0.5 // ~500ms per package
        if (seconds < 60) return `~${Math.ceil(seconds)} seconds`
        return `~${Math.ceil(seconds / 60)} minutes`
    }

    return (
        <form onSubmit={handleBulkCreate} className="space-y-4">
            {!isCreating && !progress && (
                <>
                    <div>
                        <label className="block text-sm font-medium text-gray-200 mb-2">
                            Package ID Prefix *
                        </label>
                        <input
                            type="text"
                            value={formData.prefix}
                            onChange={(e) => setFormData({ ...formData, prefix: e.target.value })}
                            placeholder="PKG_TN002_"
                            required
                            className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        />
                        <p className="text-xs text-gray-400 mt-1">
                            Will be followed by numbers (e.g., PKG_TN002_001, PKG_TN002_002...)
                        </p>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-200 mb-2">
                                Start Number *
                            </label>
                            <input
                                type="number"
                                value={formData.startNumber}
                                onChange={(e) => setFormData({ ...formData, startNumber: e.target.value })}
                                min="1"
                                required
                                className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-200 mb-2">
                                Number of Packages *
                            </label>
                            <input
                                type="number"
                                value={formData.count}
                                onChange={(e) => setFormData({ ...formData, count: e.target.value })}
                                min="1"
                                max="1000"
                                required
                                className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-200 mb-2">
                            Weight per Package (grams) *
                        </label>
                        <input
                            type="number"
                            value={formData.weight}
                            onChange={(e) => setFormData({ ...formData, weight: e.target.value })}
                            min="0"
                            step="0.01"
                            required
                            className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-200 mb-2">
                                Production Date *
                            </label>
                            <input
                                type="date"
                                value={formData.productionDate}
                                onChange={(e) => setFormData({ ...formData, productionDate: e.target.value })}
                                required
                                className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-200 mb-2">
                                Expiry Date (Optional)
                            </label>
                            <input
                                type="date"
                                value={formData.expiryDate}
                                onChange={(e) => setFormData({ ...formData, expiryDate: e.target.value })}
                                min={formData.productionDate}
                                className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                            />
                        </div>
                    </div>

                    {previewIds.length > 0 && (
                        <div className="p-4 rounded-2xl bg-white/5 border border-white/10">
                            <p className="text-sm font-medium text-gray-200 mb-2">Preview:</p>
                            <div className="space-y-1">
                                {previewIds.map((id) => (
                                    <p key={id} className="text-xs text-gray-400 font-mono">
                                        {id}
                                    </p>
                                ))}
                                {parseInt(formData.count) > 3 && (
                                    <p className="text-xs text-gray-500">
                                        ... and {parseInt(formData.count) - 3} more
                                    </p>
                                )}
                            </div>
                        </div>
                    )}

                    <div className="p-4 rounded-2xl bg-yellow-500/10 border border-yellow-500/20">
                        <div className="flex items-start gap-3">
                            <AlertCircle className="h-5 w-5 text-yellow-500 mt-0.5 flex-shrink-0" />
                            <div className="text-sm text-yellow-200">
                                <p className="font-medium mb-1">Important:</p>
                                <ul className="space-y-1 text-xs text-yellow-200/80">
                                    <li>• This will create {formData.count} blockchain transactions</li>
                                    <li>• Estimated time: {estimatedTime()}</li>
                                    <li>• Please do not close this window during creation</li>
                                </ul>
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-4 pt-4 border-t border-white/10">
                        {onCancel && (
                            <Button
                                type="button"
                                variant="secondary"
                                onClick={onCancel}
                                className="flex-1"
                            >
                                Cancel
                            </Button>
                        )}
                        <Button
                            type="submit"
                            variant="primary"
                            disabled={!formData.prefix || !formData.count}
                            className="flex-1"
                        >
                            Create {formData.count} Packages
                        </Button>
                    </div>
                </>
            )}

            {(isCreating || progress) && (
                <div className="space-y-4">
                    <div className="text-center">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            {isCreating ? 'Creating Packages...' : 'Creation Complete!'}
                        </h3>
                        <p className="text-sm text-gray-400">
                            {progress?.completed} of {progress?.total} packages created
                        </p>
                    </div>

                    {/* Progress Bar */}
                    <div className="space-y-2">
                        <div className="h-2 bg-white/10 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-gradient-to-r from-blue-500 to-green-500 transition-all duration-300"
                                style={{ width: `${((progress?.completed || 0) / (progress?.total || 1)) * 100}%` }}
                            />
                        </div>
                        <div className="flex justify-between text-xs text-gray-400">
                            <span>{Math.round(((progress?.completed || 0) / (progress?.total || 1)) * 100)}%</span>
                            <span>{progress?.completed}/{progress?.total}</span>
                        </div>
                    </div>

                    {/* Current Package */}
                    {isCreating && (
                        <div className="flex items-center justify-center gap-2 p-3 rounded-2xl bg-white/5">
                            <Loader2 className="h-4 w-4 animate-spin text-blue-400" />
                            <span className="text-sm text-gray-300">
                                Creating: <span className="font-mono text-white">{progress?.current}</span>
                            </span>
                        </div>
                    )}

                    {/* Stats */}
                    <div className="grid grid-cols-2 gap-4">
                        <div className="p-3 rounded-2xl bg-green-500/10 border border-green-500/20">
                            <div className="flex items-center gap-2">
                                <CheckCircle className="h-4 w-4 text-green-400" />
                                <span className="text-sm text-green-200">
                                    Success: {progress?.completed}
                                </span>
                            </div>
                        </div>
                        {(progress?.failed || 0) > 0 && (
                            <div className="p-3 rounded-2xl bg-red-500/10 border border-red-500/20">
                                <div className="flex items-center gap-2">
                                    <AlertCircle className="h-4 w-4 text-red-400" />
                                    <span className="text-sm text-red-200">
                                        Failed: {progress?.failed}
                                    </span>
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Errors */}
                    {progress?.errors && progress.errors.length > 0 && (
                        <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 max-h-40 overflow-y-auto">
                            <p className="text-sm font-medium text-red-200 mb-2">Errors:</p>
                            <div className="space-y-1">
                                {progress.errors.map((err, idx) => (
                                    <p key={idx} className="text-xs text-red-200/80">
                                        {err.packageId}: {err.error}
                                    </p>
                                ))}
                            </div>
                        </div>
                    )}

                    {!isCreating && (
                        <Button
                            variant="primary"
                            onClick={onSuccess}
                            className="w-full"
                        >
                            Done
                        </Button>
                    )}
                </div>
            )}
        </form>
    )
}
