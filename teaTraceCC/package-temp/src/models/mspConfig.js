"use strict";
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
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.loadMSPConfig = loadMSPConfig;
exports.getMSPIds = getMSPIds;
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
let cachedConfig = null;
function loadMSPConfig() {
    if (cachedConfig) {
        return cachedConfig;
    }
    try {
        const configPath = path.join(process.cwd(), 'msp-config.json');
        const configData = fs.readFileSync(configPath, 'utf8');
        cachedConfig = JSON.parse(configData);
        console.log('🔐 MSP Configuration loaded:');
        console.log(`  - Farmer: ${cachedConfig.mspRoles.farmer.mspId}`);
        console.log(`  - Verifier: ${cachedConfig.mspRoles.verifier.mspId}`);
        console.log(`  - Admin: ${cachedConfig.mspRoles.admin.mspId}`);
        return cachedConfig;
    }
    catch (error) {
        console.warn('⚠️  msp-config.json not found, using default config');
        cachedConfig = {
            mspRoles: {
                farmer: { mspId: 'Org1MSP', description: 'Default farmer' },
                verifier: { mspId: 'Org2MSP', description: 'Default verifier' },
                admin: { mspId: 'Org3MSP', description: 'Default admin' }
            }
        };
        console.log('🔐 Using default MSP Configuration:');
        console.log(`  - Farmer: ${cachedConfig.mspRoles.farmer.mspId}`);
        console.log(`  - Verifier: ${cachedConfig.mspRoles.verifier.mspId}`);
        console.log(`  - Admin: ${cachedConfig.mspRoles.admin.mspId}`);
        return cachedConfig;
    }
}
function getMSPIds() {
    const config = loadMSPConfig();
    return {
        FARMER: config.mspRoles.farmer.mspId,
        VERIFIER: config.mspRoles.verifier.mspId,
        ADMIN: config.mspRoles.admin.mspId
    };
}
//# sourceMappingURL=mspConfig.js.map