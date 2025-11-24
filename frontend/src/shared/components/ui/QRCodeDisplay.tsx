/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Download, Loader2, AlertCircle } from 'lucide-react'
import { Button } from './Button'
import { Spinner } from './Spinner'
import { qrcodeApi } from '@features/supply-chain/services/qrcodeApi'
import api from '@shared/utils/api'
import { API_CONFIG } from '@shared/config/api.config'

interface QRCodeDisplayProps {
  batchId?: string
  packageId?: string
  txId?: string
  size?: number
  showDownload?: boolean
  className?: string
}

export const QRCodeDisplay = ({
  batchId,
  packageId,
  txId,
  size = 256,
  showDownload = true,
  className = '',
}: QRCodeDisplayProps) => {
  const [downloadLoading, setDownloadLoading] = useState(false)

  // Determine which API to use
  const qrCodeQuery = useQuery({
    queryKey: ['qrcode', batchId || packageId || txId],
    queryFn: async () => {
      if (batchId) {
        return await qrcodeApi.getBatchQRCodeBase64(batchId)
      }
      if (packageId) {
        return await qrcodeApi.getPackageQRCodeBase64(packageId)
      }
      // For txId, we use PNG URL directly (backend auto-detects)
      if (txId) {
        return qrcodeApi.getTransactionQRCodeUrl(txId)
      }
      throw new Error('batchId, packageId, or txId is required')
    },
    enabled: !!(batchId || packageId || txId),
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  })

  const handleDownload = async () => {
    if (!qrCodeQuery.data) return

    setDownloadLoading(true)
    try {
      let imageUrl: string
      let filename: string

      if (batchId) {
        imageUrl = qrcodeApi.getBatchQRCodeUrl(batchId)
        filename = `qr-batch-${batchId}.png`
      } else if (packageId) {
        imageUrl = qrcodeApi.getPackageQRCodeUrl(packageId)
        filename = `qr-package-${packageId}.png`
      } else if (txId) {
        imageUrl = qrcodeApi.getTransactionQRCodeUrl(txId)
        filename = `qr-transaction-${txId}.png`
      } else {
        return
      }

      // Fetch image with auth token
      const response = await api.get(imageUrl, {
        responseType: 'blob',
      })

      // Create blob URL and download
      const blob = new Blob([response.data], { type: 'image/png' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download QR code:', error)
    } finally {
      setDownloadLoading(false)
    }
  }

  if (qrCodeQuery.isLoading) {
    return (
      <div
        className={`flex flex-col items-center justify-center p-8 border border-white/15 rounded-2xl bg-black/40 ${className}`}
        style={{ width: size, height: size }}
      >
        <Spinner size="lg" />
        <p className="mt-4 text-sm text-gray-400">Generating QR code...</p>
      </div>
    )
  }

  if (qrCodeQuery.isError) {
    return (
      <div
        className={`flex flex-col items-center justify-center p-8 border border-red-500/30 rounded-2xl bg-red-500/10 ${className}`}
        style={{ width: size, height: size }}
      >
        <AlertCircle className="h-12 w-12 text-red-500 mb-4" />
        <p className="text-sm text-red-400 text-center">
          Failed to load QR code
        </p>
        <p className="text-xs text-gray-400 mt-2 text-center">
          {qrCodeQuery.error instanceof Error
            ? qrCodeQuery.error.message
            : 'Unknown error'}
        </p>
      </div>
    )
  }

  if (!qrCodeQuery.data) {
    return null
  }

  // Check if data is base64 data URI or URL
  const isDataUri = qrCodeQuery.data.startsWith('data:image')
  const imageSrc = isDataUri
    ? qrCodeQuery.data
    : `${API_CONFIG.BASE_URL}${qrCodeQuery.data}`

  return (
    <div className={`flex flex-col items-center ${className}`}>
      <div
        className="relative border-4 border-white/20 rounded-2xl bg-white p-4 shadow-lg"
        style={{ width: size, height: size }}
      >
        <img
          src={imageSrc}
          alt={`QR Code for ${batchId || packageId || txId}`}
          className="w-full h-full object-contain"
          style={{ imageRendering: 'crisp-edges' }}
        />
      </div>

      {showDownload && (
        <Button
          variant="secondary"
          size="sm"
          onClick={handleDownload}
          disabled={downloadLoading}
          className="mt-4 w-full"
        >
          {downloadLoading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Downloading...
            </>
          ) : (
            <>
              <Download className="h-4 w-4 mr-2" />
              Download QR Code
            </>
          )}
        </Button>
      )}
    </div>
  )
}


 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Download, Loader2, AlertCircle } from 'lucide-react'
