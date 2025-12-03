/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

export interface Block {
  number: number
  hash: string
  previousHash: string
  dataHash: string
  timestamp: string
  transactions: Transaction[]
  channelId?: string
  channel?: string
  transactionCount?: number
}

export interface Transaction {
  txId: string
  timestamp: string
  channelId: string
  chaincodeName: string
  functionName: string
  args: string[]
  creator: string
  validationCode: number
  blockNumber: number
  isValid?: boolean
}

