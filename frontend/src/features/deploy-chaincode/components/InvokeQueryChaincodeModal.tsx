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
import { useMutation } from '@tanstack/react-query'
import { X, Play, Search, Loader, AlertCircle, Copy, Check } from 'lucide-react'
import toast from 'react-hot-toast'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { chaincodeService } from '../services/chaincodeService'

interface InvokeQueryChaincodeModalProps {
  onClose: () => void
  mode: 'invoke' | 'query'
  initialChannel?: string
  initialChaincodeName?: string
}

export const InvokeQueryChaincodeModal = ({
  onClose,
  mode,
  initialChannel = 'ibnchannel',
  initialChaincodeName = '',
}: InvokeQueryChaincodeModalProps) => {
  const [channel, setChannel] = useState(initialChannel)
  const [chaincodeName, setChaincodeName] = useState(initialChaincodeName)
  const [functionName, setFunctionName] = useState('')
  const [args, setArgs] = useState<string[]>([''])
  const [result, setResult] = useState<any>(null)
  const [copied, setCopied] = useState(false)

  // Add argument input
  const addArg = () => {
    setArgs([...args, ''])
  }

  // Remove argument input
  const removeArg = (index: number) => {
    setArgs(args.filter((_, i) => i !== index))
  }

  // Update argument value
  const updateArg = (index: number, value: string) => {
    const newArgs = [...args]
    newArgs[index] = value
    setArgs(newArgs)
  }

  // Invoke mutation
  const invokeMutation = useMutation({
    mutationFn: async () => {
      const filteredArgs = args.filter(arg => arg.trim() !== '')
      return await chaincodeService.invoke(channel, chaincodeName, {
        function: functionName,
        args: filteredArgs,
      })
    },
    onSuccess: (data) => {
      setResult(data)
      toast.success('Invoke thành công!')
    },
    onError: (error: Error) => {
      toast.error(`Lỗi invoke: ${error.message || 'Không thể invoke chaincode'}`)
      console.error('[InvokeQueryChaincodeModal] Invoke error:', error)
      setResult(null)
    },
  })

  // Query mutation
  const queryMutation = useMutation({
    mutationFn: async () => {
      const filteredArgs = args.filter(arg => arg.trim() !== '')
      return await chaincodeService.query(channel, chaincodeName, {
        function: functionName,
        args: filteredArgs,
      })
    },
    onSuccess: (data) => {
      setResult(data)
      toast.success('Query thành công!')
    },
    onError: (error: Error) => {
      toast.error(`Lỗi query: ${error.message || 'Không thể query chaincode'}`)
      console.error('[InvokeQueryChaincodeModal] Query error:', error)
      setResult(null)
    },
  })

  // Handle submit
  const handleSubmit = () => {
    if (!channel.trim()) {
      toast.error('Vui lòng nhập channel name')
      return
    }
    if (!chaincodeName.trim()) {
      toast.error('Vui lòng nhập chaincode name')
      return
    }
    if (!functionName.trim()) {
      toast.error('Vui lòng nhập function name')
      return
    }

    if (mode === 'invoke') {
      invokeMutation.mutate()
    } else {
      queryMutation.mutate()
    }
  }

  // Copy result to clipboard
  const copyResult = async () => {
    try {
      const resultStr = typeof result === 'string' ? result : JSON.stringify(result, null, 2)
      await navigator.clipboard.writeText(resultStr)
      setCopied(true)
      toast.success('Đã copy kết quả!')
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      toast.error('Không thể copy kết quả')
    }
  }

  const isLoading = invokeMutation.isPending || queryMutation.isPending

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              {mode === 'invoke' ? (
                <Play className="w-6 h-6 text-blue-500" />
              ) : (
                <Search className="w-6 h-6 text-green-500" />
              )}
              <h2 className="text-2xl font-bold">
                {mode === 'invoke' ? 'Invoke Chaincode' : 'Query Chaincode'}
              </h2>
            </div>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="w-5 h-5" />
            </Button>
          </div>

          {/* Form */}
          <div className="space-y-4">
            {/* Channel */}
            <div>
              <label className="block text-sm font-medium mb-2">Channel Name *</label>
              <Input
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                placeholder="ibnchannel"
                disabled={isLoading}
              />
            </div>

            {/* Chaincode Name */}
            <div>
              <label className="block text-sm font-medium mb-2">Chaincode Name *</label>
              <Input
                value={chaincodeName}
                onChange={(e) => setChaincodeName(e.target.value)}
                placeholder="teaTraceCC"
                disabled={isLoading}
              />
            </div>

            {/* Function Name */}
            <div>
              <label className="block text-sm font-medium mb-2">Function Name *</label>
              <Input
                value={functionName}
                onChange={(e) => setFunctionName(e.target.value)}
                placeholder="getBatchInfo"
                disabled={isLoading}
              />
            </div>

            {/* Arguments */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="block text-sm font-medium">Arguments</label>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={addArg}
                  disabled={isLoading}
                >
                  + Thêm Argument
                </Button>
              </div>
              <div className="space-y-2">
                {args.map((arg, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      value={arg}
                      onChange={(e) => updateArg(index, e.target.value)}
                      placeholder={`Argument ${index + 1}`}
                      disabled={isLoading}
                    />
                    {args.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeArg(index)}
                        disabled={isLoading}
                      >
                        <X className="w-4 h-4" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {/* Submit Button */}
            <div className="flex gap-3 pt-4">
              <Button
                onClick={handleSubmit}
                disabled={isLoading}
                className="flex-1"
              >
                {isLoading ? (
                  <>
                    <Loader className="w-4 h-4 mr-2 animate-spin" />
                    Đang xử lý...
                  </>
                ) : (
                  <>
                    {mode === 'invoke' ? (
                      <Play className="w-4 h-4 mr-2" />
                    ) : (
                      <Search className="w-4 h-4 mr-2" />
                    )}
                    {mode === 'invoke' ? 'Invoke' : 'Query'}
                  </>
                )}
              </Button>
              <Button variant="secondary" onClick={onClose} disabled={isLoading}>
                Đóng
              </Button>
            </div>

            {/* Result */}
            {result !== null && (
              <div className="mt-6">
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium">Kết quả</label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={copyResult}
                    className="flex items-center gap-2"
                  >
                    {copied ? (
                      <>
                        <Check className="w-4 h-4 text-green-500" />
                        Đã copy
                      </>
                    ) : (
                      <>
                        <Copy className="w-4 h-4" />
                        Copy
                      </>
                    )}
                  </Button>
                </div>
                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 border">
                  <pre className="text-sm overflow-x-auto">
                    {typeof result === 'string' ? result : JSON.stringify(result, null, 2)}
                  </pre>
                </div>
              </div>
            )}

            {/* Error Display */}
            {(invokeMutation.isError || queryMutation.isError) && (
              <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <div className="flex items-start gap-2">
                  <AlertCircle className="w-5 h-5 text-red-500 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-red-800 dark:text-red-200">
                      {mode === 'invoke' ? 'Lỗi invoke' : 'Lỗi query'}
                    </p>
                    <p className="text-sm text-red-600 dark:text-red-300 mt-1">
                      {invokeMutation.error?.message || queryMutation.error?.message || 'Đã xảy ra lỗi'}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </Card>
    </div>
  )
}

