#!/usr/bin/env bash

generateArtifacts() {
  printHeadline "Generating basic configs" "U1F913"

  printItalics "Generating crypto material for Orderer" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-orderer.yaml" "peerOrganizations/orderer.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Pharma" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-pharma.yaml" "peerOrganizations/pharma.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Hospital" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-hospital.yaml" "peerOrganizations/hospital.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for IoTTracker" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-iottracker.yaml" "peerOrganizations/tracker.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating genesis block for group group1" "U1F3E0"
  genesisBlockCreate "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config" "Group1Genesis"

  # Create directory for chaincode packages to avoid permission errors on linux
  mkdir -p "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"
}

startNetwork() {
  printHeadline "Starting network" "U1F680"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose up -d)
  sleep 4
}

generateChannelsArtifacts() {
  printHeadline "Generating config for 'supply'" "U1F913"
  createChannelTx "supply" "$FABLO_NETWORK_ROOT/fabric-config" "Supply" "$FABLO_NETWORK_ROOT/fabric-config/config"
}

installChannels() {
  printHeadline "Creating 'supply' on Pharma/peer0" "U1F63B"
  docker exec -i cli.pharma.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoin 'supply' 'PharmaMSP' 'peer0.pharma.example.com:7041' 'crypto/users/Admin@pharma.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'supply' on  Hospital/peer0" "U1F638"
  docker exec -i cli.hospital.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'supply' 'HospitalMSP' 'peer0.hospital.example.com:7061' 'crypto/users/Admin@hospital.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
  printItalics "Joining 'supply' on  IoTTracker/peer0" "U1F638"
  docker exec -i cli.tracker.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'supply' 'IoTTrackerMSP' 'peer0.tracker.example.com:7081' 'crypto/users/Admin@tracker.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
}

installChaincodes() {
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/LogContract")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'LogContract'" "U1F60E"
    chaincodeBuild "LogContract" "golang" "$CHAINCODES_BASE_DIR/./chaincode/LogContract" "16"
    chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" "golang" printHeadline "Installing 'LogContract' for Pharma" "U1F60E"
    chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" ""
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printHeadline "Installing 'LogContract' for Hospital" "U1F60E"
    chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "LogContract" "$version" ""
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printHeadline "Installing 'LogContract' for IoTTracker" "U1F60E"
    chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "LogContract" "$version" ""
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printItalics "Committing chaincode 'LogContract' on channel 'supply' as 'Pharma'" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""
  else
    echo "Warning! Skipping chaincode 'LogContract' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/LogContract'"
  fi
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'GenerateCertificate'" "U1F60E"
    chaincodeBuild "GenerateCertificate" "golang" "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate" "16"
    chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" "golang" printHeadline "Installing 'GenerateCertificate' for Pharma" "U1F60E"
    chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" ""
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printHeadline "Installing 'GenerateCertificate' for Hospital" "U1F60E"
    chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "GenerateCertificate" "$version" ""
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printHeadline "Installing 'GenerateCertificate' for IoTTracker" "U1F60E"
    chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "GenerateCertificate" "$version" ""
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printItalics "Committing chaincode 'GenerateCertificate' on channel 'supply' as 'Pharma'" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""
  else
    echo "Warning! Skipping chaincode 'GenerateCertificate' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate'"
  fi
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore")" ]; then
    local version="0.0.1"
    printHeadline "Packaging chaincode 'DeviceStore'" "U1F60E"
    chaincodeBuild "DeviceStore" "golang" "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore" "16"
    chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" "golang" printHeadline "Installing 'DeviceStore' for Pharma" "U1F60E"
    chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" ""
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printHeadline "Installing 'DeviceStore' for Hospital" "U1F60E"
    chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "DeviceStore" "$version" ""
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printHeadline "Installing 'DeviceStore' for IoTTracker" "U1F60E"
    chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "DeviceStore" "$version" ""
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printItalics "Committing chaincode 'DeviceStore' on channel 'supply' as 'Pharma'" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""
  else
    echo "Warning! Skipping chaincode 'DeviceStore' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/DeviceStore'"
  fi

}

installChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "LogContract" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/LogContract")" ]; then
      printHeadline "Packaging chaincode 'LogContract'" "U1F60E"
      chaincodeBuild "LogContract" "golang" "$CHAINCODES_BASE_DIR/./chaincode/LogContract" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" "golang" printHeadline "Installing 'LogContract' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'LogContract' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "LogContract" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'LogContract' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "LogContract" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printItalics "Committing chaincode 'LogContract' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'LogContract' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/LogContract'"
    fi
  fi
  if [ "$chaincodeName" = "GenerateCertificate" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate")" ]; then
      printHeadline "Packaging chaincode 'GenerateCertificate'" "U1F60E"
      chaincodeBuild "GenerateCertificate" "golang" "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" "golang" printHeadline "Installing 'GenerateCertificate' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printHeadline "Installing 'GenerateCertificate' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printHeadline "Installing 'GenerateCertificate' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printItalics "Committing chaincode 'GenerateCertificate' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'GenerateCertificate' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate'"
    fi
  fi
  if [ "$chaincodeName" = "DeviceStore" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore")" ]; then
      printHeadline "Packaging chaincode 'DeviceStore'" "U1F60E"
      chaincodeBuild "DeviceStore" "golang" "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" "golang" printHeadline "Installing 'DeviceStore' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'DeviceStore' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "DeviceStore" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'DeviceStore' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "DeviceStore" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printItalics "Committing chaincode 'DeviceStore' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'DeviceStore' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/DeviceStore'"
    fi
  fi
}

runDevModeChaincode() {
  local chaincodeName=$1
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "LogContract" ]; then
    local version="0.0.1"
    printHeadline "Approving 'LogContract' for Pharma (dev mode)" "U1F60E"
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "0.0.1" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printHeadline "Approving 'LogContract' for Hospital (dev mode)" "U1F60E"
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "LogContract" "0.0.1" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printHeadline "Approving 'LogContract' for IoTTracker (dev mode)" "U1F60E"
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "LogContract" "0.0.1" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
    printItalics "Committing chaincode 'LogContract' on channel 'supply' as 'Pharma' (dev mode)" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "0.0.1" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

  fi
  if [ "$chaincodeName" = "GenerateCertificate" ]; then
    local version="0.0.1"
    printHeadline "Approving 'GenerateCertificate' for Pharma (dev mode)" "U1F60E"
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printHeadline "Approving 'GenerateCertificate' for Hospital (dev mode)" "U1F60E"
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "GenerateCertificate" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printHeadline "Approving 'GenerateCertificate' for IoTTracker (dev mode)" "U1F60E"
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "GenerateCertificate" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
    printItalics "Committing chaincode 'GenerateCertificate' on channel 'supply' as 'Pharma' (dev mode)" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

  fi
  if [ "$chaincodeName" = "DeviceStore" ]; then
    local version="0.0.1"
    printHeadline "Approving 'DeviceStore' for Pharma (dev mode)" "U1F60E"
    chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printHeadline "Approving 'DeviceStore' for Hospital (dev mode)" "U1F60E"
    chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "DeviceStore" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printHeadline "Approving 'DeviceStore' for IoTTracker (dev mode)" "U1F60E"
    chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "DeviceStore" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
    printItalics "Committing chaincode 'DeviceStore' on channel 'supply' as 'Pharma' (dev mode)" "U1F618"
    chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "0.0.1" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

  fi
}

