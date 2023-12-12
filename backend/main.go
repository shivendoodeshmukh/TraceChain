package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	walletPath                       = "./wallet"
	PharmaCryptoPath                 = "../fablo-target/fabric-config/crypto-config/peerOrganizations/pharma.example.com"
	HospitalCryptoPath               = "../fablo-target/fabric-config/crypto-config/peerOrganizations/hospital.example.com"
	TrackerCryptoPath                = "../fablo-target/fabric-config/crypto-config/peerOrganizations/tracker.example.com"
	pharmaPeer                       = "localhost:7041"
	hospitalPeer                     = "localhost:7061"
	trackerPeer                      = "localhost:7081"
	pharmaGateway                    = "peer0.pharma.example.com"
	hospitalGateway                  = "peer0.hospital.example.com"
	trackerGateway                   = "peer0.tracker.example.com"
	pharmaCA                         = "localhost:7040"
	hospitalCA                       = "localhost:7060"
	trackerCA                        = "localhost:7080"
	channelName                      = "supply"
	DeviceStoreChaincodeName         = "DeviceStore"
	LogContractChaincodeName         = "LogContract"
	GenerateCertificateChaincodeName = "GenerateCertificate"
)

type Device struct {
	DeviceID string `json:"deviceID"`
	Password string `json:"password"`
	Role     string `json:"role"`
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
	Capabilities string `json:"capabilities"`
}

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

type PageData struct {
	Certificate Certificate
	Valid       bool
	Waypoints   []waypoint
}

type waypoint struct {
	Lat string
	Lon string
	Log Log
}

