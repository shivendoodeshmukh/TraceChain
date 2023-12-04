package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides logging functions for IoT Tracker, which reports the location of a device, temperature, humidity, max acceleration in x, y, z directions, pitch, roll, yaw, and timestamp
type LogContract struct {
	contractapi.Contract
}

type Log struct {
	DeviceID     string   `json:"deviceID"`
	Lat          string   `json:"lat"`
	Lon          string   `json:"lon"`
	IntTemp      string   `json:"intTemp"`
	ExtTemp      string   `json:"extTemp"`
	Hum          string   `json:"hum"`
	MaxXAccl     string   `json:"maxXAccl"`
	MaxYAccl     string   `json:"maxYAccl"`
	MaxZAccl     string   `json:"maxZAccl"`
	Pitch        string   `json:"pitch"`
	Roll         string   `json:"roll"`
	Yaw          string   `json:"yaw"`
	Alt          string   `json:"alt"`
	Satellites   string   `json:"satellites"`
	Timestamp    string   `json:"timestamp"`
	PrevHash     [32]byte `json:"prevHash"`
	Capabilities string   `json:"capabilities"`
}

func (r *LogContract) Init(ctx contractapi.TransactionContextInterface, deviceID string, capabilities string) error {
	fmt.Println("init")

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	log := Log{
		DeviceID:     deviceID,
		Capabilities: capabilities,
		Timestamp:    timestamp,
	}

	logJSON, err := json.Marshal(log)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(logJSON)

	err = ctx.GetStub().PutState(hash, logJSON)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(deviceID, hash)
	if err != nil {
		return err
	}

	return nil
}

func (r *LogContract) log(ctx contractapi.TransactionContextInterface, deviceID string, lat string, lon string, extTemp string, intTemp string, hum string, maxXAccl string, maxYAccl string, maxZAccl string, pitch string, roll string, yaw string, timestamp string) error {
	fmt.Println("log")

	prevHash, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return err
	}

	log := Log{
		DeviceID:  deviceID,
		Lat:       lat,
		Lon:       lon,
		ExtTemp:   extTemp,
		IntTemp:   intTemp,
		Hum:       hum,
		MaxXAccl:  maxXAccl,
		MaxYAccl:  maxYAccl,
		MaxZAccl:  maxZAccl,
		Pitch:     pitch,
		Roll:      roll,
		Yaw:       yaw,
		Timestamp: timestamp,
		PrevHash:  prevHash,
	}

	logJSON, err := json.Marshal(log)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(logJSON)

	err = ctx.GetStub().PutState(deviceID, hash)
	if err != nil {
		return err
	}

	// Append hash to hashes
	hashes, err := ctx.GetStub().GetState("hashes")
	if err != nil {
		return err
	}

	var hashesArray [][32]byte
	if hashes != nil {
		err = json.Unmarshal(hashes, &hashesArray)
		if err != nil {
			return err
		}
	}

	hashesArray = append(hashesArray, hash)

	hashesJSON, err := json.Marshal(hashesArray)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState("hashes", hashesJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(hash, logJSON)
}

func (r *LogContract) getLog(ctx contractapi.TransactionContextInterface, hash string) (string, error) {
	fmt.Println("getLog")

	logJSON, err := ctx.GetStub().GetState(hash)
	if err != nil {
		return "", err
	}

	return string(logJSON), nil
}

func (r *LogContract) getLogByDeviceID(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	fmt.Println("getLogByDeviceID")

	hash, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return "", err
	}

	logJSON, err := ctx.GetStub().GetState(string(hash))
	if err != nil {
		return "", err
	}

	return string(logJSON), nil
}

func (r *LogContract) getAllLogsByDeviceID(ctx contractapi.TransactionContextInterface, deviceID string) ([]string, error) {
	fmt.Println("getAllLogsByDeviceID")

	hash, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return nil, err
	}

	logJSON, err := ctx.GetStub().GetState(hash)
	if err != nil {
		return nil, err
	}

	var logs []string
	logs = append(logs, string(logJSON))

	for {
		log := Log{}
		err = json.Unmarshal(logJSON, &log)
		if err != nil {
			return nil, err
		}

		prevHash := log.PrevHash

		if prevHash == [32]byte{} {
			break
		}

		logJSON, err = ctx.GetStub().GetState(prevHash)
		if err != nil {
			return nil, err
		}

		logs = append(logs, string(logJSON))
	}

	return logs, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(LogContract))
	if err != nil {
		fmt.Printf("Error creating log contract: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting log contract: %s", err.Error())
	}
}
