package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

// RawMaterialContract defines the basic structure for the smart contract
type RawMaterialContract struct {
	contractapi.Contract
}

// RawMaterial defines the structure of a raw material token
type RawMaterial struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Supplier    string `json:"supplier"`
	Manufacturer string `json:"manufacturer"`
}

// Init initializes the smart contract
func (r *RawMaterialContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Raw material smart contract initialized")
	return nil
}

// CreateRawMaterial creates a new raw material token by the supplier
func (r *RawMaterialContract) CreateRawMaterial(ctx contractapi.TransactionContextInterface, name string) (string, error) {

	// Use a incrementing ID to ensure unique raw material IDs
	rawMaterialIDbyte, err := ctx.GetStub().GetState("RawIDIterator")
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	rawMaterialID := string(rawMaterialIDbyte)
	if rawMaterialIDbyte == nil {
		rawMaterialID = "0"
	}

	rawMaterialID = "RawMaterial-" + rawMaterialID
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "SupplierMSP" {
		return "", fmt.Errorf("Client of org %s is not authorized to create a token", clientOrgID)
	}

	existing, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if existing != nil {
		return "", fmt.Errorf("raw material with ID %s already exists", rawMaterialID)
	}
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to Get ID: %v", err)
	}
	if existing != nil {
		return "", fmt.Errorf("raw material with ID %s already exists", rawMaterialID)
	}
	rawMaterial := RawMaterial{
		ID:          rawMaterialID,
		Name:        name,
		Status:      "New",
		Supplier:    id,
	}

	rawMaterialJSON, err := json.Marshal(rawMaterial)
	if err != nil {
		return "", fmt.Errorf("failed to marshal raw material JSON: %v", err)
	}

	err = ctx.GetStub().PutState(rawMaterialID, rawMaterialJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write to world state: %v", err)
	}

	rawMaterialIDInt, err := strconv.Atoi(rawMaterialID)
	if err != nil {
		return "", fmt.Errorf("failed to convert rawMaterialID to int: %v", err)
	}
	rawMaterialIDInt++
	rawMaterialIDnext := strconv.Itoa(rawMaterialIDInt)

	err = ctx.GetStub().PutState("RawIDIterator", []byte(rawMaterialIDnext))
	if err != nil {
		return "", fmt.Errorf("failed to write to world state: %v", err)
	}

	return rawMaterialID, nil
}

// InitiateTransferToManufacturer Initiates transfer of the raw material token to the manufacturer
func (r *RawMaterialContract) InitiateTransferToManufacturer(ctx contractapi.TransactionContextInterface, rawMaterialID string, manufacturerID string) error {
	rawMaterialID = "RawMaterial-" + rawMaterialID
	rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if rawMaterialJSON == nil {
		return fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
	}

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "SupplierMSP" {
		return fmt.Errorf("Client of org %s is not authorized to create a token", clientOrgID)
	}

	var rawMaterial RawMaterial
	err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
	}

	SupplierID, err := ctx.GetClientIdentity().GetID()
	if SupplierID != rawMaterial.Supplier {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}
	if rawMaterial.Status != "New" {
		return fmt.Errorf("raw material with ID %s cannot be transferred. Status: %s", rawMaterialID, rawMaterial.Status)
	}
	if err != nil {
		return fmt.Errorf("failed to Get ID: %v", err)
	}

	rawMaterial.Status = "Transferring"
	rawMaterial.Manufacturer = manufacturerID

	rawMaterialJSON, err = json.Marshal(rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to marshal updated raw material JSON: %v", err)
	}

	err = ctx.GetStub().PutState(rawMaterialID, rawMaterialJSON)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

