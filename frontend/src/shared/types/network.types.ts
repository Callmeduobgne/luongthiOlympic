/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

export interface ChannelInfo {
  name: string
  peers: string[]
  orderers: string[]
  chaincodes: string[]
  blockHeight?: number
}

export interface PeerInfo {
  name: string
  url: string
  status: 'up' | 'down'
  chaincodes: string[]
  channels: string[]
  mspId?: string
  address?: string
  blockHeight?: number
}

export interface OrdererInfo {
  name: string
  url: string
  status: 'up' | 'down'
  channels: string[]
  isLeader?: boolean
  mspId?: string
  address?: string
}

export interface NetworkOverview {
  channels: ChannelInfo[]
  peers: PeerInfo[]
  orderers: OrdererInfo[]
  totalTransactions: number
  totalBlocks: number
  msps?: string[]
}

