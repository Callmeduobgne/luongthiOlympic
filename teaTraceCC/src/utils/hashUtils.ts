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

/**
 * Verify hash by comparing expected hash with provided input
 * @param expected - The hash value stored in the batch (already hashed)
 * @param provided - The raw input string to verify (will be hashed before comparison)
 * @returns true if the hash of provided input matches expected hash
 */
export function verifyHash(expected: string, provided: string): boolean {
  // Hash the provided input using SHA-256 (same algorithm as generateBatchHash)
  const providedHash = crypto.createHash("sha256").update(provided).digest("hex");
  return expected === providedHash;
}

