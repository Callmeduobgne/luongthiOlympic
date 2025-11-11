import crypto from "crypto";

import { CreateTeaBatchInput, TeaBatch } from "../models/teaBatch";

export function createBatchHashPayload(
  payload: CreateTeaBatchInput | TeaBatch
): string {
  const parts = [
    payload.batchId,
    payload.farmLocation,
    payload.harvestDate,
    payload.processingInfo,
    payload.qualityCert
  ];

  return parts.join("|");
}

export function generateBatchHash(
  payload: CreateTeaBatchInput | TeaBatch
): string {
  const raw = createBatchHashPayload(payload);
  return crypto.createHash("sha256").update(raw).digest("hex");
}

export function verifyHash(expected: string, provided: string): boolean {
  return expected === provided;
}

