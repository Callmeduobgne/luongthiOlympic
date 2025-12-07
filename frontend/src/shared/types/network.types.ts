export interface ChannelInfo {
  name: string
  peers: string[]
  orderers: string[]
  chaincodes: string[]
  blockHeight?: number
}

export interface PeerInfo {
  name: string
  address: string
  mspId?: string
  channels?: string[]
  chaincodes?: string[]
  status?: string
  blockHeight?: number
}

export interface OrdererInfo {
  name: string
  address: string
  mspId?: string
  status?: string
  isLeader?: boolean
}

export interface NetworkOverview {
  channels: ChannelInfo[]
  peers: PeerInfo[]
  orderers: OrdererInfo[]
  msps: string[]
}