// CompleteTransferToManufacturer Completes transfer of the raw material token to the manufacturer
func (r *RawMaterialContract) CompleteTransferToManufacturer(ctx contractapi.TransactionContextInterface, rawMaterialID string) error {
	rawMaterialID = "RawMaterial-" + rawMaterialID
	rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if rawMaterialJSON == nil {
		return fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
	}

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "ManufacturerMSP" {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}

	var rawMaterial RawMaterial
	err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
	}

	ManufacturerID, err := ctx.GetClientIdentity().GetID()
	if ManufacturerID != rawMaterial.Manufacturer {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}
	if err != nil {
		return fmt.Errorf("failed to Get ID: %v", err)
	}

	if rawMaterial.Status != "Transferring" {
		return fmt.Errorf("raw material with ID %s cannot be transferred. Status: %s", rawMaterialID, rawMaterial.Status)
	}

	rawMaterial.Status = "Transferred"

	rawMaterialJSON, err = json.Marshal(rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to marshal updated raw material JSON: %v", err)
	}

	err = ctx.GetStub().PutState(rawMaterialID, rawMaterialJSON)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

func (r *RawMaterialContract) GetRawMaterial(ctx contractapi.TransactionContextInterface, rawMaterialID string) (*RawMaterial, error) {
	rawMaterialID = "RawMaterial-" + rawMaterialID
	rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if rawMaterialJSON == nil {
		return nil, fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
	}

	var rawMaterial RawMaterial
	err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
	}

	return &rawMaterial, nil
}

func (r *RawMaterialContract) GetRawMaterialHistory(ctx contractapi.TransactionContextInterface, rawMaterialID string) ([]*RawMaterial, error) {
	rawMaterialID = "RawMaterial-" + rawMaterialID
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(rawMaterialID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	defer resultsIterator.Close()

	var rawMaterials []*RawMaterial
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}

		var rawMaterial RawMaterial
		err = json.Unmarshal(queryResponse.Value, &rawMaterial)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
		}
		rawMaterials = append(rawMaterials, &rawMaterial)
	}

	return rawMaterials, nil
}

func (r *RawMaterialContract) GetNumberOfRawMaterials(ctx contractapi.TransactionContextInterface) (int, error) {
	rawMaterialIDbyte, err := ctx.GetStub().GetState("RawIDIterator")
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	rawMaterialID := string(rawMaterialIDbyte)
	if rawMaterialIDbyte == nil {
		rawMaterialID = "0"
	}
	rawMaterialIDInt, err := strconv.Atoi(rawMaterialID)
	if err != nil {
		return 0, fmt.Errorf("failed to convert rawMaterialID to int: %v", err)
	}
	return rawMaterialIDInt, nil
}

func (r *RawMaterialContract) GetAllRawMaterials(ctx contractapi.TransactionContextInterface) ([]*RawMaterial, error) {
	rawMaterialIDbyte, err := ctx.GetStub().GetState("RawIDIterator")
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	rawMaterialIDStr := string(rawMaterialIDbyte)
	if rawMaterialIDbyte == nil {
		rawMaterialIDStr = "0"
	}
	rawMaterialIDInt, err := strconv.Atoi(rawMaterialIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert rawMaterialID to int: %v", err)
	}
	var rawMaterials []*RawMaterial

	for i := 0; i < rawMaterialIDInt; i++ {
		rawMaterialID := strconv.Itoa(i)
		rawMaterialID = "RawMaterial-" + rawMaterialID
		rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}
		if rawMaterialJSON == nil {
			return nil, fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
		}

		var rawMaterial RawMaterial
		err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
		}
		MSPID, err := ctx.GetClientIdentity().GetMSPID()
		if err != nil {
			return nil, fmt.Errorf("failed to Get MSPID: %v", err)
		}
		ID, err := ctx.GetClientIdentity().GetID()
		if err != nil {
			return nil, fmt.Errorf("failed to Get ID: %v", err)
		}
		if MSPID == "SupplierMSP" {
			if rawMaterial.Supplier == ID {
				rawMaterials = append(rawMaterials, &rawMaterial)
			}
		} else if MSPID == "ManufacturerMSP" {
			if rawMaterial.Manufacturer == ID {
				rawMaterials = append(rawMaterials, &rawMaterial)
			}
		} else {
			return nil, fmt.Errorf("Client of type %s is not authorized to perform this action", MSPID)
		}
	}

	return rawMaterials, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&RawMaterialContract{})
	if err != nil {
	  log.Panicf("Error creating chaincode: %v", err)
	}
  
	if err := assetChaincode.Start(); err != nil {
	  log.Panicf("Error starting chaincode: %v", err)
	}
  }
  