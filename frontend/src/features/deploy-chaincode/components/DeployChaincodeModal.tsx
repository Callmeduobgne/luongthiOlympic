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

import { useState, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { X, Upload, Package, CheckCircle, Loader, AlertCircle, TestTube, Shield, Info, Clock } from 'lucide-react'
import toast from 'react-hot-toast'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { FileUpload } from '@shared/components/ui/FileUpload'
import { extractErrorMessage } from '@shared/utils/errorParser'
import { validateChaincodePackage, getPackageWarnings } from '@shared/utils/packageValidator'
import { chaincodeService } from '../services/chaincodeService'
import { approvalService } from '../services/approvalService'
import { testingService } from '../services/testingService'
import { useAuth } from '@features/authentication/hooks/useAuth'
import type { InstallChaincodeRequest, ApproveChaincodeRequest, CommitChaincodeRequest } from '../services/chaincodeService'
import type { ApprovalRequest } from '../services/approvalService'
import type { TestSuite } from '../services/testingService'

interface DeployChaincodeModalProps {
  onClose: () => void
  initialData?: {
    packageId: string
    label: string
    name: string
    version: string
  }
}

type DeployStep = 'install' | 'approve' | 'commit' | 'success' | 'pending-approval'

export const DeployChaincodeModal = ({ onClose, initialData }: DeployChaincodeModalProps) => {
  const queryClient = useQueryClient()
  const { user } = useAuth()
  // If initialData is provided, start from 'approve' step, otherwise 'install'
  const [step, setStep] = useState<DeployStep>(initialData ? 'approve' : 'install')
  
  // Check if user is admin
  const isAdmin = user?.role === 'admin' || user?.role === 'system:admin'
  
  // Install form
  const [packagePath, setPackagePath] = useState('')
  const [label, setLabel] = useState(initialData?.label || '')
  const [packageId, setPackageId] = useState(initialData?.packageId || '')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [packageWarnings, setPackageWarnings] = useState<string[]>([])

  // Approve/Commit form
  const [channelName, setChannelName] = useState('ibnchannel')
  const [name, setName] = useState(initialData?.name || '')
  const [version, setVersion] = useState(initialData?.version || '')
  const [sequence, setSequence] = useState('1')
  const [initRequired, setInitRequired] = useState(false)
  const [endorsementPlugin, setEndorsementPlugin] = useState('escc')
  const [validationPlugin, setValidationPlugin] = useState('vscc')

  // Phase 4: Approval & Testing state
  const [versionId, setVersionId] = useState<string | null>(null)
  const [approvalRequest, setApprovalRequest] = useState<ApprovalRequest | null>(null)
  const [testSuite, setTestSuite] = useState<TestSuite | null>(null)

  // Fetch current sequence when initialData is provided
  useEffect(() => {
    if (initialData?.name && initialData?.version) {
      // Fetch committed chaincode info to get current sequence
      chaincodeService.getCommittedInfo(initialData.name, channelName)
        .then((committedInfo) => {
          // Set sequence to current + 1
          const nextSequence = (committedInfo.sequence || 0) + 1
          setSequence(String(nextSequence))
          console.log(`[DeployChaincodeModal] Auto-set sequence to ${nextSequence} (current: ${committedInfo.sequence})`)
        })
        .catch((error) => {
          // If chaincode not committed yet, use sequence 1
          console.log('[DeployChaincodeModal] Chaincode not committed yet, using sequence 1', error)
          setSequence('1')
        })
    }
  }, [initialData?.name, initialData?.version, channelName])

  // Upload mutation
  const uploadMutation = useMutation({
    mutationFn: (file: File) => chaincodeService.uploadPackage(file),
    onSuccess: (data) => {
      toast.success('Upload thành công! Đang install chaincode...')
      // After upload, automatically install
      installMutation.mutate({
        packagePath: data.filePath,
        label: label || data.filename.replace(/\.(tar\.gz|gz)$/, ''),
      })
    },
    onError: (error: Error) => {
      const errorMsg = extractErrorMessage(error)
      toast.error(`Lỗi upload: ${errorMsg}`)
      console.error('[DeployChaincodeModal] Upload error:', error)
    },
  })

  // Install mutation
  const installMutation = useMutation({
    mutationFn: (request: InstallChaincodeRequest) => chaincodeService.install(request),
    onSuccess: async (data) => {
      toast.success('Install thành công! Package ID: ' + data.packageId.substring(0, 20) + '...')
      setPackageId(data.packageId)
      queryClient.invalidateQueries({ queryKey: ['chaincodes-installed'] })
      
      // Auto-fill name from label if available
      if (!name && label) {
        const suggestedName = label.split('_')[0] || label
        setName(suggestedName)
      }

      // Check user role to determine next step
      if (isAdmin) {
        // Admin: proceed directly to approve step
        setStep('approve')
      } else {
        // User: create approval request
        try {
          // Get version ID first (if registry is enabled)
          let vId: string | null = null
          try {
            vId = await chaincodeService.getLatestVersionId(name || label.split('_')[0], channelName)
            if (vId) setVersionId(vId)
          } catch (error) {
            console.warn('Failed to get version ID (registry may not be enabled):', error)
          }

          // Create approval request
          if (vId) {
            const request = await approvalService.createRequest({
              chaincode_version_id: vId,
              operation: 'approve',
              reason: `Deploy chaincode ${name || label} version ${version || 'latest'} to ${channelName}`,
            })

            setApprovalRequest(request)
            toast.success('Approval request created! Waiting for admin approval...')
            setStep('pending-approval')
          } else {
            // Fallback: if version registry not available, just show pending state
            toast.success('Chaincode installed successfully. Please contact admin for approval.')
            setStep('pending-approval')
          }
        } catch (error) {
          console.error('Failed to create approval request:', error)
          toast.error('Chaincode installed but approval request creation failed. Please contact admin.')
          setStep('pending-approval')
        }
      }
    },
    onError: (error: Error) => {
      const errorMsg = extractErrorMessage(error)
      toast.error(`Lỗi install: ${errorMsg}`)
      console.error('[DeployChaincodeModal] Install error:', error)
    },
  })

  // Approve mutation
  const approveMutation = useMutation({
    mutationFn: (request: ApproveChaincodeRequest) => chaincodeService.approve(request),
    onSuccess: async () => {
      toast.success('Approve thành công! Chuyển sang bước commit...')
      queryClient.invalidateQueries({ queryKey: ['chaincodes-committed'] })
      // After approve, try to get version ID and check approval/tests
      // Wrap in try-catch to prevent crashes if Phase 4 features are not available
      try {
        if (name && channelName) {
          const vId = await chaincodeService.getLatestVersionId(name, channelName)
          if (vId) {
            setVersionId(vId)
            // Check approval and tests, but don't fail if they're not available
            try {
              await checkApprovalAndTests()
            } catch (error) {
              console.warn('Failed to check approval and tests (Phase 4 features may not be available):', error)
            }
          }
        }
      } catch (error) {
        console.warn('Failed to get version ID after approve (Phase 4 features may not be available):', error)
        // Continue anyway - Phase 4 features are optional
      }
      setStep('commit')
    },
    onError: (error: Error) => {
      toast.error(`Lỗi approve: ${error.message || 'Không thể approve chaincode'}`)
      console.error('[DeployChaincodeModal] Approve error:', error)
    },
  })

  // Phase 4: Check approval status
  const checkApprovalAndTests = async () => {
    if (!versionId) return

    try {
      // Check approval requests (optional - Phase 4 feature)
      try {
        const requests = await approvalService.listRequests({
          chaincode_version_id: versionId,
          operation: 'commit',
        })
        if (requests && requests.length > 0) {
          setApprovalRequest(requests[0])
        }
      } catch (error) {
        console.warn('Approval service not available:', error)
        // Continue - approval is optional
      }

      // Check test suites (optional - Phase 4 feature)
      try {
        const suites = await testingService.listTestSuites({
          chaincode_version_id: versionId,
          test_type: 'unit',
        })
        if (suites && suites.length > 0) {
          // Get latest test suite
          const latestSuite = suites.sort((a, b) => 
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
          )[0]
          setTestSuite(latestSuite)
        }
      } catch (error) {
        console.warn('Testing service not available:', error)
        // Continue - testing is optional
      }
    } catch (error) {
      console.warn('Failed to check approval and tests (Phase 4 features may not be available):', error)
      // Don't throw - Phase 4 features are optional
    }
  }

  // Phase 4: Run tests mutation
  const runTestsMutation = useMutation({
    mutationFn: async () => {
      if (!versionId) throw new Error('Version ID is required')
      return testingService.runTestSuite({
        chaincode_version_id: versionId,
        test_type: 'unit',
      })
    },
    onSuccess: (suite) => {
      setTestSuite(suite)
      queryClient.invalidateQueries({ queryKey: ['test-suites', versionId] })
    },
  })

  // Phase 4: Create approval request mutation
  const createApprovalRequestMutation = useMutation({
    mutationFn: async () => {
      if (!versionId) throw new Error('Version ID is required')
      return approvalService.createRequest({
        chaincode_version_id: versionId,
        operation: 'commit',
        reason: `Approval request for ${name} v${version} on ${channelName}`,
      })
    },
    onSuccess: (request) => {
      setApprovalRequest(request)
      queryClient.invalidateQueries({ queryKey: ['approval-requests', versionId] })
    },
  })

  // Commit mutation
  const commitMutation = useMutation({
    mutationFn: (request: CommitChaincodeRequest) => chaincodeService.commit(request),
    onSuccess: () => {
      toast.success(`Deploy thành công! Chaincode ${name} v${version} đã được commit trên ${channelName}`)
      setStep('success')
      queryClient.invalidateQueries({ queryKey: ['chaincodes-committed'] })
      queryClient.invalidateQueries({ queryKey: ['chaincodes-installed'] })
    },
    onError: (error: Error) => {
      // Handle Phase 4 errors: approval required or tests failed
      const errorMessage = error.message || ''
      if (errorMessage.includes('approval required')) {
        toast.error('Cần approval trước khi commit. Vui lòng tạo approval request.')
        // Extract approval request ID if available
        const match = errorMessage.match(/approval request ([a-f0-9-]+)/i)
        if (match && match[1]) {
          // Fetch approval request details
          approvalService.getRequest(match[1]).then(setApprovalRequest).catch(console.error)
        }
      } else if (errorMessage.includes('tests failed')) {
        toast.error('Tests failed. Vui lòng chạy lại tests và đảm bảo tất cả pass.')
        // Extract test suite ID if available
        const match = errorMessage.match(/test suite: ([a-f0-9-]+)/i)
        if (match && match[1]) {
          // Fetch test suite details
          testingService.getTestSuite(match[1]).then(setTestSuite).catch(console.error)
        }
      } else {
        toast.error(`Lỗi commit: ${errorMessage || 'Không thể commit chaincode'}`)
      }
      console.error('[DeployChaincodeModal] Commit error:', error)
    },
  })

  const handleFileSelect = (file: File) => {
    // Validate package file
    const validation = validateChaincodePackage(file)
    if (!validation.valid) {
      toast.error(validation.error || 'File không hợp lệ')
      return
    }

    // Get warnings (non-blocking)
    const warnings = getPackageWarnings(file)
    setPackageWarnings(warnings)
    if (warnings.length > 0) {
      warnings.forEach(warning => {
        toast(warning, { icon: '⚠️', duration: 3000 })
      })
    }

    setSelectedFile(file)
    // Auto-fill label from filename if not set
    if (!label && file.name) {
      const nameWithoutExt = file.name.replace(/\.(tar\.gz|gz)$/, '')
      setLabel(nameWithoutExt)
    }
    // Auto-suggest package path based on filename
    if (!packagePath && file.name) {
      const fileName = file.name
      // Suggest common paths
      const suggestedPaths = [
        `/opt/chaincode/${fileName}`,
        `/var/chaincode/${fileName}`,
        `/chaincode/${fileName}`,
        `./${fileName}`,
      ]
      // Use first suggestion as default
      setPackagePath(suggestedPaths[0])
    }
  }

  const handleFileRemove = () => {
    setSelectedFile(null)
    setPackageWarnings([])
  }

  const handleInstall = () => {
    // Validation
    if (!selectedFile && !packagePath) {
      toast.error('Vui lòng chọn file hoặc nhập đường dẫn package')
      return
    }

    // If file is selected, upload first then install
    if (selectedFile) {
      // Validate file size (max 100MB)
      const maxSize = 100 * 1024 * 1024 // 100MB
      if (selectedFile.size > maxSize) {
        toast.error('File quá lớn. Kích thước tối đa là 100MB')
        return
      }
      uploadMutation.mutate(selectedFile)
    } else if (packagePath) {
      // Fallback: use direct path if provided
      installMutation.mutate({
        packagePath,
        label: label || undefined,
      })
    }
  }

  const handleApprove = () => {
    // Validation
    if (!name || !name.trim()) {
      toast.error('Vui lòng nhập chaincode name')
      return
    }
    if (!version || !version.trim()) {
      toast.error('Vui lòng nhập version')
      return
    }
    if (!sequence || isNaN(parseInt(sequence, 10)) || parseInt(sequence, 10) < 1) {
      toast.error('Vui lòng nhập sequence hợp lệ (>= 1)')
      return
    }
    if (!channelName || !channelName.trim()) {
      toast.error('Vui lòng nhập channel name')
      return
    }

    approveMutation.mutate({
      channelName: channelName.trim(),
      name: name.trim(),
      version: version.trim(),
      sequence: parseInt(sequence, 10),
      packageId: packageId || undefined,
      initRequired,
      endorsementPlugin: endorsementPlugin || undefined,
      validationPlugin: validationPlugin || undefined,
    })
  }

  const handleCommit = () => {
    // Validation
    if (!name || !name.trim()) {
      toast.error('Vui lòng nhập chaincode name')
      return
    }
    if (!version || !version.trim()) {
      toast.error('Vui lòng nhập version')
      return
    }
    if (!sequence || isNaN(parseInt(sequence, 10)) || parseInt(sequence, 10) < 1) {
      toast.error('Vui lòng nhập sequence hợp lệ (>= 1)')
      return
    }
    if (!channelName || !channelName.trim()) {
      toast.error('Vui lòng nhập channel name')
      return
    }

    commitMutation.mutate({
      channelName: channelName.trim(),
      name: name.trim(),
      version: version.trim(),
      sequence: parseInt(sequence, 10),
      initRequired,
      endorsementPlugin: endorsementPlugin || undefined,
      validationPlugin: validationPlugin || undefined,
    })
  }

  const handleClose = () => {
    if (step === 'success') {
      // Reset form
      setStep('install')
      setPackagePath('')
      setLabel('')
      setPackageId('')
      setSelectedFile(null)
      setName('')
      setVersion('')
      setSequence('1')
      setInitRequired(false)
    }
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
              Deploy Chaincode
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              {step === 'install' && 'Bước 1: Install chaincode package'}
              {step === 'approve' && 'Bước 2: Approve chaincode definition'}
              {step === 'commit' && 'Bước 3: Commit chaincode definition'}
              {step === 'pending-approval' && 'Chờ Admin Phê Duyệt'}
              {step === 'success' && 'Deploy thành công!'}
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={handleClose}>
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Step Indicator */}
          <div className="relative">
            <div className="flex items-center justify-between mb-2">
              <div className="flex flex-col items-center flex-1">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${
                  step === 'install' ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/50' : 
                  ['approve', 'commit', 'success', 'pending-approval'].includes(step) ? 'bg-green-500 text-white shadow-lg shadow-green-500/50' : 
                  'bg-gray-300 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                }`}>
                  {['approve', 'commit', 'success', 'pending-approval'].includes(step) ? (
                    <CheckCircle className="w-5 h-5" />
                  ) : (
                    <span className="text-sm font-semibold">1</span>
                  )}
                </div>
                <span className={`text-xs font-medium mt-2 ${
                  step === 'install' ? 'text-blue-500 dark:text-blue-400' : 
                  ['approve', 'commit', 'success'].includes(step) ? 'text-green-500 dark:text-green-400' : 
                  'text-gray-500 dark:text-gray-400'
                }`}>
                  Install
                </span>
              </div>
              
              <div className={`flex-1 h-0.5 mx-4 transition-all ${
                ['approve', 'commit', 'success'].includes(step) 
                  ? 'bg-green-500' 
                  : 'bg-gray-300 dark:bg-gray-700'
              }`} />
              
              <div className="flex flex-col items-center flex-1">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${
                  step === 'approve' ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/50' : 
                  ['commit', 'success'].includes(step) ? 'bg-green-500 text-white shadow-lg shadow-green-500/50' : 
                  'bg-gray-300 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                }`}>
                  {['commit', 'success'].includes(step) ? (
                    <CheckCircle className="w-5 h-5" />
                  ) : (
                    <span className="text-sm font-semibold">2</span>
                  )}
                </div>
                <span className={`text-xs font-medium mt-2 ${
                  step === 'approve' ? 'text-blue-500 dark:text-blue-400' : 
                  ['commit', 'success'].includes(step) ? 'text-green-500 dark:text-green-400' : 
                  'text-gray-500 dark:text-gray-400'
                }`}>
                  Approve
                </span>
              </div>
              
              <div className={`flex-1 h-0.5 mx-4 transition-all ${
                step === 'success' 
                  ? 'bg-green-500' 
                  : 'bg-gray-300 dark:bg-gray-700'
              }`} />
              
              <div className="flex flex-col items-center flex-1">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${
                  step === 'commit' ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/50' : 
                  step === 'success' ? 'bg-green-500 text-white shadow-lg shadow-green-500/50' : 
                  'bg-gray-300 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                }`}>
                  {step === 'success' ? (
                    <CheckCircle className="w-5 h-5" />
                  ) : (
                    <span className="text-sm font-semibold">3</span>
                  )}
                </div>
                <span className={`text-xs font-medium mt-2 ${
                  step === 'commit' ? 'text-blue-500 dark:text-blue-400' : 
                  step === 'success' ? 'text-green-500 dark:text-green-400' : 
                  'text-gray-500 dark:text-gray-400'
                }`}>
                  Commit
                </span>
              </div>
            </div>
          </div>

          {/* Install Step */}
          {step === 'install' && (
            <div className="space-y-6">
              <div>
                <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  Chaincode Package File <span className="text-red-500">*</span>
                </label>
                <FileUpload
                  accept=".tar.gz,.gz"
                  maxSize={100}
                  onFileSelect={handleFileSelect}
                  onFileRemove={handleFileRemove}
                  selectedFile={selectedFile}
                />
                <div className="mt-3 space-y-2">
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Chọn file chaincode package từ máy tính của bạn. File sẽ được tự động upload và install.
                  </p>
                  
                  {/* User Guidance */}
                  <details className="group">
                    <summary className="text-xs text-blue-400 dark:text-blue-500 cursor-pointer hover:text-blue-500 dark:hover:text-blue-400 flex items-center gap-1">
                      <Info className="w-3 h-3" />
                      <span>Hướng dẫn tạo chaincode package đúng format</span>
                    </summary>
                    <div className="mt-2 p-3 bg-blue-500/10 border-l-4 border-blue-500 rounded-lg">
                      <p className="text-xs text-blue-200 dark:text-blue-300 mb-2">
                        <strong>Chaincode package phải được tạo bằng lệnh:</strong>
                      </p>
                      <code className="block text-xs bg-blue-500/20 px-3 py-2 rounded mb-2 font-mono text-blue-100">
                        peer lifecycle chaincode package mychaincode.tar.gz --path ./chaincode --lang golang --label mychaincode
                      </code>
                      <p className="text-xs text-blue-200/80 dark:text-blue-300/80 mb-1">
                        <strong>Package phải chứa:</strong>
                      </p>
                      <ul className="text-xs text-blue-200/70 dark:text-blue-300/70 list-disc list-inside space-y-1 ml-2">
                        <li><code className="bg-blue-500/20 px-1 rounded">metadata.json</code> - Chaincode metadata</li>
                        <li><code className="bg-blue-500/20 px-1 rounded">code.tar.gz</code> - Compiled chaincode source</li>
                      </ul>
                      <p className="text-xs text-yellow-300/80 dark:text-yellow-400/80 mt-2">
                        ⚠️ <strong>Lưu ý:</strong> Không upload file tar.gz thông thường. Package phải được tạo bằng peer CLI.
                      </p>
                    </div>
                  </details>
                </div>
                {selectedFile && (
                  <div className="mt-3 space-y-2">
                    <div className="p-3 bg-green-500/10 border-l-4 border-green-500 rounded-lg">
                      <div className="flex items-start gap-2">
                        <CheckCircle className="w-4 h-4 text-green-400 flex-shrink-0 mt-0.5" />
                        <div className="flex-1">
                          <p className="text-xs font-semibold text-green-300 mb-1">
                            File đã sẵn sàng để upload:
                          </p>
                          <p className="text-xs text-green-200">
                            <code className="font-mono bg-green-500/20 px-2 py-1 rounded">{selectedFile.name}</code>
                            {' '}({(selectedFile.size / (1024 * 1024)).toFixed(2)} MB)
                          </p>
                          <p className="text-xs text-green-300/80 mt-1">
                            ✅ File sẽ được tự động upload và install khi bạn nhấn "Install"
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* Package Warnings */}
                    {packageWarnings.length > 0 && (
                      <div className="p-3 bg-yellow-500/10 border-l-4 border-yellow-500 rounded-lg">
                        <div className="flex items-start gap-2">
                          <AlertCircle className="w-4 h-4 text-yellow-400 flex-shrink-0 mt-0.5" />
                          <div className="flex-1">
                            <p className="text-xs font-semibold text-yellow-300 mb-1">
                              Cảnh báo:
                            </p>
                            <ul className="text-xs text-yellow-200/80 list-disc list-inside space-y-1">
                              {packageWarnings.map((warning, idx) => (
                                <li key={idx}>{warning}</li>
                              ))}
                            </ul>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* Optional: Direct path input (for advanced users) */}
              <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
                <details className="group">
                  <summary className="text-sm font-medium text-gray-600 dark:text-gray-400 cursor-pointer hover:text-gray-900 dark:hover:text-gray-200">
                    Advanced: Sử dụng đường dẫn trực tiếp (Optional)
                  </summary>
                  <div className="mt-3 space-y-2">
                    <Input
                      placeholder="/opt/chaincode/teaTraceCC.tar.gz"
                      value={packagePath}
                      onChange={(e) => setPackagePath(e.target.value)}
                      className="font-mono text-sm"
                    />
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Chỉ sử dụng nếu file đã có sẵn trên server peer. Nếu đã chọn file ở trên, file sẽ được upload tự động.
                    </p>
                  </div>
                </details>
              </div>

              <div>
                <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  Label <span className="text-xs font-normal text-gray-500">(Optional)</span>
                </label>
                <Input
                  placeholder="teaTraceCC_1.0"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                  Label cho chaincode package. Nếu để trống, sẽ tự động điền từ tên file.
                </p>
              </div>

              {(uploadMutation.isError || installMutation.isError) && (
                <div className="p-4 bg-red-500/10 border-l-4 border-red-500 rounded-lg">
                  <div className="flex items-start gap-3">
                    <X className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-red-400 mb-1">
                        {uploadMutation.isError ? 'Lỗi khi upload file' : 'Lỗi khi install chaincode'}
                      </p>
                      <p className="text-xs text-red-300">
                        {extractErrorMessage(uploadMutation.error || installMutation.error)}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
                <Button variant="secondary" onClick={handleClose} disabled={installMutation.isPending}>
                  Hủy
                </Button>
                <Button
                  variant="primary"
                  onClick={handleInstall}
                  disabled={(!selectedFile && !packagePath) || uploadMutation.isPending || installMutation.isPending}
                  className="min-w-[120px]"
                >
                  {(uploadMutation.isPending || installMutation.isPending) ? (
                    <>
                      <Loader className="w-4 h-4 mr-2 animate-spin" />
                      {uploadMutation.isPending ? 'Đang upload...' : 'Đang install...'}
                    </>
                  ) : (
                    <>
                      <Package className="w-4 h-4 mr-2" />
                      {selectedFile ? 'Upload & Install' : 'Install'}
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Approve Step */}
          {step === 'approve' && (
            <div className="space-y-6">
              {packageId && (
                <div className="p-4 bg-blue-500/10 border-l-4 border-blue-500 rounded-lg">
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-5 h-5 text-blue-400" />
                    <div>
                      <p className="text-sm font-semibold text-blue-300 mb-1">
                        Chaincode đã được install thành công
                      </p>
                      <p className="text-xs text-blue-200">
                        Package ID: <code className="font-mono bg-blue-500/20 px-2 py-1 rounded">{packageId}</code>
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Channel Name <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={channelName}
                    onChange={(e) => setChannelName(e.target.value)}
                    className="font-mono"
                  />
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Chaincode Name <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="teaTraceCC"
                    required
                  />
                  {name && !/^[a-zA-Z0-9_-]+$/.test(name) && (
                    <p className="text-xs text-yellow-400 mt-1 flex items-center gap-1">
                      <Info className="w-3 h-3" />
                      Tên chỉ nên chứa chữ cái, số, dấu gạch dưới và dấu gạch ngang
                    </p>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Version <span className="text-red-500">*</span>
                  </label>
                  <Input
                    value={version}
                    onChange={(e) => setVersion(e.target.value)}
                    placeholder="1.0"
                    required
                  />
                  {version && !/^[0-9]+\.[0-9]+(\.[0-9]+)?$/.test(version) && (
                    <p className="text-xs text-yellow-400 mt-1 flex items-center gap-1">
                      <Info className="w-3 h-3" />
                      Format khuyến nghị: x.y hoặc x.y.z (ví dụ: 1.0, 1.0.0)
                    </p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Sequence <span className="text-red-500">*</span>
                  </label>
                  <Input
                    type="number"
                    value={sequence}
                    onChange={(e) => {
                      const val = e.target.value
                      if (val === '' || (parseInt(val, 10) >= 1)) {
                        setSequence(val)
                      }
                    }}
                    min="1"
                    required
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Sequence number cho chaincode definition (phải &gt;= 1)
                  </p>
                  {sequence && (isNaN(parseInt(sequence, 10)) || parseInt(sequence, 10) < 1) && (
                    <p className="text-xs text-red-400 mt-1">Sequence phải là số &gt;= 1</p>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Endorsement Plugin
                  </label>
                  <Input
                    value={endorsementPlugin}
                    onChange={(e) => setEndorsementPlugin(e.target.value)}
                    placeholder="escc"
                  />
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                    Validation Plugin
                  </label>
                  <Input
                    value={validationPlugin}
                    onChange={(e) => setValidationPlugin(e.target.value)}
                    placeholder="vscc"
                  />
                </div>
              </div>

              <div className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
                <input
                  type="checkbox"
                  id="initRequired"
                  checked={initRequired}
                  onChange={(e) => setInitRequired(e.target.checked)}
                  className="w-4 h-4 rounded border-gray-300 dark:border-gray-600 text-blue-500 focus:ring-blue-500"
                />
                <label htmlFor="initRequired" className="text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
                  Init Required - Chaincode cần gọi hàm init trước khi sử dụng
                </label>
              </div>

              {approveMutation.isError && (
                <div className="p-4 bg-red-500/10 border-l-4 border-red-500 rounded-lg">
                  <div className="flex items-start gap-3">
                    <X className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-red-400 mb-1">
                        Lỗi khi approve chaincode
                      </p>
                      <p className="text-xs text-red-300">
                        {approveMutation.error instanceof Error 
                          ? approveMutation.error.message 
                          : 'Không thể approve chaincode. Vui lòng kiểm tra lại thông tin và thử lại.'}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
                <Button variant="secondary" onClick={() => setStep('install')} disabled={approveMutation.isPending}>
                  Quay lại
                </Button>
                <Button
                  variant="primary"
                  onClick={handleApprove}
                  disabled={!name || !version || !sequence || approveMutation.isPending}
                  className="min-w-[120px]"
                >
                  {approveMutation.isPending ? (
                    <>
                      <Loader className="w-4 h-4 mr-2 animate-spin" />
                      Đang approve...
                    </>
                  ) : (
                    <>
                      <CheckCircle className="w-4 h-4 mr-2" />
                      Approve
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Commit Step */}
          {step === 'commit' && (
            <div className="space-y-6">
              <div className="p-4 bg-green-500/10 border-l-4 border-green-500 rounded-lg">
                <div className="flex items-center gap-2">
                  <CheckCircle className="w-5 h-5 text-green-400" />
                  <p className="text-sm font-semibold text-green-300">
                    Chaincode đã được approve thành công!
                  </p>
                </div>
              </div>

              <div className="p-5 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 space-y-3">
                <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  Thông tin Chaincode Definition
                </h4>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <span className="text-xs text-gray-500 dark:text-gray-400">Channel:</span>
                    <p className="text-sm font-medium text-gray-900 dark:text-white mt-1">{channelName}</p>
                  </div>
                  <div>
                    <span className="text-xs text-gray-500 dark:text-gray-400">Name:</span>
                    <p className="text-sm font-medium text-gray-900 dark:text-white mt-1">{name}</p>
                  </div>
                  <div>
                    <span className="text-xs text-gray-500 dark:text-gray-400">Version:</span>
                    <p className="text-sm font-medium text-gray-900 dark:text-white mt-1">{version}</p>
                  </div>
                  <div>
                    <span className="text-xs text-gray-500 dark:text-gray-400">Sequence:</span>
                    <p className="text-sm font-medium text-gray-900 dark:text-white mt-1">{sequence}</p>
                  </div>
                </div>
              </div>

              {/* Phase 4: Approval Status */}
              {approvalRequest && (
                <div className={`p-4 border-l-4 rounded-lg ${
                  approvalRequest.status === 'approved' 
                    ? 'bg-green-500/10 border-green-500' 
                    : approvalRequest.status === 'rejected'
                    ? 'bg-red-500/10 border-red-500'
                    : 'bg-yellow-500/10 border-yellow-500'
                }`}>
                  <div className="flex items-start gap-3">
                    <Shield className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                      approvalRequest.status === 'approved' ? 'text-green-400' :
                      approvalRequest.status === 'rejected' ? 'text-red-400' :
                      'text-yellow-400'
                    }`} />
                    <div className="flex-1">
                      <p className={`text-sm font-semibold mb-1 ${
                        approvalRequest.status === 'approved' ? 'text-green-300' :
                        approvalRequest.status === 'rejected' ? 'text-red-300' :
                        'text-yellow-300'
                      }`}>
                        Approval Status: {approvalRequest.status.toUpperCase()}
                      </p>
                      {approvalRequest.status === 'pending' && (
                        <p className="text-xs text-yellow-200 mb-2">
                          Approval request đang chờ phê duyệt. Bạn cần approval trước khi commit.
                        </p>
                      )}
                      {approvalRequest.status === 'rejected' && (
                        <p className="text-xs text-red-200">
                          Approval request đã bị từ chối. Vui lòng tạo request mới.
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* Phase 4: Test Results */}
              {testSuite && (
                <div className={`p-4 border-l-4 rounded-lg ${
                  testSuite.status === 'passed' 
                    ? 'bg-green-500/10 border-green-500' 
                    : testSuite.status === 'failed'
                    ? 'bg-red-500/10 border-red-500'
                    : testSuite.status === 'running'
                    ? 'bg-blue-500/10 border-blue-500'
                    : 'bg-gray-500/10 border-gray-500'
                }`}>
                  <div className="flex items-start gap-3">
                    <TestTube className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                      testSuite.status === 'passed' ? 'text-green-400' :
                      testSuite.status === 'failed' ? 'text-red-400' :
                      testSuite.status === 'running' ? 'text-blue-400' :
                      'text-gray-400'
                    }`} />
                    <div className="flex-1">
                      <p className={`text-sm font-semibold mb-1 ${
                        testSuite.status === 'passed' ? 'text-green-300' :
                        testSuite.status === 'failed' ? 'text-red-300' :
                        testSuite.status === 'running' ? 'text-blue-300' :
                        'text-gray-300'
                      }`}>
                        Test Status: {testSuite.status.toUpperCase()}
                      </p>
                      {testSuite.status === 'passed' && (
                        <p className="text-xs text-green-200">
                          ✓ {testSuite.passed_tests}/{testSuite.total_tests} tests passed
                        </p>
                      )}
                      {testSuite.status === 'failed' && (
                        <div>
                          <p className="text-xs text-red-200 mb-1">
                            ✗ {testSuite.failed_tests}/{testSuite.total_tests} tests failed
                          </p>
                          {testSuite.error_message && (
                            <p className="text-xs text-red-300 mt-1">
                              {testSuite.error_message}
                            </p>
                          )}
                        </div>
                      )}
                      {testSuite.status === 'running' && (
                        <p className="text-xs text-blue-200">
                          Tests đang chạy...
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* Phase 4: Actions for approval and testing (optional) */}
              {step === 'commit' && versionId && (!approvalRequest || approvalRequest.status !== 'approved' || !testSuite || testSuite.status !== 'passed') && (
                <div className="p-4 bg-blue-500/10 border-l-4 border-blue-500 rounded-lg">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-blue-300 mb-2">
                        Phase 4: Approval và Tests (Tùy chọn)
                      </p>
                      <p className="text-xs text-blue-200 mb-2">
                        Các tính năng này là tùy chọn. Bạn có thể commit trực tiếp nếu không cần approval/tests.
                      </p>
                      <div className="flex gap-2 mt-2">
                        {(!approvalRequest || approvalRequest.status !== 'approved') && (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => {
                              try {
                                createApprovalRequestMutation.mutate()
                              } catch (error) {
                                console.warn('Approval service not available:', error)
                              }
                            }}
                            disabled={createApprovalRequestMutation.isPending || !versionId}
                          >
                            {createApprovalRequestMutation.isPending ? (
                              <>
                                <Loader className="w-3 h-3 mr-1 animate-spin" />
                                Đang tạo...
                              </>
                            ) : (
                              <>
                                <Shield className="w-3 h-3 mr-1" />
                                Tạo Approval Request
                              </>
                            )}
                          </Button>
                        )}
                        {(!testSuite || testSuite.status !== 'passed') && (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => {
                              try {
                                runTestsMutation.mutate()
                              } catch (error) {
                                console.warn('Testing service not available:', error)
                              }
                            }}
                            disabled={runTestsMutation.isPending || !versionId}
                          >
                            {runTestsMutation.isPending ? (
                              <>
                                <Loader className="w-3 h-3 mr-1 animate-spin" />
                                Đang chạy...
                              </>
                            ) : (
                              <>
                                <TestTube className="w-3 h-3 mr-1" />
                                Chạy Tests
                              </>
                            )}
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {commitMutation.isError && (
                <div className="p-4 bg-red-500/10 border-l-4 border-red-500 rounded-lg">
                  <div className="flex items-start gap-3">
                    <X className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-red-400 mb-1">
                        Lỗi khi commit chaincode
                      </p>
                      <p className="text-xs text-red-300">
                        {commitMutation.error instanceof Error 
                          ? commitMutation.error.message 
                          : 'Không thể commit chaincode. Vui lòng kiểm tra lại và thử lại.'}
                      </p>
                      {(commitMutation.error as Error)?.message?.includes('approval required') && (
                        <p className="text-xs text-yellow-300 mt-2">
                          💡 Vui lòng tạo approval request và chờ phê duyệt trước khi commit.
                        </p>
                      )}
                      {(commitMutation.error as Error)?.message?.includes('tests failed') && (
                        <p className="text-xs text-yellow-300 mt-2">
                          💡 Vui lòng chạy tests và đảm bảo tất cả tests pass trước khi commit.
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
                <Button variant="secondary" onClick={() => setStep('approve')} disabled={commitMutation.isPending}>
                  Quay lại
                </Button>
                <Button
                  variant="primary"
                  onClick={handleCommit}
                  disabled={
                    commitMutation.isPending || 
                    (approvalRequest !== null && approvalRequest.status === 'rejected') ||
                    (testSuite !== null && testSuite.status === 'failed')
                  }
                  className="min-w-[120px]"
                >
                  {commitMutation.isPending ? (
                    <>
                      <Loader className="w-4 h-4 mr-2 animate-spin" />
                      Đang commit...
                    </>
                  ) : (
                    <>
                      <Upload className="w-4 h-4 mr-2" />
                      Commit
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Pending Approval Step */}
          {step === 'pending-approval' && (
            <div className="space-y-6">
              <div className="bg-yellow-500/10 border-l-4 border-yellow-500 rounded-lg p-6">
                <div className="flex items-start gap-3">
                  <Clock className="w-6 h-6 text-yellow-400 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <h3 className="font-semibold text-yellow-400 mb-2 text-lg">
                      ⏳ Chờ Admin Phê Duyệt
                    </h3>
                    <p className="text-sm text-gray-300 mb-4">
                      Chaincode đã được install thành công trên peer. Để tiếp tục approve và commit, 
                      cần có sự phê duyệt từ admin.
                    </p>
                    
                    <div className="bg-black/20 p-4 rounded-lg space-y-3">
                      <div className="grid grid-cols-2 gap-3 text-xs">
                        <div>
                          <span className="text-gray-400 block mb-1">Chaincode Name:</span>
                          <span className="text-white font-medium">{name || label.split('_')[0] || 'N/A'}</span>
                        </div>
                        <div>
                          <span className="text-gray-400 block mb-1">Version:</span>
                          <span className="text-white font-medium">{version || 'N/A'}</span>
                        </div>
                      </div>
                      
                      <div className="text-xs">
                        <span className="text-gray-400 block mb-1">Package ID:</span>
                        <code className="text-green-200 font-mono text-[10px] break-all">
                          {packageId}
                        </code>
                      </div>
                      
                      {approvalRequest && (
                        <div className="text-xs border-t border-gray-700 pt-3 mt-3">
                          <span className="text-gray-400 block mb-1">Approval Request ID:</span>
                          <code className="text-blue-200 font-mono text-[10px]">
                            {approvalRequest.id}
                          </code>
                          <div className="mt-2 flex items-center gap-2">
                            <span className="text-gray-400">Status:</span>
                            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${
                              approvalRequest.status === 'pending' ? 'bg-yellow-500/20 text-yellow-300' :
                              approvalRequest.status === 'approved' ? 'bg-green-500/20 text-green-300' :
                              'bg-red-500/20 text-red-300'
                            }`}>
                              {approvalRequest.status.toUpperCase()}
                            </span>
                          </div>
                        </div>
                      )}
                    </div>

                    <div className="mt-4 p-3 bg-blue-500/10 border border-blue-500/20 rounded text-xs">
                      <p className="text-blue-200 flex items-start gap-2">
                        <Info className="w-4 h-4 flex-shrink-0 mt-0.5" />
                        <span>
                          Admin sẽ review chaincode này và phê duyệt trong vòng 24 giờ. 
                          Bạn sẽ nhận được thông báo qua email khi approval request được xử lý.
                        </span>
                      </p>
                    </div>

                    {!isAdmin && (
                      <div className="mt-4 text-xs text-gray-400">
                        <p>📧 Email thông báo sẽ được gửi đến: <strong className="text-gray-300">{user?.email || 'your email'}</strong></p>
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex justify-between gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
                <Button 
                  variant="secondary" 
                  onClick={() => setStep('install')}
                >
                  Quay lại Install
                </Button>
                <div className="flex gap-2">
                  {approvalRequest && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        // Copy approval request ID to clipboard
                        navigator.clipboard.writeText(approvalRequest.id)
                        toast.success('Approval Request ID copied to clipboard!')
                      }}
                    >
                      Copy Request ID
                    </Button>
                  )}
                  <Button 
                    variant="primary" 
                    onClick={handleClose}
                  >
                    Đóng
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* Success Step */}
          {step === 'success' && (
            <div className="text-center space-y-6 py-4">
              <div className="w-20 h-20 bg-gradient-to-br from-green-500 to-green-600 rounded-full flex items-center justify-center mx-auto shadow-lg shadow-green-500/50 animate-pulse">
                <CheckCircle className="w-12 h-12 text-white" />
              </div>
              <div>
                <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                  🎉 Deploy thành công!
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Chaincode <strong className="text-gray-900 dark:text-white">{name}</strong> version{' '}
                  <strong className="text-gray-900 dark:text-white">{version}</strong> đã được deploy thành công trên channel{' '}
                  <strong className="text-gray-900 dark:text-white">{channelName}</strong>
                </p>
              </div>
              
              {/* Deployment Summary */}
              <div className="p-5 bg-green-500/10 border border-green-500/20 rounded-lg text-left">
                <h4 className="text-sm font-semibold text-green-300 mb-3">Tóm tắt deployment:</h4>
                <div className="space-y-2 text-xs">
                  {packageId && (
                    <div className="flex justify-between">
                      <span className="text-gray-400">Package ID:</span>
                      <code className="text-green-200 font-mono">{packageId.substring(0, 30)}...</code>
                    </div>
                  )}
                  <div className="flex justify-between">
                    <span className="text-gray-400">Chaincode Name:</span>
                    <span className="text-white font-medium">{name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Version:</span>
                    <span className="text-white font-medium">{version}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Sequence:</span>
                    <span className="text-white font-medium">{sequence}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Channel:</span>
                    <span className="text-white font-medium">{channelName}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Init Required:</span>
                    <span className="text-white font-medium">{initRequired ? 'Yes' : 'No'}</span>
                  </div>
                </div>
              </div>

              <div className="flex justify-center gap-3 pt-4">
                <Button variant="primary" onClick={handleClose} className="min-w-[120px]">
                  Đóng
                </Button>
              </div>
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}

