package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
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
	Lock 	  	string `json:"lock"`
}

// Init initializes the smart contract
func (r *RawMaterialContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Raw material smart contract initialized")
	return nil
}

// CreateRawMaterial creates a new raw material token by the supplier
func (r *RawMaterialContract) CreateRawMaterial(ctx contractapi.TransactionContextInterface, rawMaterialID string, name string) error {

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "SupplierMSP" {
		return fmt.Errorf("Client of org %s is not authorized to create a token", clientOrgID)
	}

	existing, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("raw material with ID %s already exists", rawMaterialID)
	}
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to Get ID: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("raw material with ID %s already exists", rawMaterialID)
	}
	rawMaterial := RawMaterial{
		ID:          rawMaterialID,
		Name:        name,
		Status:      "New",
		Supplier:    id,
	}

	rawMaterialJSON, err := json.Marshal(rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to marshal raw material JSON: %v", err)
	}

	err = ctx.GetStub().PutState(rawMaterialID, rawMaterialJSON)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

// TransferToManufacturer transfers the raw material token to the manufacturer
func (r *RawMaterialContract) TransferToManufacturer(ctx contractapi.TransactionContextInterface, rawMaterialID string, manufacturerID string) error {
	rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if rawMaterialJSON == nil {
		return fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
	}

	var rawMaterial RawMaterial
	err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
	if err != nil {
		return fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
	}

	if rawMaterial.Status != "New" {
		return fmt.Errorf("raw material with ID %s cannot be transferred. Status: %s", rawMaterialID, rawMaterial.Status)
	}

	rawMaterial.Status = "Received"
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

func main() {
	assetChaincode, err := contractapi.NewChaincode(&RawMaterialContract{})
	if err != nil {
	  log.Panicf("Error creating chaincode: %v", err)
	}
  
	if err := assetChaincode.Start(); err != nil {
	  log.Panicf("Error starting chaincode: %v", err)
	}
  }
  