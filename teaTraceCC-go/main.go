
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
