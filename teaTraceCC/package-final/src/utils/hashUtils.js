"use strict";

var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.createBatchHashPayload = createBatchHashPayload;
exports.generateBatchHash = generateBatchHash;
exports.verifyHash = verifyHash;
const crypto_1 = __importDefault(require("crypto"));
function createBatchHashPayload(payload) {
    const parts = [
        payload.batchId,
        payload.farmLocation,
        payload.harvestDate,
        payload.processingInfo,
        payload.qualityCert
    ];
    return parts.join("|");
}
function generateBatchHash(payload) {
    const raw = createBatchHashPayload(payload);
    return crypto_1.default.createHash("sha256").update(raw).digest("hex");
}
/**
 * Verify hash by comparing expected hash with provided input
 * @param expected - The hash value stored in the batch (already hashed)
 * @param provided - The raw input string to verify (will be hashed before comparison)
 * @returns true if the hash of provided input matches expected hash
 */
function verifyHash(expected, provided) {
    // Hash the provided input using SHA-256 (same algorithm as generateBatchHash)
    const providedHash = crypto_1.default.createHash("sha256").update(provided).digest("hex");
    return expected === providedHash;
}
//# sourceMappingURL=hashUtils.js.map