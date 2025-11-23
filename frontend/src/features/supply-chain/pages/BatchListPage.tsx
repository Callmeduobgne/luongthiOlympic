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
import { Plus, Search } from 'lucide-react'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { BatchCard } from '../components/BatchCard'
import { EmptyState } from '@shared/components/common/EmptyState'
import { LoadingState } from '@shared/components/common/LoadingState'
import { useBatch } from '../hooks/useBatches'
import { useDebounce } from '@shared/hooks/useDebounce'
import { useNavigate } from 'react-router-dom'

export const BatchListPage = () => {
  const navigate = useNavigate()
  const [searchId, setSearchId] = useState('')
  const debouncedSearchId = useDebounce(searchId, 500)

  // Query batch by ID if search ID is provided
  const {
    data: batch,
    isLoading,
    error,
  } = useBatch(debouncedSearchId)

  const handleCreate = () => {
    navigate('/supply-chain/create')
  }

  const batches = batch ? [batch] : []

  return (
    <div className="space-y-6 text-white">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">
            Supply Chain
          </h1>
          <p className="mt-1 text-sm text-gray-400">
            Track and manage tea batches
          </p>
        </div>
        <Button onClick={handleCreate} variant="primary">
          <Plus className="h-4 w-4 mr-2" />
          Create Batch
        </Button>
      </div>

      {/* Search */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-white/50" />
          <Input
            placeholder="Search by Batch ID (e.g., BATCH001)"
            value={searchId}
            onChange={(e) => setSearchId(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* Results */}
      {isLoading && <LoadingState text="Searching batch..." />}

      {error && !isLoading && (
        <EmptyState
          icon="search"
          title="Batch not found"
          description={
            searchId
              ? `No batch found with ID: ${searchId}`
              : 'Enter a Batch ID to search'
          }
        />
      )}

      {!isLoading && !error && batches.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {batches.map((batch) => (
            <BatchCard key={batch.batchId} batch={batch} />
          ))}
        </div>
      )}

      {!isLoading && !error && batches.length === 0 && !searchId && (
        <EmptyState
          icon="package"
          title="No batches found"
          description="Search for a batch by ID or create a new batch"
          action={{
            label: 'Create Batch',
            onClick: handleCreate,
          }}
        />
      )}
    </div>
  )
}

