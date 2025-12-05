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

