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

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import type { ChannelInfo, NetworkOverview, OrdererInfo, PeerInfo } from '@shared/types/network.types'

type ApiResponse<T> = {
  success: boolean
  data: T
  message?: string
}

const normalizeChannel = (channel: any): ChannelInfo => ({
  name: channel?.name || 'unknown',
  peers: Array.isArray(channel?.peers) ? channel.peers : [],
  orderers: Array.isArray(channel?.orderers) ? channel.orderers : [],
  chaincodes: Array.isArray(channel?.chaincodes) ? channel.chaincodes : [],
  blockHeight: typeof channel?.blockHeight === 'number' ? channel.blockHeight : undefined,
})

const normalizePeer = (peer: any): PeerInfo => ({
  name: peer?.name || peer?.id || 'peer',
  address: peer?.address || peer?.endpoint || '',
  mspId: peer?.mspId || peer?.MSPID,
  channels: Array.isArray(peer?.channels) ? peer.channels : [],
  chaincodes: Array.isArray(peer?.chaincodes) ? peer.chaincodes : [],
  status: peer?.status || peer?.health || 'unknown',
  blockHeight: typeof peer?.blockHeight === 'number' ? peer.blockHeight : undefined,
})

const normalizeOrderer = (orderer: any): OrdererInfo => ({
  name: orderer?.name || 'orderer',
  address: orderer?.address || '',
  mspId: orderer?.mspId || orderer?.MSPID,
  status: orderer?.status || 'unknown',
  isLeader: Boolean(orderer?.isLeader),
})

export const networkService = {
  async getOverview(): Promise<NetworkOverview> {
    const response = await api.get<ApiResponse<any>>(API_ENDPOINTS.NETWORK.INFO)
    const payload = response.data?.data || {}

    const channels = Array.isArray(payload.channels) ? payload.channels.map(normalizeChannel) : []
    const peers = Array.isArray(payload.peers) ? payload.peers.map(normalizePeer) : []
    const orderers = Array.isArray(payload.orderers) ? payload.orderers.map(normalizeOrderer) : []
    const msps = Array.isArray(payload.msps) ? payload.msps : []

    return {
      channels,
      peers,
      orderers,
      msps,
    }
  },

  async listPeers(): Promise<PeerInfo[]> {
    const response = await api.get<ApiResponse<any[]>>(API_ENDPOINTS.NETWORK.PEERS)
    const peers = Array.isArray(response.data?.data) ? response.data.data : []
    return peers.map(normalizePeer)
  },

  async listOrderers(): Promise<OrdererInfo[]> {
    const response = await api.get<ApiResponse<any[]>>(API_ENDPOINTS.NETWORK.ORDERERS)
    const orderers = Array.isArray(response.data?.data) ? response.data.data : []
    return orderers.map(normalizeOrderer)
  },

  async listChannels(): Promise<ChannelInfo[]> {
    const response = await api.get<ApiResponse<any[]>>(API_ENDPOINTS.NETWORK.CHANNELS)
    const channels = Array.isArray(response.data?.data) ? response.data.data : []
    return channels.map(normalizeChannel)
  },

  async getChannelInfo(name: string): Promise<ChannelInfo> {
    const response = await api.get<ApiResponse<any>>(API_ENDPOINTS.NETWORK.CHANNEL_INFO(name))
    return normalizeChannel(response.data?.data)
  },
}


