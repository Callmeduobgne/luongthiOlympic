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

import * as fs from 'fs';
import * as path from 'path';

export interface MSPRole {
  mspId: string;
  description?: string;
}

export interface MSPConfig {
  mspRoles: {
    farmer: MSPRole;
    verifier: MSPRole;
    admin: MSPRole;
  };
}

let cachedConfig: MSPConfig | null = null;

export function loadMSPConfig(): MSPConfig {
  if (cachedConfig) {
    return cachedConfig;
  }

  try {
    const configPath = path.join(process.cwd(), 'msp-config.json');
    const configData = fs.readFileSync(configPath, 'utf8');
    cachedConfig = JSON.parse(configData) as MSPConfig;
    console.log('🔐 MSP Configuration loaded:');
    console.log(`  - Farmer: ${cachedConfig.mspRoles.farmer.mspId}`);
    console.log(`  - Verifier: ${cachedConfig.mspRoles.verifier.mspId}`);
    console.log(`  - Admin: ${cachedConfig.mspRoles.admin.mspId}`);
    return cachedConfig;
  } catch (error) {
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

export function getMSPIds() {
  const config = loadMSPConfig();
  return {
    FARMER: config.mspRoles.farmer.mspId,
    VERIFIER: config.mspRoles.verifier.mspId,
    ADMIN: config.mspRoles.admin.mspId
  };
}

