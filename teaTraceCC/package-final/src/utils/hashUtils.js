"use strict";
/**
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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