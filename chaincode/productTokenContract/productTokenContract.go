package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

// ProductContract defines the basic structure for the smart contract
type ProductContract struct {
	contractapi.Contract
}

type Product struct {
	ID          	string `json:"id"`
	Name        	string `json:"name"`
	Status      	string `json:"status"`
	Distributor    	string `json:"distributor"`
	Manufacturer 	string `json:"manufacturer"`
	RawMaterialIDs 	[]string `json:"rawMaterialIDs"`
}

type RawMaterial struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Supplier    string `json:"supplier"`
	Manufacturer string `json:"manufacturer"`
}

// Init initializes the smart contract
func (r *ProductContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Product smart contract initialized")
	return nil
}

// CreateRawMaterial creates a new product token by burning the raw material tokens
func (r *ProductContract) CreateProduct(ctx contractapi.TransactionContextInterface, name string, rawMaterialIDs []string) (string, error) {

	// Use a incrementing ID to ensure unique product IDs
	productIDbyte, err := ctx.GetStub().GetState("ProductIDIterator")
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	productID := string(productIDbyte)
	if productIDbyte == nil {
		productID = "0"
	}
	
	productID = "Product-" + productID
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "ManufacturerMSP" {
		return "", fmt.Errorf("Client of org %s is not authorized to create a token", clientOrgID)
	}

	existing, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if existing != nil {
		return "", fmt.Errorf("product with ID %s already exists", productID)
	}

	ManufacturerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	var rawMaterialIDsToBurn []string
	for _, rawMaterialID := range rawMaterialIDs {
		rawMaterialID = "RawMaterial-" + rawMaterialID
		rawMaterialJSON, err := ctx.GetStub().GetState(rawMaterialID)
		if err != nil {
			return "", fmt.Errorf("failed to read from world state: %v", err)
		}
		if rawMaterialJSON == nil {
			return "", fmt.Errorf("raw material with ID %s does not exist", rawMaterialID)
		}

		var rawMaterial RawMaterial
		err = json.Unmarshal(rawMaterialJSON, &rawMaterial)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal raw material JSON: %v", err)
		}

		if rawMaterial.Status != "Transferred" {
			return "", fmt.Errorf("raw material with ID %s cannot be burned. Status: %s", rawMaterialID, rawMaterial.Status)
		}

		if rawMaterial.Manufacturer != ManufacturerID {
			return "", fmt.Errorf("raw material with ID %s cannot be burned. Manufacturer: %s", rawMaterialID, rawMaterial.Manufacturer)
		}

		rawMaterialIDsToBurn = append(rawMaterialIDsToBurn, rawMaterialID)
	}

	product := Product{
		ID:          	productID,
		Name:        	name,
		Status:      	"New",
		Manufacturer: 	ManufacturerID,
		RawMaterialIDs: rawMaterialIDsToBurn,
	}

	productJSON, err := json.Marshal(product)
	if err != nil {
		return "", fmt.Errorf("failed to marshal product JSON: %v", err)
	}

	err = ctx.GetStub().PutState(productID, productJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write to world state: %v", err)
	}

	for _, rawMaterialID := range rawMaterialIDsToBurn {
		err = ctx.GetStub().DelState(rawMaterialID)
		if err != nil {
			return "", fmt.Errorf("failed to delete from world state: %v", err)			
		}
	}

	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		return "", fmt.Errorf("failed to convert productID to int: %v", err)
	}
	productIDInt++
	productIDnext := strconv.Itoa(productIDInt)
	err = ctx.GetStub().PutState("ProductIDIterator", []byte(productIDnext))
	if err != nil {
		return "", fmt.Errorf("failed to write to world state: %v", err)
	}


	return "", nil
}

// InitiateTransferProduct transfers a product token to a distributor
func (r *ProductContract) InitiateTransferProduct(ctx contractapi.TransactionContextInterface, productID string, distributorID string) error {
	productID = "Product-" + productID
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "ManufacturerMSP" {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}

	productJSON, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return fmt.Errorf("product with ID %s does not exist", productID)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}
	ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}

	if product.Manufacturer != ID {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}
	if product.Status != "New" {
		return fmt.Errorf("product with ID %s cannot be transferred. Status: %s", productID, product.Status)
	}

	product.Status = "Transferring"
	product.Distributor = distributorID
	
	productJSON, err = json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal updated product JSON: %v", err)
	}

	err = ctx.GetStub().PutState(productID, productJSON)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

// CompleteTransferProduct completes transfer of the product token to the distributor
func (r *ProductContract) CompleteTransferProduct(ctx contractapi.TransactionContextInterface, productID string) error {
	productID = "Product-" + productID
	productJSON, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return fmt.Errorf("product with ID %s does not exist", productID)
	}

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}
	if clientOrgID != "DistributorMSP" {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	DistributorID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed getting client's orgID: %v", err)
	}

	if DistributorID != product.Distributor {
		return fmt.Errorf("Client of org %s is not authorized to perform this action", clientOrgID)
	}
	
	if product.Status != "Transferring" {
		return fmt.Errorf("product with ID %s cannot be transferred. Status: %s", productID, product.Status)
	}

	product.Status = "Transferred"

	productJSON, err = json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal updated product JSON: %v", err)
	}

	err = ctx.GetStub().PutState(productID, productJSON)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

// GetProduct returns the product token stored in the world state with given id
func (r *ProductContract) GetProduct(ctx contractapi.TransactionContextInterface, productID string) (*Product, error) {
	productID = "Product-" + productID
	productJSON, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("product with ID %s does not exist", productID)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	return &product, nil
}

// GetProductHistory returns the history of the product token with given id
func (r *ProductContract) GetProductHistory(ctx contractapi.TransactionContextInterface, productID string) ([]*Product, error) {
	productID = "Product-" + productID
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}
	
		var product Product
		err = json.Unmarshal(queryResponse.Value, &product)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *ProductContract) GetTotalNumberOfProducts(ctx contractapi.TransactionContextInterface) (int, error) {
	productIDbyte, err := ctx.GetStub().GetState("ProductIDIterator")
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	productID := string(productIDbyte)
	if productIDbyte == nil {
		productID = "0"
	}

	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		return 0, fmt.Errorf("failed to convert productID to int: %v", err)
	}
	return productIDInt, nil
}

func (r *ProductContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	productIDbyte, err := ctx.GetStub().GetState("ProductIDIterator")
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	productID := string(productIDbyte)
	if productIDbyte == nil {
		productID = "0"
	}

	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert productID to int: %v", err)
	}

	var products []*Product
	for i := 1; i <= productIDInt; i++ {
		productID = strconv.Itoa(i)
		productJSON, err := r.GetProduct(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}
		ID, err := ctx.GetClientIdentity().GetMSPID()
		if err != nil {
			return nil, fmt.Errorf("failed getting client's orgID: %v", err)
		}
		MSPID, err := ctx.GetClientIdentity().GetMSPID()
		if err != nil {
			return nil, fmt.Errorf("failed getting client's orgID: %v", err)
		}

		if MSPID == "ManufacturerMSP" {
			if productJSON.Manufacturer == ID {
				products = append(products, productJSON)
			}
		} else if MSPID == "DistributorMSP" {
			if productJSON.Distributor == ID {
				products = append(products, productJSON)
			}
		} else {
			return nil, fmt.Errorf("Client of org %s is not authorized to perform this action", MSPID)
		}
	}

	return products, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&ProductContract{})
	if err != nil {
	  log.Panicf("Error creating chaincode: %v", err)
	}
  
	if err := assetChaincode.Start(); err != nil {
	  log.Panicf("Error starting chaincode: %v", err)
	}
  }
  