package main

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract generates certificate for constraints on IoT Tracker, and verifies the constraints are met, validating the certificate
type GenerateCertificate struct {
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

// Certificate represents a certificate
type Certificate struct {
	DeviceID    string `json:"deviceID"`
	MaxIntTemp  string `json:"maxIntTemp"`
	MinIntTemp  string `json:"minIntTemp"`
	MaxExtTemp  string `json:"maxExtTemp"`
	MinExtTemp  string `json:"minExtTemp"`
	MaxHum      string `json:"maxHum"`
	MaxXAccl    string `json:"maxXAccl"`
	MaxYAccl    string `json:"maxYAccl"`
	MaxZAccl    string `json:"maxZAccl"`
	MaxPitch    string `json:"maxPitch"`
	MaxRoll     string `json:"maxRoll"`
	MaxYaw      string `json:"maxYaw"`
	MaxAlt      string `json:"maxAlt"`
	CreatedAt   string `json:"createdAt"`
	LastUpdated string `json:"lastUpdated"`
}

// Init initializes the certificate
func (r *GenerateCertificate) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// GenerateCertificate generates a certificate, returning the certificate hash
func (r *GenerateCertificate) GenerateCertificate(ctx contractapi.TransactionContextInterface, deviceID string, maxIntTemp string, minIntTemp string, maxExtTemp string, minExtTemp string, maxHum string, maxXAccl string, maxYAccl string, maxZAccl string, maxPitch string, maxRoll string, maxYaw string, maxAlt string) ([32]byte, error) {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

	certificate := Certificate{
		DeviceID:    deviceID,
		MaxIntTemp:  maxIntTemp,
		MinIntTemp:  minIntTemp,
		MaxExtTemp:  maxExtTemp,
		MinExtTemp:  minExtTemp,
		MaxHum:      maxHum,
		MaxXAccl:    maxXAccl,
		MaxYAccl:    maxYAccl,
		MaxZAccl:    maxZAccl,
		MaxPitch:    maxPitch,
		MaxRoll:     maxRoll,
		MaxYaw:      maxYaw,
		MaxAlt:      maxAlt,
		CreatedAt:   timestamp,
		LastUpdated: timestamp,
	}

	certificateJSON, err := json.Marshal(certificate)
	if err != nil {
		return [32]byte{}, err
	}

	hash := sha256.Sum256(certificateJSON)

	err = ctx.GetStub().PutState(hash, certificateJSON)
	if err != nil {
		return [32]byte{}, err
	}

	return hash, nil
}

// ValidateCertificate validates a certificate against a logstream
func (r *GenerateCertificate) ValidateCertificate(ctx contractapi.TransactionContextInterface, certificateHash [32]byte, logstreamHash [32]byte) (bool, error) {
	certificateBytes, err := ctx.GetStub().GetState(certificateHash)
	if err != nil {
		return false, err
	}
	if certificateBytes == nil {
		return false, nil
	}

	chainCodeArgs := ctx.util.ToChaincodeArgs("getAllLogsByDeviceID", certificateBytes.(Certificate).DeviceID)
	// Get logstream by calling LogContract chaincode
	resp := ctx.GetStub().InvokeChaincode("LogContract", chainCodeArgs, "supply")

	if resp.Status != ctx.shim.OK {
		return false, ctx.shim.error(resp.Message)
	}

	logstream := resp.Payload
	var logs []Log
	err = json.Unmarshal([]byte(logstream), &logs)
	if err != nil {
		return false, err
	}

	// Check if logstream is valid
	for i := 0; i < len(logs); i++ {
		if logs[i].PrevHash != logs[i-1].PrevHash {
			return false, err("Logstream is invalid (prevHash mismatch)")
		}
	}

	// Check if logstream meets certificate constraints
	for i := 0; i < len(logs); i++ {
		if logs[i].IntTemp > certificateBytes.(Certificate).MaxIntTemp || logs[i].IntTemp < certificateBytes.(Certificate).MinIntTemp {
			return false, err("Log invalidates certificate (intTemp out of bounds)")
		}
		if logs[i].ExtTemp > certificateBytes.(Certificate).MaxExtTemp || logs[i].ExtTemp < certificateBytes.(Certificate).MinExtTemp {
			return false, err("Log invalidates certificate (extTemp out of bounds)")
		}
		if logs[i].Hum > certificateBytes.(Certificate).MaxHum {
			return false, err("Log invalidates certificate (hum out of bounds)")
		}
		if logs[i].MaxXAccl > certificateBytes.(Certificate).MaxXAccl {
			return false, err("Log invalidates certificate (maxXAccl out of bounds)")
		}
		if logs[i].MaxYAccl > certificateBytes.(Certificate).MaxYAccl {
			return false, err("Log invalidates certificate (maxYAccl out of bounds)")
		}
		if logs[i].MaxZAccl > certificateBytes.(Certificate).MaxZAccl {
			return false, err("Log invalidates certificate (maxZAccl out of bounds)")
		}
		if logs[i].Pitch > certificateBytes.(Certificate).MaxPitch {
			return false, err("Log invalidates certificate (pitch out of bounds)")
		}
		if logs[i].Roll > certificateBytes.(Certificate).MaxRoll {
			return false, err("Log invalidates certificate (roll out of bounds)")
		}
		if logs[i].Yaw > certificateBytes.(Certificate).MaxYaw {
			return false, err("Log invalidates certificate (yaw out of bounds)")
		}
		if logs[i].Alt > certificateBytes.(Certificate).MaxAlt {
			return false, err("Log invalidates certificate (alt out of bounds)")
		}
	}

	return true, nil
}
