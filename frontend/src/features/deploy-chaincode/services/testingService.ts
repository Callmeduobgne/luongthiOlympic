import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

export interface TestSuite {
  id: string
  chaincode_version_id: string
  name: string
  description?: string
  test_type: 'unit' | 'integration' | 'e2e'
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped'
  started_at?: string
  completed_at?: string
  duration_ms?: number
  total_tests: number
  passed_tests: number
  failed_tests: number
  skipped_tests: number
  output?: string
  error_message?: string
  error_code?: string
  metadata?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface TestCase {
  id: string
  test_suite_id: string
  name: string
  description?: string
  test_function?: string
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped'
  started_at?: string
  completed_at?: string
  duration_ms?: number
  output?: string
  error_message?: string
  error_stack?: string
  assertions?: Record<string, any>
  metadata?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface RunTestSuiteRequest {
  chaincode_version_id: string
  test_type: 'unit' | 'integration' | 'e2e'
  test_command?: string
  metadata?: Record<string, any>
}

export const testingService = {
  /**
   * Run a test suite
   */
  async runTestSuite(request: RunTestSuiteRequest): Promise<TestSuite> {
    const response = await api.post<{ success: boolean; data: TestSuite }>(
      API_ENDPOINTS.CHAINCODE.TESTING.RUN,
      request
    )
    return response.data.data
  },

  /**
   * Get test suite by ID
   */
  async getTestSuite(id: string): Promise<TestSuite> {
    const response = await api.get<{ success: boolean; data: TestSuite }>(
      API_ENDPOINTS.CHAINCODE.TESTING.GET_SUITE(id)
    )
    return response.data.data
  },

  /**
   * List test suites
   */
  async listTestSuites(filters?: {
    chaincode_version_id?: string
    status?: string
    test_type?: string
  }): Promise<TestSuite[]> {
    const params = new URLSearchParams()
    if (filters?.chaincode_version_id) params.append('chaincode_version_id', filters.chaincode_version_id)
    if (filters?.status) params.append('status', filters.status)
    if (filters?.test_type) params.append('test_type', filters.test_type)

    const queryString = params.toString()
    const url = queryString
      ? `${API_ENDPOINTS.CHAINCODE.TESTING.LIST_SUITES}?${queryString}`
      : API_ENDPOINTS.CHAINCODE.TESTING.LIST_SUITES

    const response = await api.get<{ success: boolean; data: { suites: TestSuite[]; count: number } }>(url)
    return response.data.data.suites || []
  },

  /**
   * Get test cases for a suite
   */
  async getTestCases(suiteId: string): Promise<TestCase[]> {
    const response = await api.get<{ success: boolean; data: { test_cases: TestCase[]; count: number } }>(
      API_ENDPOINTS.CHAINCODE.TESTING.GET_CASES(suiteId)
    )
    return response.data.data.test_cases || []
  },
}