func main() {
	PharmaConnection := newGrpcConnection(pharmaGateway, pharmaPeer, PharmaCryptoPath+"/peers/peer0.pharma.example.com/tls/ca.crt")
	HospitalConnection := newGrpcConnection(hospitalGateway, hospitalPeer, HospitalCryptoPath+"/peers/peer0.hospital.example.com/tls/ca.crt")
	TrackerConnection := newGrpcConnection(trackerGateway, trackerPeer, TrackerCryptoPath+"/peers/peer0.tracker.example.com/tls/ca.crt")
	defer PharmaConnection.Close()
	defer HospitalConnection.Close()
	defer TrackerConnection.Close()

	pharmaIdentity := newIdentity(PharmaCryptoPath+"/users/User1@pharma.example.com/msp/signcerts/User1@pharma.example.com-cert.pem", "PharmaMSP")
	hospitalIdentity := newIdentity(HospitalCryptoPath+"/users/User1@hospital.example.com/msp/signcerts/User1@hospital.example.com-cert.pem", "HospitalMSP")
	trackerIdentity := newIdentity(TrackerCryptoPath+"/users/User1@tracker.example.com/msp/signcerts/User1@tracker.example.com-cert.pem", "IoTTrackerMSP")

	pharmaSign := newSign(PharmaCryptoPath + "/users/User1@pharma.example.com/msp/keystore/")
	hospitalSign := newSign(HospitalCryptoPath + "/users/User1@hospital.example.com/msp/keystore/")
	trackerSign := newSign(TrackerCryptoPath + "/users/User1@tracker.example.com/msp/keystore")

	gwpharma, err := client.Connect(
		pharmaIdentity, client.WithSign(pharmaSign),
		client.WithClientConnection(PharmaConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gwpharma.Close()

	gwhospital, err := client.Connect(
		hospitalIdentity, client.WithSign(hospitalSign),
		client.WithClientConnection(HospitalConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gwhospital.Close()

	gwtracker, err := client.Connect(
		trackerIdentity, client.WithSign(trackerSign),
		client.WithClientConnection(TrackerConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gwtracker.Close()

	pharmanet := gwpharma.GetNetwork(channelName)
	hospitalnet := gwhospital.GetNetwork(channelName)
	trackernet := gwtracker.GetNetwork(channelName)
	// Test
	Enroll(pharmanet.GetContract(DeviceStoreChaincodeName), "Admin-002", "adminpw", "admin")
	AuthHelper(pharmanet.GetContract(DeviceStoreChaincodeName), "Admin-002", "adminpw")

	// fmt.Println(string1, err)

	r := mux.NewRouter()
	r.HandleFunc("/api/enroll", func(w http.ResponseWriter, r *http.Request) {
		// Verify Cookie username and password resolves to admin
		// If not, return 401
		fmt.Println(r.Header)
		fmt.Println(r.PostForm)

		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("No token provided"))
			return
		}
		role, err := AuthHelper(pharmanet.GetContract(DeviceStoreChaincodeName), "Admin-002", token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}
		if role != "admin" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not an admin"))
			return
		}

		// Extract deviceID, password, role from request body
		deviceID := r.PostFormValue("deviceID")
		password := r.PostFormValue("password")
		role = r.PostFormValue("role")

		fmt.Println(deviceID, password, role)
		// Enroll the device
		err = Enroll(pharmanet.GetContract(DeviceStoreChaincodeName), deviceID, password, role)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Device enrolled successfully as " + role + " with ID " + deviceID))
	})
	r.HandleFunc("/api/init", func(w http.ResponseWriter, r *http.Request) {
		// Verify Cookie username and password resolves to tracker

		username := r.PostFormValue("deviceID")
		password := r.PostFormValue("password")
		role, err := AuthHelper(trackernet.GetContract(DeviceStoreChaincodeName), username, password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}
		if role != "tracker" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not a tracker"))
			return
		}

		// Extract deviceID, capabilities from request body
		deviceID := r.PostFormValue("deviceID")
		capabilities := r.PostFormValue("capabilities")

		fmt.Println(deviceID, capabilities)
		// Initialize the log
		err = InitLog(pharmanet.GetContract(LogContractChaincodeName), deviceID, capabilities)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Log initialized successfully"))
	})
	r.HandleFunc("/api/append", func(w http.ResponseWriter, r *http.Request) {
		// Verify Cookie username and password resolves to tracker

		username := r.PostFormValue("deviceID")
		password := r.PostFormValue("password")
		role, err := AuthHelper(trackernet.GetContract(DeviceStoreChaincodeName), username, password)
		fmt.Println(username, password, role)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			fmt.Println("Error authenticating")
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}
		if role != "tracker" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not a tracker"))
			return
		}

		var logs []Log
		err = json.Unmarshal([]byte(r.PostFormValue("logs")), &logs)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			fmt.Println(err.Error())
			fmt.Println("Error decoding logs")
			return
		}

		for _, log := range logs {
			err = LogAppend(pharmanet.GetContract(LogContractChaincodeName), log)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				actualErr, _ := status.FromError(err)
				detailsString := fmt.Sprintf("%v", actualErr.Details())
				w.Write([]byte(detailsString))
				fmt.Println(detailsString)
				return
			}
		}

	})
	r.HandleFunc("/api/generateCertificate", func(w http.ResponseWriter, r *http.Request) {
		// Verify username and password resolves to hospital

		username := r.PostFormValue("deviceID")
		password := r.PostFormValue("password")
		role, err := AuthHelper(hospitalnet.GetContract(DeviceStoreChaincodeName), username, password)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}

		if role != "hospital" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not a hospital"))
			return
		}

		var cert Certificate
		cert.DeviceID = r.PostFormValue("targetDeviceID")
		cert.MaxIntTemp = r.PostFormValue("maxIntTemp")
		cert.MinIntTemp = r.PostFormValue("minIntTemp")
		cert.MaxExtTemp = r.PostFormValue("maxExtTemp")
		cert.MinExtTemp = r.PostFormValue("minExtTemp")
		cert.MaxHum = r.PostFormValue("maxHum")
		cert.MaxXAccl = r.PostFormValue("maxXAccl")
		cert.MaxYAccl = r.PostFormValue("maxYAccl")
		cert.MaxZAccl = r.PostFormValue("maxZAccl")
		cert.MaxPitch = r.PostFormValue("maxPitch")
		cert.MaxRoll = r.PostFormValue("maxRoll")
		cert.MaxYaw = r.PostFormValue("maxYaw")
		cert.MaxAlt = r.PostFormValue("maxAlt")

		fmt.Println(cert)
		// Create certificate
		hash, err := CertificateCreator(pharmanet.GetContract(GenerateCertificateChaincodeName), cert)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(hash))
	})
	r.HandleFunc("/api/getCertificate", func(w http.ResponseWriter, r *http.Request) {
		// Verify username and password resolves to hospital

		username := r.PostFormValue("deviceID")
		password := r.PostFormValue("password")
		role, err := AuthHelper(hospitalnet.GetContract(DeviceStoreChaincodeName), username, password)

		fmt.Println(username, password, role)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}

		if role != "hospital" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not a hospital or admin"))
			return
		}

		certificateHash := r.PostFormValue("certificateHash")

		valid, err := hospitalnet.GetContract(GenerateCertificateChaincodeName).EvaluateTransaction("ValidateCertificate", certificateHash)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			actualErr, _ := status.FromError(err)
			detailsString := fmt.Sprintf("%v", actualErr.Details())
			w.Write([]byte(detailsString))
			fmt.Println(detailsString)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valid))
	})
	// Depending on POST or GET, call different functions
	r.HandleFunc("/CertificateForm", func(w http.ResponseWriter, r *http.Request) {
		// Render CertificateForm.html template on GET
		if r.Method == "GET" {
			http.ServeFile(w, r, "templates/CertificateForm.html")
		}
	})
	r.HandleFunc("/VerifyCertificateForm", func(w http.ResponseWriter, r *http.Request) {
		// Render VerifyCertificateForm.html template on GET
		if r.Method == "GET" {
			http.ServeFile(w, r, "templates/VerifyCertificateForm.html")
		}
		if r.Method == "POST" {
			// Pass request to /api/getCertificate and also render Certificate.html template with certificate and map waypoints
			fmt.Println(r.PostFormValue("certificateHash"))
			certificateHash := r.PostFormValue("certificateHash")
			certificate, err := CertificateFetch(hospitalnet.GetContract(GenerateCertificateChaincodeName), certificateHash)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				fmt.Println(err.Error())
				fmt.Println("Error fetching certificate")
				return
			}
			fmt.Println(certificate)

			logs, err := LogFetch(hospitalnet.GetContract(LogContractChaincodeName), certificate.DeviceID)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				fmt.Println(err.Error())
				fmt.Println("Error fetching logs")
				return
			}

			// Get valid from /api/getCertificate endpoint
			// Call /api/getCertificate endpoint
			fmt.Println("Calling /api/getCertificate endpoint")
			fmt.Println(r.PostFormValue("deviceID"))
			fmt.Println(r.PostFormValue("password"))
			form := url.Values{}
			form.Add("deviceID", r.PostFormValue("deviceID"))
			form.Add("password", r.PostFormValue("password"))
			form.Add("certificateHash", certificateHash)
			resp, err := http.PostForm("http://localhost:8000/api/getCertificate", form)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				fmt.Println(err.Error())
				fmt.Println("Error creating request")
				return
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(body))

			valid, err := strconv.ParseBool(string(body))
			if err != nil {
				valid = false
			}
			fmt.Println(valid)
			waypoints := make([]waypoint, len(logs)-1)
			for i, log := range logs {
				if i == len(logs)-1 {
					continue
				}
				waypoints[i].Lat = log.Lat
				waypoints[i].Lon = log.Lon
				waypoints[i].Log = logs[i]
			}

			data := PageData{
				Certificate: certificate,
				Waypoints:   waypoints,
				Valid:       valid,
			}
			fmt.Println(data)

			tmpl := template.Must(template.ParseFiles("templates/Certificate.html"))
			tmpl.Execute(w, data)
		}
	})

	fmt.Println("Server started at port 8000")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", handlers.CORS()(r)))

}

