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

import { ChaincodeServer } from "fabric-shim";
import { TeaTraceContract } from "./teaTraceContract";

// Start chaincode server for external mode
const server = new ChaincodeServer({
    chaincode: TeaTraceContract,
    chaincodeId: process.env.CHAINCODE_ID || process.env.CORE_CHAINCODE_ID_NAME || "teaTraceCC_1.0:latest",
    chaincodeAddress: process.env.CHAINCODE_SERVER_ADDRESS || "0.0.0.0:9999"
});

server.start();

