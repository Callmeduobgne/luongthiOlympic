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

import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AuthPage } from '@features/authentication/components/AuthPage'
import { Dashboard } from '@features/dashboard/components/Dashboard'
import { DeployChaincodePage } from '@features/deploy-chaincode/pages/DeployChaincodePage'
import { ExplorerPage } from '@features/explorer/pages/ExplorerPage'
import { AnalyticsPage } from '@features/analytics/pages/AnalyticsPage'
import { NetworkPage } from '@features/network/pages/NetworkPage'
import { SettingsPage } from '@features/settings/pages/SettingsPage'
import { TeaShopHomepage } from '@/pages/TeaShopHomepage'
import VerifyPackage from '@/pages/VerifyPackage'
import VerifyHashPage from '@/pages/VerifyHashPage'
import { ProtectedRoute } from './ProtectedRoute'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <TeaShopHomepage />,
  },
  {
    path: '/login',
    element: <AuthPage />,
  },
  {
    path: '/verify/packages/:packageId',
    element: <VerifyPackage />,
  },
  {
    path: '/verify/batches/:batchId',
    element: <VerifyPackage />,
  },
  {
    path: '/verify/hash',
    element: <VerifyHashPage />,
  },
  {
    path: '/dashboard',
    element: (
      <ProtectedRoute>
        <Dashboard />
      </ProtectedRoute>
    ),
  },
  {
    path: '/deploy-chaincode',
    element: (
      <ProtectedRoute>
        <DeployChaincodePage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/explorer',
    element: (
      <ProtectedRoute>
        <ExplorerPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/analytics',
    element: (
      <ProtectedRoute>
        <AnalyticsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/network',
    element: (
      <ProtectedRoute>
        <NetworkPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/settings',
    element: (
      <ProtectedRoute>
        <SettingsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
])