upgradeChaincode() {
  local chaincodeName="$1"
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  local version="$2"
  if [ -z "$version" ]; then
    echo "Error: chaincode version is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "LogContract" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/LogContract")" ]; then
      printHeadline "Packaging chaincode 'LogContract'" "U1F60E"
      chaincodeBuild "LogContract" "golang" "$CHAINCODES_BASE_DIR/./chaincode/LogContract" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" "golang" printHeadline "Installing 'LogContract' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "LogContract" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'LogContract' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "LogContract" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'LogContract' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "LogContract" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" ""
      printItalics "Committing chaincode 'LogContract' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "LogContract" "$version" "orderer0.group1.orderer.example.com:7030" "OR('Pharma.member', 'IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'LogContract' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/LogContract'"
    fi
  fi
  if [ "$chaincodeName" = "GenerateCertificate" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate")" ]; then
      printHeadline "Packaging chaincode 'GenerateCertificate'" "U1F60E"
      chaincodeBuild "GenerateCertificate" "golang" "$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" "golang" printHeadline "Installing 'GenerateCertificate' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printHeadline "Installing 'GenerateCertificate' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printHeadline "Installing 'GenerateCertificate' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "GenerateCertificate" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" ""
      printItalics "Committing chaincode 'GenerateCertificate' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "GenerateCertificate" "$version" "orderer0.group1.orderer.example.com:7030" "AND('Hospital.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'GenerateCertificate' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/GenerateCertificate'"
    fi
  fi
  if [ "$chaincodeName" = "DeviceStore" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore")" ]; then
      printHeadline "Packaging chaincode 'DeviceStore'" "U1F60E"
      chaincodeBuild "DeviceStore" "golang" "$CHAINCODES_BASE_DIR/./chaincode/DeviceStore" "16"
      chaincodePackage "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" "golang" printHeadline "Installing 'DeviceStore' for Pharma" "U1F60E"
      chaincodeInstall "cli.pharma.example.com" "peer0.pharma.example.com:7041" "DeviceStore" "$version" ""
      chaincodeApprove "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'DeviceStore' for Hospital" "U1F60E"
      chaincodeInstall "cli.hospital.example.com" "peer0.hospital.example.com:7061" "DeviceStore" "$version" ""
      chaincodeApprove "cli.hospital.example.com" "peer0.hospital.example.com:7061" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printHeadline "Installing 'DeviceStore' for IoTTracker" "U1F60E"
      chaincodeInstall "cli.tracker.example.com" "peer0.tracker.example.com:7081" "DeviceStore" "$version" ""
      chaincodeApprove "cli.tracker.example.com" "peer0.tracker.example.com:7081" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" ""
      printItalics "Committing chaincode 'DeviceStore' on channel 'supply' as 'Pharma'" "U1F618"
      chaincodeCommit "cli.pharma.example.com" "peer0.pharma.example.com:7041" "supply" "DeviceStore" "$version" "orderer0.group1.orderer.example.com:7030" "AND('IoTTracker.member')" "false" "" "peer0.pharma.example.com:7041,peer0.hospital.example.com:7061,peer0.tracker.example.com:7081" "" ""

    else
      echo "Warning! Skipping chaincode 'DeviceStore' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincode/DeviceStore'"
    fi
  fi
}

notifyOrgsAboutChannels() {
  printHeadline "Creating new channel config blocks" "U1F537"
  createNewChannelUpdateTx "supply" "PharmaMSP" "Supply" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "supply" "HospitalMSP" "Supply" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "supply" "IoTTrackerMSP" "Supply" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"

  printHeadline "Notyfing orgs about channels" "U1F4E2"
  notifyOrgAboutNewChannel "supply" "PharmaMSP" "cli.pharma.example.com" "peer0.pharma.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "supply" "HospitalMSP" "cli.hospital.example.com" "peer0.hospital.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "supply" "IoTTrackerMSP" "cli.tracker.example.com" "peer0.tracker.example.com" "orderer0.group1.orderer.example.com:7030"

  printHeadline "Deleting new channel config blocks" "U1F52A"
  deleteNewChannelUpdateTx "supply" "PharmaMSP" "cli.pharma.example.com"
  deleteNewChannelUpdateTx "supply" "HospitalMSP" "cli.hospital.example.com"
  deleteNewChannelUpdateTx "supply" "IoTTrackerMSP" "cli.tracker.example.com"
}

printStartSuccessInfo() {
  printHeadline "Done! Enjoy your fresh network" "U1F984"
}

stopNetwork() {
  printHeadline "Stopping network" "U1F68F"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose stop)
  sleep 4
}

networkDown() {
  printHeadline "Destroying network" "U1F916"
  (cd "$FABLO_NETWORK_ROOT"/fabric-docker && docker-compose down)

  printf "Removing chaincode containers & images... \U1F5D1 \n"
  for container in $(docker ps -a | grep "dev-peer0.pharma.example.com-LogContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.pharma.example.com-LogContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.hospital.example.com-LogContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.hospital.example.com-LogContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.tracker.example.com-LogContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.tracker.example.com-LogContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.pharma.example.com-GenerateCertificate" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.pharma.example.com-GenerateCertificate*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.hospital.example.com-GenerateCertificate" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.hospital.example.com-GenerateCertificate*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.tracker.example.com-GenerateCertificate" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.tracker.example.com-GenerateCertificate*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.pharma.example.com-DeviceStore" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.pharma.example.com-DeviceStore*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.hospital.example.com-DeviceStore" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.hospital.example.com-DeviceStore*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.tracker.example.com-DeviceStore" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.tracker.example.com-DeviceStore*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done

  printf "Removing generated configs... \U1F5D1 \n"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/crypto-config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"

  printHeadline "Done! Network was purged" "U1F5D1"
}
