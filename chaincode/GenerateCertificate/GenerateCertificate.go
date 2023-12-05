package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract generates certificate for constraints on IoT Tracker, and verifies the constraints are met, validating the certificate
type GenerateCertificate struct {
	contractapi.Contract
}

type Log struct {
	DeviceID     string `json:"deviceID"`
	Lat          string `json:"lat"`
	Lon          string `json:"lon"`
	IntTemp      string `json:"intTemp"`
	ExtTemp      string `json:"extTemp"`
	Hum          string `json:"hum"`
	MaxXAccl     string `json:"maxXAccl"`
	MaxYAccl     string `json:"maxYAccl"`
	MaxZAccl     string `json:"maxZAccl"`
	Pitch        string `json:"pitch"`
	Roll         string `json:"roll"`
	Yaw          string `json:"yaw"`
	Alt          string `json:"alt"`
	Satellites   string `json:"satellites"`
	Timestamp    string `json:"timestamp"`
	PrevHash     string `json:"prevHash"`
	Capabilities string `json:"capabilities"`
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
func (r *GenerateCertificate) GenerateCertificate(ctx contractapi.TransactionContextInterface, deviceID string, maxIntTemp string, minIntTemp string, maxExtTemp string, minExtTemp string, maxHum string, maxXAccl string, maxYAccl string, maxZAccl string, maxPitch string, maxRoll string, maxYaw string, maxAlt string) (string, error) {
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
		return "", err
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(certificateJSON))

	err = ctx.GetStub().PutState(hash, certificateJSON)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// ValidateCertificate validates a certificate against a logstream
func (r *GenerateCertificate) ValidateCertificate(ctx contractapi.TransactionContextInterface, certificateHash string, logstreamHash string) (bool, error) {
	certificateBytes, err := ctx.GetStub().GetState(certificateHash)
	if err != nil {
		return false, err
	}
	if certificateBytes == nil {
		return false, nil
	}

	var certificate Certificate
	err = json.Unmarshal([]byte(certificateBytes), &certificate)
	if err != nil {
		return false, err
	}

	chainCodeArgs := [][]byte{[]byte("getAllLogsByDeviceID"), []byte(logstreamHash)}
	// Get logstream by calling LogContract chaincode
	resp := ctx.GetStub().InvokeChaincode("LogContract", chainCodeArgs, "supply")

	if resp.Status != 200 {
		return false, fmt.Errorf("failed to get logstream: %s", resp.Message)
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
			return false, fmt.Errorf("logstream invalidates certificate (prevHash mismatch)")
		}
	}

	// Check if logstream meets certificate constraints
	for i := 0; i < len(logs); i++ {
		if logs[i].IntTemp > certificate.MaxIntTemp || logs[i].IntTemp < certificate.MinIntTemp {
			return false, fmt.Errorf("logstream invalidates certificate (intTemp constraint not met)")
		}
		if logs[i].ExtTemp > certificate.MaxExtTemp || logs[i].ExtTemp < certificate.MinExtTemp {
			return false, fmt.Errorf("logstream invalidates certificate (extTemp constraint not met)")
		}
		if logs[i].Hum > certificate.MaxHum {
			return false, fmt.Errorf("logstream invalidates certificate (hum constraint not met)")
		}
		if logs[i].MaxXAccl > certificate.MaxXAccl {
			return false, fmt.Errorf("logstream invalidates certificate (maxXAccl constraint not met)")
		}
		if logs[i].MaxYAccl > certificate.MaxYAccl {
			return false, fmt.Errorf("logstream invalidates certificate (maxYAccl constraint not met)")
		}
		if logs[i].MaxZAccl > certificate.MaxZAccl {
			return false, fmt.Errorf("logstream invalidates certificate (maxZAccl constraint not met)")
		}
		if logs[i].Pitch > certificate.MaxPitch {
			return false, fmt.Errorf("logstream invalidates certificate (pitch constraint not met)")
		}
		if logs[i].Roll > certificate.MaxRoll {
			return false, fmt.Errorf("logstream invalidates certificate (roll constraint not met)")
		}
		if logs[i].Yaw > certificate.MaxYaw {
			return false, fmt.Errorf("logstream invalidates certificate (yaw constraint not met)")
		}
		if logs[i].Alt > certificate.MaxAlt {
			return false, fmt.Errorf("logstream invalidates certificate (alt constraint not met)")
		}
	}

	return true, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(GenerateCertificate))
	if err != nil {
		panic(err.Error())
	}

	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