import { Button } from './Button'
import { Spinner } from './Spinner'
import { qrcodeApi } from '@features/supply-chain/services/qrcodeApi'
import api from '@shared/utils/api'
import { API_CONFIG } from '@shared/config/api.config'

interface QRCodeDisplayProps {
  batchId?: string
  packageId?: string
  txId?: string
  size?: number
  showDownload?: boolean
  className?: string
}

export const QRCodeDisplay = ({
  batchId,
  packageId,
  txId,
  size = 256,
  showDownload = true,
  className = '',
}: QRCodeDisplayProps) => {
  const [downloadLoading, setDownloadLoading] = useState(false)

  // Determine which API to use
  const qrCodeQuery = useQuery({
    queryKey: ['qrcode', batchId || packageId || txId],
    queryFn: async () => {
      if (batchId) {
        return await qrcodeApi.getBatchQRCodeBase64(batchId)
      }
      if (packageId) {
        return await qrcodeApi.getPackageQRCodeBase64(packageId)
      }
      // For txId, we use PNG URL directly (backend auto-detects)
      if (txId) {
        return qrcodeApi.getTransactionQRCodeUrl(txId)
      }
      throw new Error('batchId, packageId, or txId is required')
    },
    enabled: !!(batchId || packageId || txId),
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  })

  const handleDownload = async () => {
    if (!qrCodeQuery.data) return

    setDownloadLoading(true)
    try {
      let imageUrl: string
      let filename: string

      if (batchId) {
        imageUrl = qrcodeApi.getBatchQRCodeUrl(batchId)
        filename = `qr-batch-${batchId}.png`
      } else if (packageId) {
        imageUrl = qrcodeApi.getPackageQRCodeUrl(packageId)
        filename = `qr-package-${packageId}.png`
      } else if (txId) {
        imageUrl = qrcodeApi.getTransactionQRCodeUrl(txId)
        filename = `qr-transaction-${txId}.png`
      } else {
        return
      }

      // Fetch image with auth token
      const response = await api.get(imageUrl, {
        responseType: 'blob',
      })

      // Create blob URL and download
      const blob = new Blob([response.data], { type: 'image/png' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download QR code:', error)
    } finally {
      setDownloadLoading(false)
    }
  }

  if (qrCodeQuery.isLoading) {
    return (
      <div
        className={`flex flex-col items-center justify-center p-8 border border-white/15 rounded-2xl bg-black/40 ${className}`}
        style={{ width: size, height: size }}
      >
        <Spinner size="lg" />
        <p className="mt-4 text-sm text-gray-400">Generating QR code...</p>
      </div>
    )
  }

  if (qrCodeQuery.isError) {
    return (
      <div
        className={`flex flex-col items-center justify-center p-8 border border-red-500/30 rounded-2xl bg-red-500/10 ${className}`}
        style={{ width: size, height: size }}
      >
        <AlertCircle className="h-12 w-12 text-red-500 mb-4" />
        <p className="text-sm text-red-400 text-center">
          Failed to load QR code
        </p>
        <p className="text-xs text-gray-400 mt-2 text-center">
          {qrCodeQuery.error instanceof Error
            ? qrCodeQuery.error.message
            : 'Unknown error'}
        </p>
      </div>
    )
  }

  if (!qrCodeQuery.data) {
    return null
  }

  // Check if data is base64 data URI or URL
  const isDataUri = qrCodeQuery.data.startsWith('data:image')
  const imageSrc = isDataUri
    ? qrCodeQuery.data
    : `${API_CONFIG.BASE_URL}${qrCodeQuery.data}`

  return (
    <div className={`flex flex-col items-center ${className}`}>
      <div
        className="relative border-4 border-white/20 rounded-2xl bg-white p-4 shadow-lg"
        style={{ width: size, height: size }}
      >
        <img
          src={imageSrc}
          alt={`QR Code for ${batchId || packageId || txId}`}
          className="w-full h-full object-contain"
          style={{ imageRendering: 'crisp-edges' }}
        />
      </div>

      {showDownload && (
        <Button
          variant="secondary"
          size="sm"
          onClick={handleDownload}
          disabled={downloadLoading}
          className="mt-4 w-full"
        >
          {downloadLoading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Downloading...
            </>
          ) : (
            <>
              <Download className="h-4 w-4 mr-2" />
              Download QR Code
            </>
          )}
        </Button>
      )}
    </div>
  )
}

