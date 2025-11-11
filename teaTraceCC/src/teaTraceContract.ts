import { Context, Contract } from "fabric-contract-api";

import {
  CreateTeaBatchInput,
  TeaBatch,
  TeaBatchStatus,
  isTeaBatchStatus
} from "./models/teaBatch";
import { generateBatchHash, verifyHash } from "./utils/hashUtils";
import { getMSPIds } from "./models/mspConfig";

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
    this.ensureOrg(ctx, [MSP_CONFIG.VERIFIER, MSP_CONFIG.ADMIN, MSP_CONFIG.FARMER], "verify batches");

    const batch = await this.getBatchOrThrow(ctx, batchId);
    const isValid = verifyHash(batch.hashValue, hashInput);

    if (isValid && batch.status !== "VERIFIED") {
      batch.status = "VERIFIED";
      await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    }

    return { isValid, batch };
  }

  public async getBatchInfo(ctx: Context, batchId: string): Promise<TeaBatch> {
    const batch = await this.getBatchOrThrow(ctx, batchId);
    return batch;
  }

  public async updateBatchStatus(
    ctx: Context,
    batchId: string,
    status: string
  ): Promise<TeaBatch> {
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

