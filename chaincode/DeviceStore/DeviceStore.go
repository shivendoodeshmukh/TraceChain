package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"golang.org/x/crypto/bcrypt"
)

// SmartContract provides authentication for IoT Tracker
type DeviceStore struct {
	contractapi.Contract
}

// Device represents a device
type Device struct {
	DeviceID     string `json:"deviceID"`
	PasswordHash string `json:"password"`
}

// Init initializes the device store
func (r *DeviceStore) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// RegisterDevice registers a device (avoid overwriting existing device)
func (r *DeviceStore) RegisterDevice(ctx contractapi.TransactionContextInterface, deviceID string, password string) error {

	// Salt and hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	device := Device{
		DeviceID:     deviceID,
		PasswordHash: string(hash),
	}

	// Check if device already exists
	deviceBytes, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return err
	}
	if deviceBytes != nil {
		return err("Device already exists")
	}

	err = ctx.GetStub().PutState(deviceID, device)
	if err != nil {
		return err
	}

	return nil
}

// AuthenticateDevice authenticates a device
func (r *DeviceStore) AuthenticateDevice(ctx contractapi.TransactionContextInterface, deviceID string, password string) (bool, error) {

	device, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return false, err
	}
	if device == nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(device.(Device).PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(DeviceStore))
	if err != nil {
		panic(err.Error())
	}

	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
