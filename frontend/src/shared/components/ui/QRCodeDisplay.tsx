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

import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Download, Loader2 } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { Button } from './Button'
import { Spinner } from './Spinner'
import { qrcodeApi } from '@features/supply-chain/services/qrcodeApi'
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

  // Generate verification URL for QR code
  const verificationUrl = useMemo(() => {
    if (packageId) {
      return `${window.location.origin}/verify/packages/${packageId}`
    }
    if (txId) {
      return `${window.location.origin}/verify/hash?hash=${txId}`
    }
    if (batchId) {
      return `${window.location.origin}/verify/batches/${batchId}`
    }
    return ''
  }, [packageId, txId, batchId])

  // For package and batch, try to get QR code from backend (with fallback to client-side)
  const qrCodeQuery = useQuery({
    queryKey: ['qrcode', batchId || packageId],
    queryFn: async () => {
      if (batchId) {
        try {
          return await qrcodeApi.getBatchQRCodeBase64(batchId)
        } catch {
          return null // Fallback to client-side
        }
      }
      if (packageId) {
        try {
          return await qrcodeApi.getPackageQRCodeBase64(packageId)
        } catch {
          return null // Fallback to client-side
        }
      }
      return null
    },
    enabled: !!(batchId || packageId) && !txId, // Only for batch/package, not transaction
    staleTime: 5 * 60 * 1000,
    retry: false, // Don't retry if backend endpoint doesn't exist
  })

  // Use backend QR code if available, otherwise use client-side generation
  const useBackendQR = !!(qrCodeQuery.data && (batchId || packageId))

  const handleDownload = async () => {
    setDownloadLoading(true)
    try {
      // If using backend QR code, download from data URI
      if (useBackendQR && qrCodeQuery.data) {
        const isDataUri = typeof qrCodeQuery.data === 'string' && qrCodeQuery.data.startsWith('data:image')
        if (isDataUri) {
          const link = document.createElement('a')
          link.href = qrCodeQuery.data
          link.download = `qr-${batchId || packageId || txId}.png`
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          return
        }
      }

      // For client-side QR code (SVG), convert to PNG
      const svgElement = document.querySelector(`#qrcode-${batchId || packageId || txId}`) as SVGSVGElement
      if (!svgElement) return

      // Create canvas and draw SVG
      const canvas = document.createElement('canvas')
      canvas.width = size
      canvas.height = size
      const ctx = canvas.getContext('2d')
      if (!ctx) return

      // Convert SVG to image
      const svgData = new XMLSerializer().serializeToString(svgElement)
      const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' })
      const url = URL.createObjectURL(svgBlob)

      const img = new Image()
      img.onload = () => {
        ctx.fillStyle = 'white'
        ctx.fillRect(0, 0, canvas.width, canvas.height)
        ctx.drawImage(img, 0, 0, size, size)
        URL.revokeObjectURL(url)

        // Download as PNG
        canvas.toBlob((blob) => {
          if (!blob) return
          const downloadUrl = window.URL.createObjectURL(blob)
          const link = document.createElement('a')
          link.href = downloadUrl
          link.download = `qr-${batchId || packageId || txId}.png`
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          window.URL.revokeObjectURL(downloadUrl)
        }, 'image/png')
      }
      img.src = url
    } catch (error) {
      console.error('Failed to download QR code:', error)
    } finally {
      setDownloadLoading(false)
    }
  }

  if (!verificationUrl) {
    return null
  }

  // Show loading only if trying to fetch from backend
  if (qrCodeQuery.isLoading && (batchId || packageId)) {
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

  // Use backend QR code if available, otherwise generate client-side
  if (useBackendQR && qrCodeQuery.data) {
    const isDataUri = typeof qrCodeQuery.data === 'string' && qrCodeQuery.data.startsWith('data:image')
    const isFullUrl = typeof qrCodeQuery.data === 'string' && (qrCodeQuery.data.startsWith('http://') || qrCodeQuery.data.startsWith('https://'))
    const imageSrc = isDataUri
      ? qrCodeQuery.data
      : isFullUrl
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

  // Client-side QR code generation
  return (
    <div className={`flex flex-col items-center ${className}`}>
      <div
        className="relative border-4 border-white/20 rounded-2xl bg-white p-4 shadow-lg flex items-center justify-center"
        style={{ width: size, height: size }}
      >
        <QRCodeSVG
          id={`qrcode-${batchId || packageId || txId}`}
          value={verificationUrl}
          size={size - 32} // Account for padding
          level="H" // High error correction
          includeMargin={false}
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
