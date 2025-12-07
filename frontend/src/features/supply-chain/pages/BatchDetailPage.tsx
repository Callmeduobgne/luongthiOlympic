import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Package, MapPin, Calendar, CheckCircle, QrCode, PackagePlus, Boxes } from 'lucide-react'
import { Button } from '@shared/components/ui/Button'
import { Card } from '@shared/components/ui/Card'
import { Badge } from '@shared/components/ui/Badge'
import { QRCodeDisplay } from '@shared/components/ui/QRCodeDisplay'
import { LoadingState } from '@shared/components/common/LoadingState'
import { EmptyState } from '@shared/components/common/EmptyState'
import { useBatch, useUpdateBatchStatus } from '../hooks/useBatches'
import { usePackagesByBatch } from '../hooks/usePackages'
import { CreatePackageForm } from '../components/CreatePackageForm'
import { BulkCreatePackageForm } from '../components/BulkCreatePackageForm'
import { PackageList } from '../components/PackageList'
import { formatDate } from '@shared/utils/formatters'
import { BATCH_STATUS, type BatchStatus } from '@shared/utils/constants'
import { useState } from 'react'
import { Modal } from '@shared/components/ui/Modal'
import type { BadgeVariant } from '@shared/components/ui/Badge'

export const BatchDetailPage = () => {
  const { batchId } = useParams<{ batchId: string }>()
  const navigate = useNavigate()
  const [showStatusModal, setShowStatusModal] = useState(false)
  const [showQRModal, setShowQRModal] = useState(false)
  const [showCreatePackageModal, setShowCreatePackageModal] = useState(false)
  const [showBulkCreateModal, setShowBulkCreateModal] = useState(false)
  const [showPackageQRModal, setShowPackageQRModal] = useState(false)
  const [selectedPackageId, setSelectedPackageId] = useState<string>('')
  const [selectedStatus, setSelectedStatus] = useState<string>('')

  const { data: batch, isLoading, error } = useBatch(batchId || '')
  const { data: packagesData } = usePackagesByBatch({
    batchId: batchId || '',
    limit: 100,
    offset: 0
  })
  // const packagesData = { packages: [], total: 0 } // Temporary: empty state
  const updateStatus = useUpdateBatchStatus()

  const handleUpdateStatus = async () => {
    if (!batchId || !selectedStatus) return

    try {
      await updateStatus.mutateAsync({
        batchId,
        status: selectedStatus as BatchStatus,
      })
      setShowStatusModal(false)
      setSelectedStatus('')
    } catch {
      // Error handling is done in the mutation
    }
  }

  if (isLoading) {
    return <LoadingState text="Loading batch details..." fullScreen />
  }

  if (error || !batch) {
    return (
      <EmptyState
        icon="package"
        title="Batch not found"
        description={`Batch with ID "${batchId}" does not exist`}
        action={{
          label: 'Back to Supply Chain',
          onClick: () => navigate('/supply-chain'),
        }}
      />
    )
  }

  const getStatusVariant = (status: string): BadgeVariant => {
    switch (status) {
      case 'CREATED':
        return 'created'
      case 'VERIFIED':
        return 'verified'
      case 'PROCESSING':
        return 'processing'
      case 'SHIPPED':
        return 'shipped'
      case 'DELIVERED':
        return 'delivered'
      case 'FAILED':
        return 'failed'
      default:
        return 'default'
    }
  }

  return (
    <div className="space-y-6 text-white">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            onClick={() => navigate('/supply-chain')}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold">
              {batch.batchId}
            </h1>
            <p className="mt-1 text-sm text-gray-400">
              Batch Details
            </p>
          </div>
        </div>
        <Badge variant={getStatusVariant(batch.status)} size="lg">
          {batch.status}
        </Badge>
      </div>

      {/* Main Info */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="p-6 lg:col-span-2 text-white">
          <h2 className="text-xl font-semibold mb-4">
            Batch Information
          </h2>
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <MapPin className="h-5 w-5 text-white/60 mt-0.5" />
              <div>
                <p className="text-sm text-gray-400">
                  Farm Location
                </p>
                <p className="text-base font-medium text-white">
                  {batch.farmLocation}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Calendar className="h-5 w-5 text-white/60 mt-0.5" />
              <div>
                <p className="text-sm text-gray-400">
                  Harvest Date
                </p>
                <p className="text-base font-medium text-white">
                  {formatDate(batch.harvestDate, 'dd/MM/yyyy')}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <CheckCircle className="h-5 w-5 text-white/60 mt-0.5" />
              <div>
                <p className="text-sm text-gray-400">
                  Quality Certificate
                </p>
                <p className="text-base font-medium text-white">
                  {batch.qualityCert}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Package className="h-5 w-5 text-white/60 mt-0.5" />
              <div>
                <p className="text-sm text-gray-400">
                  Processing Info
                </p>
                <p className="text-base font-medium text-white">
                  {batch.processingInfo}
                </p>
              </div>
            </div>
          </div>
        </Card>

        {/* Metadata */}
        <Card className="p-6 text-white">
          <h2 className="text-xl font-semibold mb-4">
            Metadata
          </h2>
          <div className="space-y-4">
            <div>
              <p className="text-sm text-gray-400 mb-1">
                Hash Value
              </p>
              <p className="text-xs font-mono text-white break-all">
                {batch.hashValue}
              </p>
            </div>

            <div>
              <p className="text-sm text-gray-400 mb-1">
                Owner
              </p>
              <p className="text-base font-medium text-white">
                {batch.owner}
              </p>
            </div>

            <div>
              <p className="text-sm text-gray-400 mb-1">
                Created At
              </p>
              <p className="text-base font-medium text-white">
                {formatDate(batch.timestamp, 'dd/MM/yyyy HH:mm')}
              </p>
            </div>

            <div className="pt-4 border-t border-white/10 space-y-2">
              <Button
                variant="secondary"
                onClick={() => setShowStatusModal(true)}
                className="w-full"
              >
                Update Status
              </Button>
              <Button
                variant="primary"
                onClick={() => setShowQRModal(true)}
                className="w-full"
              >
                <QrCode className="h-4 w-4 mr-2" />
                View QR Code
              </Button>
            </div>
          </div>
        </Card>
      </div>

      {/* Packages Section */}
      <Card className="p-6 text-white">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">
            Packages from this Batch ({packagesData?.total || 0})
          </h2>
          <div className="flex gap-2">
            <Button
              variant="secondary"
              onClick={() => setShowCreatePackageModal(true)}
              size="sm"
            >
              <PackagePlus className="h-4 w-4 mr-2" />
              Create Package
            </Button>
            <Button
              variant="primary"
              onClick={() => setShowBulkCreateModal(true)}
              size="sm"
            >
              <Boxes className="h-4 w-4 mr-2" />
              Bulk Create
            </Button>
          </div>
        </div>
        <PackageList
          packages={packagesData?.packages || []}
          onViewQRCode={(packageId) => {
            setSelectedPackageId(packageId)
            setShowPackageQRModal(true)
          }}
        />
      </Card>

      {/* Status Update Modal */}
      <Modal
        isOpen={showStatusModal}
        onClose={() => {
          setShowStatusModal(false)
          setSelectedStatus('')
        }}
        title="Update Batch Status"
      >
        <div className="space-y-4 text-white">
          <div>
            <label className="block text-sm font-medium text-gray-200 mb-2">
              Select Status
            </label>
            <select
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
              className="flex h-11 w-full rounded-2xl border border-white/15 bg-black/40 px-4 py-2 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
            >
              <option value="">Select a status...</option>
              {Object.values(BATCH_STATUS).map((status) => (
                <option key={status} value={status}>
                  {status}
                </option>
              ))}
            </select>
          </div>
          <div className="flex gap-4">
            <Button
              variant="secondary"
              onClick={() => {
                setShowStatusModal(false)
                setSelectedStatus('')
              }}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleUpdateStatus}
              isLoading={updateStatus.isPending}
              disabled={!selectedStatus}
              className="flex-1"
            >
              Update
            </Button>
          </div>
        </div>
      </Modal>

      {/* QR Code Modal */}
      <Modal
        isOpen={showQRModal}
        onClose={() => setShowQRModal(false)}
        title="QR Code - Batch Verification"
      >
        <div className="space-y-4 text-white">
          <p className="text-sm text-gray-300 text-center">
            Scan this QR code to verify the batch authenticity on the blockchain
          </p>
          {batchId && (
            <div className="flex justify-center">
              <QRCodeDisplay batchId={batchId} size={256} showDownload />
            </div>
          )}
          <div className="pt-4 border-t border-white/10">
            <Button
              variant="secondary"
              onClick={() => setShowQRModal(false)}
              className="w-full"
            >
              Close
            </Button>
          </div>
        </div>
      </Modal>

      {/* Create Package Modal */}
      <Modal
        isOpen={showCreatePackageModal}
        onClose={() => setShowCreatePackageModal(false)}
        title="Create New Package"
      >
        <div className="text-white">
          <p className="text-sm text-gray-300 mb-4">
            Create a new package from batch <span className="font-medium">{batchId}</span>
          </p>
          <CreatePackageForm
            batchId={batchId || ''}
            onSuccess={() => setShowCreatePackageModal(false)}
            onCancel={() => setShowCreatePackageModal(false)}
          />
        </div>
      </Modal>

      {/* Package QR Code Modal */}
      <Modal
        isOpen={showPackageQRModal}
        onClose={() => setShowPackageQRModal(false)}
        title="QR Code - Package Verification"
      >
        <div className="space-y-4 text-white">
          <p className="text-sm text-gray-300 text-center">
            Scan this QR code to verify the package authenticity on the blockchain
          </p>
          {selectedPackageId && (
            <div className="flex justify-center">
              <QRCodeDisplay packageId={selectedPackageId} size={256} showDownload />
            </div>
          )}
          <div className="pt-4 border-t border-white/10">
            <Button
              variant="secondary"
              onClick={() => setShowPackageQRModal(false)}
              className="w-full"
            >
              Close
            </Button>
          </div>
        </div>
      </Modal>

      {/* Bulk Create Package Modal */}
      <Modal
        isOpen={showBulkCreateModal}
        onClose={() => setShowBulkCreateModal(false)}
        title="Bulk Create Packages"
        size="lg"
      >
        <div className="text-white">
          <p className="text-sm text-gray-300 mb-4">
            Create multiple packages from batch <span className="font-medium">{batchId}</span>
          </p>
          <BulkCreatePackageForm
            batchId={batchId || ''}
            onSuccess={() => setShowBulkCreateModal(false)}
            onCancel={() => setShowBulkCreateModal(false)}
          />
        </div>
      </Modal>
    </div>
  )
}



