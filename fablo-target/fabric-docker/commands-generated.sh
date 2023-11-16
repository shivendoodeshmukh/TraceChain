#!/usr/bin/env bash

generateArtifacts() {
  printHeadline "Generating basic configs" "U1F913"

  printItalics "Generating crypto material for Orderer" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-orderer.yaml" "peerOrganizations/orderer.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Supplier" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-supplier.yaml" "peerOrganizations/supplier.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Manufacturer" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-manufacturer.yaml" "peerOrganizations/manufacturer.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

  printItalics "Generating crypto material for Distributor" "U1F512"
  certsGenerate "$FABLO_NETWORK_ROOT/fabric-config" "crypto-config-distributor.yaml" "peerOrganizations/distributor.example.com" "$FABLO_NETWORK_ROOT/fabric-config/crypto-config/"

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
  printHeadline "Generating config for 'distribution'" "U1F913"
  createChannelTx "distribution" "$FABLO_NETWORK_ROOT/fabric-config" "Distribution" "$FABLO_NETWORK_ROOT/fabric-config/config"
}

installChannels() {
  printHeadline "Creating 'supply' on Supplier/peer0" "U1F63B"
  docker exec -i cli.supplier.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoin 'supply' 'SupplierMSP' 'peer0.supplier.example.com:7041' 'crypto/users/Admin@supplier.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'supply' on  Supplier/peer1" "U1F638"
  docker exec -i cli.supplier.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'supply' 'SupplierMSP' 'peer1.supplier.example.com:7042' 'crypto/users/Admin@supplier.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
  printItalics "Joining 'supply' on  Manufacturer/peer0" "U1F638"
  docker exec -i cli.manufacturer.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'supply' 'ManufacturerMSP' 'peer0.manufacturer.example.com:7061' 'crypto/users/Admin@manufacturer.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
  printHeadline "Creating 'distribution' on Manufacturer/peer0" "U1F63B"
  docker exec -i cli.manufacturer.example.com bash -c "source scripts/channel_fns.sh; createChannelAndJoin 'distribution' 'ManufacturerMSP' 'peer0.manufacturer.example.com:7061' 'crypto/users/Admin@manufacturer.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"

  printItalics "Joining 'distribution' on  Distributor/peer0" "U1F638"
  docker exec -i cli.distributor.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'distribution' 'DistributorMSP' 'peer0.distributor.example.com:7081' 'crypto/users/Admin@distributor.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
  printItalics "Joining 'distribution' on  Distributor/peer1" "U1F638"
  docker exec -i cli.distributor.example.com bash -c "source scripts/channel_fns.sh; fetchChannelAndJoin 'distribution' 'DistributorMSP' 'peer1.distributor.example.com:7082' 'crypto/users/Admin@distributor.example.com/msp' 'orderer0.group1.orderer.example.com:7030';"
}

installChaincodes() {
  if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes")" ]; then
    local version="0.0.3"
    printHeadline "Packaging chaincode 'rawMaterialTokenContract'" "U1F60E"
    chaincodeBuild "rawMaterialTokenContract" "golang" "$CHAINCODES_BASE_DIR/./chaincodes" "16"
    chaincodePackage "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" "golang" printHeadline "Installing 'rawMaterialTokenContract' for Supplier" "U1F60E"
    chaincodeInstall "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" ""
    chaincodeInstall "cli.supplier.example.com" "peer1.supplier.example.com:7042" "rawMaterialTokenContract" "$version" ""
    chaincodeApprove "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Installing 'rawMaterialTokenContract' for Manufacturer" "U1F60E"
    chaincodeInstall "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "rawMaterialTokenContract" "$version" ""
    chaincodeApprove "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'rawMaterialTokenContract' on channel 'supply' as 'Supplier'" "U1F618"
    chaincodeCommit "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.supplier.example.com:7041,peer0.manufacturer.example.com:7061" "" ""
  else
    echo "Warning! Skipping chaincode 'rawMaterialTokenContract' installation. Chaincode directory is empty."
    echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes'"
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

  if [ "$chaincodeName" = "rawMaterialTokenContract" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes")" ]; then
      printHeadline "Packaging chaincode 'rawMaterialTokenContract'" "U1F60E"
      chaincodeBuild "rawMaterialTokenContract" "golang" "$CHAINCODES_BASE_DIR/./chaincodes" "16"
      chaincodePackage "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" "golang" printHeadline "Installing 'rawMaterialTokenContract' for Supplier" "U1F60E"
      chaincodeInstall "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" ""
      chaincodeInstall "cli.supplier.example.com" "peer1.supplier.example.com:7042" "rawMaterialTokenContract" "$version" ""
      chaincodeApprove "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printHeadline "Installing 'rawMaterialTokenContract' for Manufacturer" "U1F60E"
      chaincodeInstall "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "rawMaterialTokenContract" "$version" ""
      chaincodeApprove "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printItalics "Committing chaincode 'rawMaterialTokenContract' on channel 'supply' as 'Supplier'" "U1F618"
      chaincodeCommit "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.supplier.example.com:7041,peer0.manufacturer.example.com:7061" "" ""

    else
      echo "Warning! Skipping chaincode 'rawMaterialTokenContract' install. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes'"
    fi
  fi
}