func newGrpcConnection(gatewayPeer string, peerEndpoint string, tlspath string) *grpc.ClientConn {
	certificate, err := loadCertificate(tlspath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

func newIdentity(certPath string, MSPID string) *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(MSPID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func newSign(keyPath string) identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func AuthHelper(contract *client.Contract, deviceID string, password string) (string, error) {
	// Authenticate the device
	result, err := contract.EvaluateTransaction("AuthenticateDevice", deviceID, password)
	e, ok := status.FromError(err)

	if ok {
		if e.Message() == "evaluate call to endorser returned error: chaincode response 500, device does not exist" {
			return "", fmt.Errorf("device not authenticated")
		}
	}

	fmt.Printf("Device authenticated with role: %s\n", result)

	return string(result), nil
}

func Enroll(contract *client.Contract, deviceID string, password string, role string) error {
	// Enroll the device
	_, err := contract.SubmitTransaction("RegisterDevice", deviceID, password, role)
	if err != nil {
		return err
	}

	fmt.Println("Device enrolled successfully")

	return nil
}

func InitLog(contract *client.Contract, deviceID string, capabilities string) error {
	// Initialize the log
	_, err := contract.SubmitTransaction("Init", deviceID, capabilities)
	if err != nil {
		return err
	}

	fmt.Println("Log initialized successfully")

	return nil
}

func LogAppend(contract *client.Contract, log Log) error {
	// Append to the log
	fmt.Println(log)
	_, err := contract.SubmitTransaction("AppendLog", log.DeviceID, log.Lat, log.Lon, log.ExtTemp, log.IntTemp, log.Hum, log.MaxXAccl, log.MaxYAccl, log.MaxZAccl, log.Pitch, log.Roll, log.Yaw, log.Alt, log.Satellites, log.Timestamp)
	if err != nil {
		return err
	}

	fmt.Println("Log appended successfully")

	return nil
}

func CertificateCreator(contract *client.Contract, cert Certificate) (string, error) {
	// Create certificate
	hash, err := contract.SubmitTransaction("GenerateCertificate", cert.DeviceID, cert.MaxIntTemp, cert.MinIntTemp, cert.MaxExtTemp, cert.MinExtTemp, cert.MaxHum, cert.MaxXAccl, cert.MaxYAccl, cert.MaxZAccl, cert.MaxPitch, cert.MaxRoll, cert.MaxYaw, cert.MaxAlt)
	if err != nil {
		return "", err
	}

	fmt.Println("Certificate created successfully")
	fmt.Println(string(hash))

	return string(hash), nil
}

func LogFetch(contract *client.Contract, deviceID string) ([]Log, error) {
	// Fetch the log
	resp, err := contract.EvaluateTransaction("getAllLogsByDeviceID", deviceID)
	if err != nil {
		return nil, err
	}

	logstreamraw := resp
	var logstream []string
	err = json.Unmarshal(logstreamraw, &logstream)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal logstream: %s", err.Error())
	}
	var logs []Log

	for i := 0; i < len(logstream); i++ {
		var log Log
		err = json.Unmarshal([]byte(logstream[i]), &log)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal log: %s %s", err.Error(), logstream[i])
		}
		logs = append(logs, log)
	}

	fmt.Println("Log fetched successfully")

	return logs, nil
}

func CertificateFetch(contract *client.Contract, certificateHash string) (Certificate, error) {
	// Fetch the certificate
	fmt.Println(certificateHash)
	resp, err := contract.EvaluateTransaction("GetCertificateByHash", certificateHash)
	if err != nil {
		return Certificate{}, err
	}
	fmt.Println(resp)

	fmt.Println("Certificate fetched successfully")

	var certificate Certificate
	err = json.Unmarshal([]byte(resp), &certificate)
	if err != nil {
		fmt.Print(resp)
		return Certificate{}, fmt.Errorf("failed to unmarshal certificate: %s", err.Error())
	}

	return certificate, nil
}
