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