runDevModeChaincode() {
  local chaincodeName=$1
  if [ -z "$chaincodeName" ]; then
    echo "Error: chaincode name is not provided"
    exit 1
  fi

  if [ "$chaincodeName" = "rawMaterialTokenContract" ]; then
    local version="0.0.3"
    printHeadline "Approving 'rawMaterialTokenContract' for Supplier (dev mode)" "U1F60E"
    chaincodeApprove "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "0.0.3" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printHeadline "Approving 'rawMaterialTokenContract' for Manufacturer (dev mode)" "U1F60E"
    chaincodeApprove "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "supply" "rawMaterialTokenContract" "0.0.3" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
    printItalics "Committing chaincode 'rawMaterialTokenContract' on channel 'supply' as 'Supplier' (dev mode)" "U1F618"
    chaincodeCommit "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "0.0.3" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.supplier.example.com:7041,peer0.manufacturer.example.com:7061" "" ""

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

  if [ "$chaincodeName" = "rawMaterialTokenContract" ]; then
    if [ -n "$(ls "$CHAINCODES_BASE_DIR/./chaincodes")" ]; then
      printHeadline "Packaging chaincode 'rawMaterialTokenContract'" "U1F60E"
      chaincodeBuild "rawMaterialTokenContract" "golang" "$CHAINCODES_BASE_DIR/./chaincodes" "16"
      chaincodePackage "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" "golang" printHeadline "Installing 'rawMaterialTokenContract' for Supplier" "U1F60E"
      chaincodeInstall "cli.supplier.example.com" "peer0.supplier.example.com:7041" "rawMaterialTokenContract" "$version" ""
      chaincodeInstall "cli.supplier.example.com" "peer1.supplier.example.com:7042" "rawMaterialTokenContract" "$version" ""
      chaincodeApprove "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printHeadline "Installing 'rawMaterialTokenContract' for Manufacturer" "U1F60E"
      chaincodeInstall "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "rawMaterialTokenContract" "$version" ""
      chaincodeApprove "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" ""
      printItalics "Committing chaincode 'rawMaterialTokenContract' on channel 'supply' as 'Supplier'" "U1F618"
      chaincodeCommit "cli.supplier.example.com" "peer0.supplier.example.com:7041" "supply" "rawMaterialTokenContract" "$version" "orderer0.group1.orderer.example.com:7030" "" "false" "" "peer0.supplier.example.com:7041,peer0.manufacturer.example.com:7061" "" ""

    else
      echo "Warning! Skipping chaincode 'rawMaterialTokenContract' upgrade. Chaincode directory is empty."
      echo "Looked in dir: '$CHAINCODES_BASE_DIR/./chaincodes'"
    fi
  fi
}

notifyOrgsAboutChannels() {
  printHeadline "Creating new channel config blocks" "U1F537"
  createNewChannelUpdateTx "supply" "SupplierMSP" "Supply" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "supply" "ManufacturerMSP" "Supply" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "distribution" "ManufacturerMSP" "Distribution" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"
  createNewChannelUpdateTx "distribution" "DistributorMSP" "Distribution" "$FABLO_NETWORK_ROOT/fabric-config" "$FABLO_NETWORK_ROOT/fabric-config/config"

  printHeadline "Notyfing orgs about channels" "U1F4E2"
  notifyOrgAboutNewChannel "supply" "SupplierMSP" "cli.supplier.example.com" "peer0.supplier.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "supply" "ManufacturerMSP" "cli.manufacturer.example.com" "peer0.manufacturer.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "distribution" "ManufacturerMSP" "cli.manufacturer.example.com" "peer0.manufacturer.example.com" "orderer0.group1.orderer.example.com:7030"
  notifyOrgAboutNewChannel "distribution" "DistributorMSP" "cli.distributor.example.com" "peer0.distributor.example.com" "orderer0.group1.orderer.example.com:7030"

  printHeadline "Deleting new channel config blocks" "U1F52A"
  deleteNewChannelUpdateTx "supply" "SupplierMSP" "cli.supplier.example.com"
  deleteNewChannelUpdateTx "supply" "ManufacturerMSP" "cli.manufacturer.example.com"
  deleteNewChannelUpdateTx "distribution" "ManufacturerMSP" "cli.manufacturer.example.com"
  deleteNewChannelUpdateTx "distribution" "DistributorMSP" "cli.distributor.example.com"
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
  for container in $(docker ps -a | grep "dev-peer0.supplier.example.com-rawMaterialTokenContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.supplier.example.com-rawMaterialTokenContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer1.supplier.example.com-rawMaterialTokenContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer1.supplier.example.com-rawMaterialTokenContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done
  for container in $(docker ps -a | grep "dev-peer0.manufacturer.example.com-rawMaterialTokenContract" | awk '{print $1}'); do
    echo "Removing container $container..."
    docker rm -f "$container" || echo "docker rm of $container failed. Check if all fabric dockers properly was deleted"
  done
  for image in $(docker images "dev-peer0.manufacturer.example.com-rawMaterialTokenContract*" -q); do
    echo "Removing image $image..."
    docker rmi "$image" || echo "docker rmi of $image failed. Check if all fabric dockers properly was deleted"
  done

  printf "Removing generated configs... \U1F5D1 \n"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/crypto-config"
  rm -rf "$FABLO_NETWORK_ROOT/fabric-config/chaincode-packages"

  printHeadline "Done! Network was purged" "U1F5D1"
}
