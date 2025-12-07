import { useState } from 'react'
import { Button } from '@shared/components/ui/Button'
import { useCreatePackage } from '../hooks/usePackages'
import type { CreatePackageRequest } from '../types/package.types'

interface CreatePackageFormProps {
    batchId: string
    onSuccess?: () => void
    onCancel?: () => void
}

export const CreatePackageForm = ({ batchId, onSuccess, onCancel }: CreatePackageFormProps) => {
    const createPackage = useCreatePackage()
    const [formData, setFormData] = useState({
        packageId: '',
        weight: '',
        productionDate: new Date().toISOString().split('T')[0],
        expiryDate: '',
    })

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        const request: CreatePackageRequest = {
            package_id: formData.packageId,
            batch_id: batchId,
            weight: parseFloat(formData.weight),
            production_date: formData.productionDate,
            expiry_date: formData.expiryDate || undefined,
        }

        try {
            await createPackage.mutateAsync(request)
            // Reset form
            setFormData({
                packageId: '',
                weight: '',
                productionDate: new Date().toISOString().split('T')[0],
                expiryDate: '',
            })
            onSuccess?.()
        } catch {
            // Error handling is done in the mutation
        }
    }

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-200 mb-2">
                    Package ID *
                </label>
                <input
                    type="text"
                    value={formData.packageId}
                    onChange={(e) => setFormData({ ...formData, packageId: e.target.value })}
                    placeholder="PKG_001"
                    required
                    className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                />
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-200 mb-2">
                    Weight (grams) *
                </label>
                <input
                    type="number"
                    value={formData.weight}
                    onChange={(e) => setFormData({ ...formData, weight: e.target.value })}
                    placeholder="100"
                    min="0"
                    step="0.01"
                    required
                    className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                />
            </div>

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

            <div className="pt-4 border-t border-white/10">
                <p className="text-xs text-gray-400 mb-4">
                    * This will create a blockchain transaction
                </p>
                <div className="flex gap-4">
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
                        isLoading={createPackage.isPending}
                        disabled={!formData.packageId || !formData.weight}
                        className="flex-1"
                    >
                        Create Package
                    </Button>
                </div>
            </div>
        </form>
    )
}
