// Copyright 2025 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"teaTraceCC/chaincode"
)

func main() {
	teaTraceChaincode, err := contractapi.NewChaincode(chaincode.NewTeaTraceContract())
	if err != nil {
		log.Panicf("Error creating teaTraceCC chaincode: %v", err)
	}

	// Check if running as external chaincode (CCaaS mode)
	ccAddress := os.Getenv("CHAINCODE_SERVER_ADDRESS")
	ccID := os.Getenv("CHAINCODE_ID")
	
	// Fallback to alternative env var names
	if ccID == "" {
		ccID = os.Getenv("CORE_CHAINCODE_ID_NAME")
	}

	if ccAddress != "" && ccID != "" {
		// CCaaS mode - run as external chaincode server
		log.Printf("Starting chaincode as external service at %s with ID %s", ccAddress, ccID)
		
		server := &shim.ChaincodeServer{
			CCID:    ccID,
			Address: ccAddress,
			CC:      teaTraceChaincode,
			TLSProps: shim.TLSProperties{
				Disabled: true,
			},
		}

		if err := server.Start(); err != nil {
			log.Panicf("Error starting chaincode server: %v", err)
		}
	} else {
		// Traditional mode - peer manages the chaincode
		log.Println("Starting chaincode in traditional mode (managed by peer)")
		if err := teaTraceChaincode.Start(); err != nil {
			log.Panicf("Error starting teaTraceCC chaincode: %v", err)
		}
	}
}
