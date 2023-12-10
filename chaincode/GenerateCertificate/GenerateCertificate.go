package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
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
func (r *GenerateCertificate) ValidateCertificate(ctx contractapi.TransactionContextInterface, certificateHash string) (bool, error) {
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

	chainCodeArgs := [][]byte{[]byte("getAllLogsByDeviceID"), []byte(certificate.DeviceID)}
	// Get logstream by calling LogContract chaincode
	resp := ctx.GetStub().InvokeChaincode("LogContract", chainCodeArgs, "supply")

	if resp.Status != 200 {
		return false, fmt.Errorf("failed to get logstream: %s", resp.Message)
	}

	logstreamraw := resp.Payload
	var logstream []string
	err = json.Unmarshal(logstreamraw, &logstream)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal logstream: %s", err.Error())
	}
	var logs []Log

	for i := 0; i < len(logstream); i++ {
		var log Log
		err = json.Unmarshal([]byte(logstream[i]), &log)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal log: %s %s", err.Error(), logstream[i])
		}
		logs = append(logs, log)
	}

	// Check if logstream meets certificate constraints
	for i := 0; i < len(logs); i++ {
		intTemp, _ := strconv.ParseFloat(logs[i].IntTemp, 10)
		extTemp, _ := strconv.ParseFloat(logs[i].ExtTemp, 10)
		hum, _ := strconv.ParseFloat(logs[i].Hum, 10)
		maxXAccl, _ := strconv.ParseFloat(logs[i].MaxXAccl, 10)
		maxYAccl, _ := strconv.ParseFloat(logs[i].MaxYAccl, 10)
		maxZAccl, _ := strconv.ParseFloat(logs[i].MaxZAccl, 10)
		pitch, _ := strconv.ParseFloat(logs[i].Pitch, 10)
		roll, _ := strconv.ParseFloat(logs[i].Roll, 10)
		yaw, _ := strconv.ParseFloat(logs[i].Yaw, 10)
		alt, _ := strconv.ParseFloat(logs[i].Alt, 10)

		maxIntTemp, _ := strconv.ParseFloat(certificate.MaxIntTemp, 10)
		minIntTemp, _ := strconv.ParseFloat(certificate.MinIntTemp, 10)
		maxExtTemp, _ := strconv.ParseFloat(certificate.MaxExtTemp, 10)
		minExtTemp, _ := strconv.ParseFloat(certificate.MinExtTemp, 10)
		maxHum, _ := strconv.ParseFloat(certificate.MaxHum, 10)
		maxXAcclcert, _ := strconv.ParseFloat(certificate.MaxXAccl, 10)
		maxYAcclcert, _ := strconv.ParseFloat(certificate.MaxYAccl, 10)
		maxZAcclcert, _ := strconv.ParseFloat(certificate.MaxZAccl, 10)
		maxPitch, _ := strconv.ParseFloat(certificate.MaxPitch, 10)
		maxRoll, _ := strconv.ParseFloat(certificate.MaxRoll, 10)
		maxYaw, _ := strconv.ParseFloat(certificate.MaxYaw, 10)
		maxAlt, _ := strconv.ParseFloat(certificate.MaxAlt, 10)

		if (certificate.MaxIntTemp != "" && intTemp > maxIntTemp) || (certificate.MinIntTemp != "" && intTemp < minIntTemp) {
			return false, fmt.Errorf("intTemp constraint not met: %s not between %s and %s", logs[i].IntTemp, certificate.MinIntTemp, certificate.MaxIntTemp)
		}
		if (certificate.MaxExtTemp != "" && extTemp > maxExtTemp) || (certificate.MinExtTemp != "" && extTemp < minExtTemp) {
			return false, fmt.Errorf("extTemp constraint not met: %s not between %s and %s", logs[i].ExtTemp, certificate.MinExtTemp, certificate.MaxExtTemp)
		}
		if certificate.MaxHum != "" && hum > maxHum {
			return false, fmt.Errorf("hum constraint not met: %s not below %s", logs[i].Hum, certificate.MaxHum)
		}
		if certificate.MaxXAccl != "" && maxXAccl > maxXAcclcert {
			return false, fmt.Errorf("maxXAccl constraint not met: %s not below %s", logs[i].MaxXAccl, certificate.MaxXAccl)
		}
		if certificate.MaxYAccl != "" && maxYAccl > maxYAcclcert {
			return false, fmt.Errorf("maxYAccl constraint not met: %s not below %s", logs[i].MaxYAccl, certificate.MaxYAccl)
		}
		if certificate.MaxZAccl != "" && maxZAccl > maxZAcclcert {
			return false, fmt.Errorf("maxZAccl constraint not met: %s not below %s", logs[i].MaxZAccl, certificate.MaxZAccl)
		}
		if certificate.MaxPitch != "" && pitch > maxPitch {
			return false, fmt.Errorf("pitch constraint not met: %s not below %s", logs[i].Pitch, certificate.MaxPitch)
		}
		if certificate.MaxRoll != "" && roll > maxRoll {
			return false, fmt.Errorf("roll constraint not met: %s not below %s", logs[i].Roll, certificate.MaxRoll)
		}
		if certificate.MaxYaw != "" && yaw > maxYaw {
			return false, fmt.Errorf("yaw constraint not met: %s not below %s", logs[i].Yaw, certificate.MaxYaw)
		}
		if certificate.MaxAlt != "" && alt > maxAlt {
			return false, fmt.Errorf("alt constraint not met: %s not below %s", logs[i].Alt, certificate.MaxAlt)
		}
	}

	return true, nil
}

func (r *GenerateCertificate) GetCertificateByHash(ctx contractapi.TransactionContextInterface, certificateHash string) (string, error) {
	certificateBytes, err := ctx.GetStub().GetState(certificateHash)
	if err != nil {
		return "", err
	}
	if certificateBytes == nil {
		return "", fmt.Errorf("certificate %s does not exist", certificateHash)
	}

	return string(certificateBytes), nil
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
