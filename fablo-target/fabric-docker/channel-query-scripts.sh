#!/usr/bin/env bash

source "$FABLO_NETWORK_ROOT/fabric-docker/scripts/channel-query-functions.sh"

set -eu

channelQuery() {
  echo "-> Channel query: " + "$@"

  if [ "$#" -eq 1 ]; then
    printChannelsHelp

  elif [ "$1" = "list" ] && [ "$2" = "supplier" ] && [ "$3" = "peer0" ]; then

    peerChannelList "cli.supplier.example.com" "peer0.supplier.example.com:7041"

  elif
    [ "$1" = "list" ] && [ "$2" = "supplier" ] && [ "$3" = "peer1" ]
  then

    peerChannelList "cli.supplier.example.com" "peer1.supplier.example.com:7042"

  elif
    [ "$1" = "list" ] && [ "$2" = "manufacturer" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061"

  elif
    [ "$1" = "list" ] && [ "$2" = "distributor" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.distributor.example.com" "peer0.distributor.example.com:7081"

  elif
    [ "$1" = "list" ] && [ "$2" = "distributor" ] && [ "$3" = "peer1" ]
  then

    peerChannelList "cli.distributor.example.com" "peer1.distributor.example.com:7082"

  elif

    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "supplier" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "supply" "cli.supplier.example.com" "peer0.supplier.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "supplier" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.supplier.example.com" "$TARGET_FILE" "peer0.supplier.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "supplier" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.supplier.example.com" "${BLOCK_NAME}" "peer0.supplier.example.com:7041" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "supplier" ] && [ "$4" = "peer1" ]
  then

    peerChannelGetInfo "supply" "cli.supplier.example.com" "peer1.supplier.example.com:7042"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "supplier" ] && [ "$5" = "peer1" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.supplier.example.com" "$TARGET_FILE" "peer1.supplier.example.com:7042"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "supplier" ] && [ "$5" = "peer1" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.supplier.example.com" "${BLOCK_NAME}" "peer1.supplier.example.com:7042" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "manufacturer" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "supply" "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "manufacturer" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.manufacturer.example.com" "$TARGET_FILE" "peer0.manufacturer.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "manufacturer" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.manufacturer.example.com" "${BLOCK_NAME}" "peer0.manufacturer.example.com:7061" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "distribution" ] && [ "$3" = "manufacturer" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "distribution" "cli.manufacturer.example.com" "peer0.manufacturer.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "distribution" ] && [ "$4" = "manufacturer" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "distribution" "cli.manufacturer.example.com" "$TARGET_FILE" "peer0.manufacturer.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$3" = "distribution" ] && [ "$4" = "manufacturer" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "distribution" "cli.manufacturer.example.com" "${BLOCK_NAME}" "peer0.manufacturer.example.com:7061" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "distribution" ] && [ "$3" = "distributor" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "distribution" "cli.distributor.example.com" "peer0.distributor.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "distribution" ] && [ "$4" = "distributor" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "distribution" "cli.distributor.example.com" "$TARGET_FILE" "peer0.distributor.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$3" = "distribution" ] && [ "$4" = "distributor" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "distribution" "cli.distributor.example.com" "${BLOCK_NAME}" "peer0.distributor.example.com:7081" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "distribution" ] && [ "$3" = "distributor" ] && [ "$4" = "peer1" ]
  then

    peerChannelGetInfo "distribution" "cli.distributor.example.com" "peer1.distributor.example.com:7082"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "distribution" ] && [ "$4" = "distributor" ] && [ "$5" = "peer1" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "distribution" "cli.distributor.example.com" "$TARGET_FILE" "peer1.distributor.example.com:7082"

  elif [ "$1" = "fetch" ] && [ "$3" = "distribution" ] && [ "$4" = "distributor" ] && [ "$5" = "peer1" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "distribution" "cli.distributor.example.com" "${BLOCK_NAME}" "peer1.distributor.example.com:7082" "$TARGET_FILE"

  else

    echo "$@"
    echo "$1, $2, $3, $4, $5, $6, $7, $#"
    printChannelsHelp
  fi

}

printChannelsHelp() {
  echo "Channel management commands:"
  echo ""

  echo "fablo channel list supplier peer0"
  echo -e "\t List channels on 'peer0' of 'Supplier'".
  echo ""

  echo "fablo channel list supplier peer1"
  echo -e "\t List channels on 'peer1' of 'Supplier'".
  echo ""

  echo "fablo channel list manufacturer peer0"
  echo -e "\t List channels on 'peer0' of 'Manufacturer'".
  echo ""

  echo "fablo channel list distributor peer0"
  echo -e "\t List channels on 'peer0' of 'Distributor'".
  echo ""

  echo "fablo channel list distributor peer1"
  echo -e "\t List channels on 'peer1' of 'Distributor'".
  echo ""

  echo "fablo channel getinfo supply supplier peer0"
  echo -e "\t Get channel info on 'peer0' of 'Supplier'".
  echo ""
  echo "fablo channel fetch config supply supplier peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Supplier'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply supplier peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Supplier'".
  echo ""

  echo "fablo channel getinfo supply supplier peer1"
  echo -e "\t Get channel info on 'peer1' of 'Supplier'".
  echo ""
  echo "fablo channel fetch config supply supplier peer1 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer1' of 'Supplier'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply supplier peer1 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer1' of 'Supplier'".
  echo ""

  echo "fablo channel getinfo supply manufacturer peer0"
  echo -e "\t Get channel info on 'peer0' of 'Manufacturer'".
  echo ""
  echo "fablo channel fetch config supply manufacturer peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Manufacturer'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply manufacturer peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Manufacturer'".
  echo ""

  echo "fablo channel getinfo distribution manufacturer peer0"
  echo -e "\t Get channel info on 'peer0' of 'Manufacturer'".
  echo ""
  echo "fablo channel fetch config distribution manufacturer peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Manufacturer'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> distribution manufacturer peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Manufacturer'".
  echo ""

  echo "fablo channel getinfo distribution distributor peer0"
  echo -e "\t Get channel info on 'peer0' of 'Distributor'".
  echo ""
  echo "fablo channel fetch config distribution distributor peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Distributor'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> distribution distributor peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Distributor'".
  echo ""

  echo "fablo channel getinfo distribution distributor peer1"
  echo -e "\t Get channel info on 'peer1' of 'Distributor'".
  echo ""
  echo "fablo channel fetch config distribution distributor peer1 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer1' of 'Distributor'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> distribution distributor peer1 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer1' of 'Distributor'".
  echo ""

}
