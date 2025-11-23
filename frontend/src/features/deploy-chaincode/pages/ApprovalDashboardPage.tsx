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
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, Clock, CheckCircle, XCircle, MessageSquare, AlertCircle, Loader } from 'lucide-react'
import toast from 'react-hot-toast'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { LoadingState } from '@shared/components/common/LoadingState'
import { EmptyState } from '@shared/components/common/EmptyState'
import { approvalService } from '../services/approvalService'

export const ApprovalDashboardPage = () => {
  const queryClient = useQueryClient()
  const [selectedStatus, setSelectedStatus] = useState<string | undefined>('pending')
  const [comment, setComment] = useState<Record<string, string>>({})

  // Fetch approval requests
  const { data: requests, isLoading } = useQuery({
    queryKey: ['approval-requests', selectedStatus],
    queryFn: () => approvalService.listRequests({ 
      status: selectedStatus 
    }),
    refetchInterval: 30000, // Auto-refresh every 30s
  })

  // Vote mutation
  const voteMutation = useMutation({
    mutationFn: ({ requestId, vote, requestComment }: { 
      requestId: string
      vote: 'approve' | 'reject'
      requestComment?: string 
    }) =>
      approvalService.vote({
        approval_request_id: requestId,
        vote,
        comment: requestComment,
      }),
    onSuccess: (_, variables) => {
      toast.success(`Successfully ${variables.vote}d the request!`)
      queryClient.invalidateQueries({ queryKey: ['approval-requests'] })
      // Clear comment
      setComment(prev => {
        const newComments = { ...prev }
        delete newComments[variables.requestId]
        return newComments
      })
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit vote: ${error.message}`)
    },
  })

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <Badge variant="warning" className="text-xs">⏳ Pending</Badge>
      case 'approved':
        return <Badge variant="success" className="text-xs">✅ Approved</Badge>
      case 'rejected':
        return <Badge variant="danger" className="text-xs">❌ Rejected</Badge>
      case 'expired':
        return <Badge variant="default" className="text-xs">⏰ Expired</Badge>
      default:
        return <Badge variant="default" className="text-xs">{status}</Badge>
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('vi-VN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          <Shield className="w-8 h-8 inline-block mr-2 text-blue-500" />
          Approval Requests
        </h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Review and approve chaincode deployment requests from developers
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6">
        <div className="flex gap-2 flex-wrap">
          {[
            { value: undefined, label: 'All' },
            { value: 'pending', label: 'Pending' },
            { value: 'approved', label: 'Approved' },
            { value: 'rejected', label: 'Rejected' },
          ].map(({ value, label }) => (
            <button
              key={label}
              onClick={() => setSelectedStatus(value)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                selectedStatus === value
                  ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/50'
                  : 'bg-white/5 text-gray-300 border border-white/10 hover:bg-white/10'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Loading State */}
      {isLoading && <LoadingState text="Loading approval requests..." />}

      {/* Empty State */}
      {!isLoading && requests && requests.length === 0 && (
        <EmptyState
          icon="shield"
          title="No approval requests found"
          description={
            selectedStatus === 'pending'
              ? 'No pending approval requests at the moment'
              : `No ${selectedStatus || 'approval'} requests found`
          }
        />
      )}

      {/* Requests List */}
      {!isLoading && requests && requests.length > 0 && (
        <div className="space-y-4">
          {requests.map((request) => (
            <Card key={request.id} className="p-6 hover:shadow-lg transition-shadow">
              <div className="flex justify-between items-start gap-4">
                {/* Request Info */}
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-3">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                      {request.operation.toUpperCase()} Operation
                    </h3>
                    {getStatusBadge(request.status)}
                  </div>

                  <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                    <div>
                      <span className="text-gray-500 dark:text-gray-400 block mb-1">
                        <Clock className="w-4 h-4 inline mr-1" />
                        Requested At:
                      </span>
                      <span className="text-gray-900 dark:text-white font-medium">
                        {formatDate(request.requested_at)}
                      </span>
                    </div>

                    <div>
                      <span className="text-gray-500 dark:text-gray-400 block mb-1">
                        Requested By:
                      </span>
                      <span className="text-gray-900 dark:text-white font-medium font-mono text-xs">
                        {request.requested_by}
                      </span>
                    </div>

                    <div>
                      <span className="text-gray-500 dark:text-gray-400 block mb-1">
                        Chaincode Version:
                      </span>
                      <span className="text-gray-900 dark:text-white font-medium font-mono text-xs">
                        {request.chaincode_version_id}
                      </span>
                    </div>

                    {request.expires_at && (
                      <div>
                        <span className="text-gray-500 dark:text-gray-400 block mb-1">
                          <AlertCircle className="w-4 h-4 inline mr-1" />
                          Expires At:
                        </span>
                        <span className={`font-medium ${
                          new Date(request.expires_at) < new Date()
                            ? 'text-red-400'
                            : 'text-gray-900 dark:text-white'
                        }`}>
                          {formatDate(request.expires_at)}
                        </span>
                      </div>
                    )}
                  </div>

                  {request.reason && (
                    <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg mb-4">
                      <p className="text-sm text-gray-700 dark:text-gray-300">
                        <MessageSquare className="w-4 h-4 inline mr-1" />
                        <strong>Reason:</strong> {request.reason}
                      </p>
                    </div>
                  )}

                  {/* Comment Input (only for pending requests) */}
                  {request.status === 'pending' && (
                    <div className="mt-4">
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Comment (optional):
                      </label>
                      <textarea
                        value={comment[request.id] || ''}
                        onChange={(e) => setComment({ ...comment, [request.id]: e.target.value })}
                        placeholder="Add a comment about your decision..."
                        className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm"
                        rows={2}
                      />
                    </div>
                  )}
                </div>

                {/* Actions */}
                {request.status === 'pending' && (
                  <div className="flex flex-col gap-2 min-w-[120px]">
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() =>
                        voteMutation.mutate({
                          requestId: request.id,
                          vote: 'approve',
                          requestComment: comment[request.id],
                        })
                      }
                      disabled={voteMutation.isPending}
                      className="w-full"
                    >
                      {voteMutation.isPending ? (
                        <Loader className="w-4 h-4 animate-spin" />
                      ) : (
                        <>
                          <CheckCircle className="w-4 h-4 mr-1" />
                          Approve
                        </>
                      )}
                    </Button>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() =>
                        voteMutation.mutate({
                          requestId: request.id,
                          vote: 'reject',
                          requestComment: comment[request.id],
                        })
                      }
                      disabled={voteMutation.isPending}
                      className="w-full"
                    >
                      {voteMutation.isPending ? (
                        <Loader className="w-4 h-4 animate-spin" />
                      ) : (
                        <>
                          <XCircle className="w-4 h-4 mr-1" />
                          Reject
                        </>
                      )}
                    </Button>
                  </div>
                )}

                {/* Status Icons for approved/rejected */}
                {request.status === 'approved' && (
                  <div className="text-green-500">
                    <CheckCircle className="w-12 h-12" />
                  </div>
                )}
                {request.status === 'rejected' && (
                  <div className="text-red-500">
                    <XCircle className="w-12 h-12" />
                  </div>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
