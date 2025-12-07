/**
 * Blockchain-related types
 */

export interface Block {
  number: number
  hash: string
  previousHash: string
  dataHash: string
  timestamp: string
  transactionCount: number
  transactions?: Transaction[]
  channel?: string
}

export interface Transaction {
  txId: string
  timestamp: string
  type: string
  chaincodeName?: string
  functionName?: string
  args?: string[]
  creator?: {
    mspId: string
    certificate: string
  }
  validationCode?: number
  isValid?: boolean
}

export interface Channel {
  name: string
  blocks: number
  chaincodes: Chaincode[]
  peers: string[]
  orderers: string[]
}

export interface Chaincode {
  name: string
  version: string
  sequence: number
  channel: string
  packageId?: string
}

export interface Peer {
  name: string
  mspId: string
  endpoint: string
  status: 'online' | 'offline'
  height?: number
}

export interface Orderer {
  name: string
  endpoint: string
  status: 'online' | 'offline'
}

export interface NetworkInfo {
  name?: string
  channels?: Channel[]
  chaincodes?: Chaincode[]
  peers?: number
  orderers?: number
}

