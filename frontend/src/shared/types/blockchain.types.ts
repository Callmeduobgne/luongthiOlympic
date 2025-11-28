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

