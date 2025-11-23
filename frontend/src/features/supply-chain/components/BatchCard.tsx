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

import { useNavigate } from 'react-router-dom'
import { Package, MapPin, Calendar, CheckCircle } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Badge } from '@shared/components/ui/Badge'
import { formatDate, formatHash } from '@shared/utils/formatters'
import type { TeaBatch } from '../types/batch.types'
import type { BadgeVariant } from '@shared/components/ui/Badge'

interface BatchCardProps {
  batch: TeaBatch
}

export const BatchCard = ({ batch }: BatchCardProps) => {
  const navigate = useNavigate()

  const handleClick = () => {
    navigate(`/supply-chain/${batch.batchId}`)
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
    <Card
      className="p-6 cursor-pointer hover:shadow-[0_20px_45px_rgba(0,0,0,0.6)] transition-all"
      onClick={handleClick}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-2xl border border-white/15 bg-white/10">
            <Package className="h-5 w-5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-lg text-white">
              {batch.batchId}
            </h3>
            <p className="text-sm text-gray-300">
              {formatDate(batch.timestamp, 'dd/MM/yyyy HH:mm')}
            </p>
          </div>
        </div>
        <Badge variant={getStatusVariant(batch.status)}>
          {batch.status}
        </Badge>
      </div>

      <div className="space-y-2 text-gray-300">
        <div className="flex items-center gap-2 text-sm">
          <MapPin className="h-4 w-4" />
          <span>{batch.farmLocation}</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <Calendar className="h-4 w-4" />
          <span>Harvest: {formatDate(batch.harvestDate, 'dd/MM/yyyy')}</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <CheckCircle className="h-4 w-4" />
          <span>Cert: {batch.qualityCert}</span>
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-white/10">
        <p className="text-xs text-gray-400 font-mono">
          Hash: {formatHash(batch.hashValue)}
        </p>
      </div>
    </Card>
  )
}

