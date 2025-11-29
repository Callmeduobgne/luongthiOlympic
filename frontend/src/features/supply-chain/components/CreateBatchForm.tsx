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

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { Card } from '@shared/components/ui/Card'
import { useCreateBatch } from '../hooks/useBatches'
import { useNavigate } from 'react-router-dom'
import { batchIdSchema } from '@shared/utils/validators'

const createBatchSchema = z.object({
  batchId: batchIdSchema,
  farmLocation: z.string().min(1, 'Farm location is required'),
  harvestDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (YYYY-MM-DD)'),
  processingInfo: z.string().min(1, 'Processing info is required'),
  qualityCert: z.string().min(1, 'Quality certificate is required'),
})

type CreateBatchFormData = z.infer<typeof createBatchSchema>

export const CreateBatchForm = () => {
  const navigate = useNavigate()
  const createBatch = useCreateBatch()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CreateBatchFormData>({
    resolver: zodResolver(createBatchSchema),
    defaultValues: {
      harvestDate: new Date().toISOString().split('T')[0],
    },
  })

  const onSubmit = async (data: CreateBatchFormData) => {
    try {
      // Transform form data to match backend API
      const apiData = {
        batch_id: data.batchId,
        farm_name: data.farmLocation,
        harvest_date: data.harvestDate,
        certification: data.processingInfo, // Using processingInfo as certification
        certificate_id: data.qualityCert,
      }
      const result = await createBatch.mutateAsync(apiData)
      navigate(`/supply-chain/${result.batchId}`)
    } catch {
      // Error handling is done in the mutation
    }
  }

  return (
    <Card className="p-8 max-w-2xl mx-auto text-white">
      <h2 className="text-2xl font-bold mb-6">
        Create New Batch
      </h2>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <Input
          label="Batch ID"
          placeholder="BATCH001"
          {...register('batchId')}
          error={errors.batchId?.message}
          required
        />

        <Input
          label="Farm Location"
          placeholder="Moc Chau, Son La"
          {...register('farmLocation')}
          error={errors.farmLocation?.message}
          required
        />

        <Input
          label="Harvest Date"
          type="date"
          {...register('harvestDate')}
          error={errors.harvestDate?.message}
          required
        />

        <div>
          <label className="block text-sm font-medium text-gray-200 mb-2">
            Processing Info
          </label>
          <textarea
            className="flex min-h-[100px] w-full rounded-2xl border border-white/15 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-white/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
            placeholder="Organic processing, no pesticides..."
            {...register('processingInfo')}
          />
          {errors.processingInfo && (
            <p className="mt-1 text-sm text-red-400" role="alert">
              {errors.processingInfo.message}
            </p>
          )}
        </div>

        <Input
          label="Quality Certificate"
          placeholder="VN-ORG-2024"
          {...register('qualityCert')}
          error={errors.qualityCert?.message}
          required
        />

        <div className="flex gap-4 pt-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => navigate('/supply-chain')}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            isLoading={isSubmitting}
            className="flex-1"
          >
            Create Batch
          </Button>
        </div>
      </form>
    </Card>
  )
}

