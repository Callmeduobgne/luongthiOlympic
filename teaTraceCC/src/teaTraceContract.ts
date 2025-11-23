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

import { Context, Contract } from "fabric-contract-api";

import {
  CreateTeaBatchInput,
  TeaBatch,
  TeaBatchStatus,
  isTeaBatchStatus
} from "./models/teaBatch";
import { generateBatchHash, verifyHash } from "./utils/hashUtils";
import { getMSPIds } from "./models/mspConfig";
import {
  validateBatchId,
  validateDate,
  validateString,
  validatePagination
} from "./utils/validation";

// Load MSP configuration
const MSP_CONFIG = getMSPIds();

export class TeaTraceContract extends Contract {
  constructor() {
    super("teaTraceContract");
  }

  public async createBatch(
    ctx: Context,
    batchId: string,
    farmLocation: string,
    harvestDate: string,
    processingInfo: string,
    qualityCert: string
  ): Promise<TeaBatch> {
    // Input validation
    validateBatchId(batchId);
    validateString(farmLocation, "Farm location", 200);
    validateDate(harvestDate);
    validateString(processingInfo, "Processing info", 1000);
    validateString(qualityCert, "Quality certificate", 100);

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER], "create batches");

    await this.assertBatchDoesNotExist(ctx, batchId);

    const owner =
      ctx.clientIdentity.getAttributeValue("owner") ||
      ctx.clientIdentity.getAttributeValue("organization") ||
      ctx.clientIdentity.getMSPID();

    const input: CreateTeaBatchInput = {
      batchId,
      farmLocation,
      harvestDate,
      processingInfo,
      qualityCert
    };

    const hashValue = generateBatchHash(input);

    const batch: TeaBatch = {
      ...input,
      hashValue,
      owner,
      timestamp: this.getCurrentTimestamp(ctx),
      status: "CREATED"
    };

    await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    return batch;
  }

  public async verifyBatch(
    ctx: Context,
    batchId: string,
    hashInput: string
  ): Promise<{ isValid: boolean; batch: TeaBatch }> {
    // Input validation
    validateBatchId(batchId);
    if (!hashInput || hashInput.trim().length === 0) {
      throw new Error("Hash input cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.VERIFIER, MSP_CONFIG.ADMIN, MSP_CONFIG.FARMER], "verify batches");

    const batch = await this.getBatchOrThrow(ctx, batchId);
    const isValid = verifyHash(batch.hashValue, hashInput);

    if (isValid && batch.status !== "VERIFIED") {
      batch.status = "VERIFIED";
      await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    }

    return { isValid, batch };
  }

  public async getBatchInfo(ctx: Context, batchId: string): Promise<TeaBatch | null> {
    validateBatchId(batchId);
    const buffer = await ctx.stub.getState(batchId);
    if (!buffer || buffer.length === 0) {
      // Return null instead of throwing error for better API compatibility
      return null;
    }
    return JSON.parse(this.bytesToString(buffer)) as TeaBatch;
  }

  /**
   * Get all batches with pagination
   * Args: [limit?, offset?]
   */
  public async getAllBatches(
    ctx: Context,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      total++;
      if (total > offsetNum && batches.length < limitNum) {
        const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
        batches.push(batch);
      }
    }

    return { batches, total };
  }

  /**
   * Get batches by status
   * Args: [status, limit?, offset?]
   */
  public async getBatchesByStatus(
    ctx: Context,
    status: string,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    const normalizedStatus = status.toUpperCase();
    if (!isTeaBatchStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`
      );
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      if (batch.status === normalizedStatus) {
        total++;
        if (total > offsetNum && batches.length < limitNum) {
          batches.push(batch);
        }
      }
    }

    return { batches, total };
  }

  /**
   * Get batches by owner
   * Args: [owner, limit?, offset?]
   */
  public async getBatchesByOwner(
    ctx: Context,
    owner: string,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    if (!owner || owner.trim().length === 0) {
      throw new Error("Owner cannot be empty");
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      if (batch.owner === owner) {
        total++;
        if (total > offsetNum && batches.length < limitNum) {
          batches.push(batch);
        }
      }
    }

    return { batches, total };
  }

  /**
   * Get batch history (all changes)
   */
  public async getBatchHistory(ctx: Context, batchId: string): Promise<TeaBatch[]> {
    validateBatchId(batchId);

    const historyIterator = await ctx.stub.getHistoryForKey(batchId);
    const history: TeaBatch[] = [];

    while (true) {
      const result = await historyIterator.next();
      if (result.done) {
        await historyIterator.close();
        break;
      }

      if (result.value.isDelete) {
        continue;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      history.push(batch);
    }

    return history.reverse(); // Oldest first
  }

  public async updateBatchStatus(
    ctx: Context,
    batchId: string,
    status: string
  ): Promise<TeaBatch> {
    // Input validation
    validateBatchId(batchId);
    if (!status || status.trim().length === 0) {
      throw new Error("Status cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update batch status");

    const normalizedStatus = status.toUpperCase();
    if (!isTeaBatchStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`
      );
    }

    const batch = await this.getBatchOrThrow(ctx, batchId);
    batch.status = normalizedStatus as TeaBatchStatus;
    batch.timestamp = this.getCurrentTimestamp(ctx);

    await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    return batch;
  }

  private ensureOrg(ctx: Context, allowedMsps: string[], action: string): void {
    const callerMsp = ctx.clientIdentity.getMSPID();
    if (!allowedMsps.includes(callerMsp)) {
      throw new Error(
        `MSP '${callerMsp}' is not authorized to ${action}. Allowed MSPs: ${allowedMsps.join(", ")}`
      );
    }
  }

  private async getBatchOrThrow(ctx: Context, batchId: string): Promise<TeaBatch> {
    const buffer = await ctx.stub.getState(batchId);
    if (!buffer || buffer.length === 0) {
      throw new Error(`Batch with id '${batchId}' does not exist.`);
    }

    return JSON.parse(this.bytesToString(buffer)) as TeaBatch;
  }

  private async assertBatchDoesNotExist(ctx: Context, batchId: string): Promise<void> {
    const buffer = await ctx.stub.getState(batchId);
    if (buffer && buffer.length > 0) {
      throw new Error(`Batch with id '${batchId}' already exists.`);
    }
  }

  private getCurrentTimestamp(ctx: Context): string {
    const timestamp = ctx.stub.getTxTimestamp();
    const millis = timestamp.seconds.toNumber() * 1000 + Math.floor(timestamp.nanos / 1_000_000);
    return new Date(millis).toISOString();
  }

  private bytesToString(bytes: Uint8Array): string {
    return Buffer.from(bytes).toString("utf8");
  }
}

export const contracts = [TeaTraceContract];

