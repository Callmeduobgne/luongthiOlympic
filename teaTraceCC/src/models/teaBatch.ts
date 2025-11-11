export type TeaBatchStatus = "CREATED" | "VERIFIED" | "EXPIRED";

export interface TeaBatch {
  batchId: string;
  farmLocation: string;
  harvestDate: string;
  processingInfo: string;
  qualityCert: string;
  hashValue: string;
  owner: string;
  timestamp: string;
  status: TeaBatchStatus;
}

export interface CreateTeaBatchInput {
  batchId: string;
  farmLocation: string;
  harvestDate: string;
  processingInfo: string;
  qualityCert: string;
}

export function isTeaBatchStatus(value: string): value is TeaBatchStatus {
  return value === "CREATED" || value === "VERIFIED" || value === "EXPIRED";
}

