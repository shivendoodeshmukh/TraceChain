#!/usr/bin/env bash

source "$FABLO_NETWORK_ROOT/fabric-docker/scripts/channel-query-functions.sh"

set -eu

channelQuery() {
  echo "-> Channel query: " + "$@"

  if [ "$#" -eq 1 ]; then
    printChannelsHelp

  elif [ "$1" = "list" ] && [ "$2" = "pharma" ] && [ "$3" = "peer0" ]; then

    peerChannelList "cli.pharma.example.com" "peer0.pharma.example.com:7041"

  elif
    [ "$1" = "list" ] && [ "$2" = "hospital" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.hospital.example.com" "peer0.hospital.example.com:7061"

  elif
    [ "$1" = "list" ] && [ "$2" = "iottracker" ] && [ "$3" = "peer0" ]
  then

    peerChannelList "cli.tracker.example.com" "peer0.tracker.example.com:7081"

  elif

    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "pharma" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "supply" "cli.pharma.example.com" "peer0.pharma.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "pharma" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.pharma.example.com" "$TARGET_FILE" "peer0.pharma.example.com:7041"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "pharma" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.pharma.example.com" "${BLOCK_NAME}" "peer0.pharma.example.com:7041" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "hospital" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "supply" "cli.hospital.example.com" "peer0.hospital.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "hospital" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.hospital.example.com" "$TARGET_FILE" "peer0.hospital.example.com:7061"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "hospital" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.hospital.example.com" "${BLOCK_NAME}" "peer0.hospital.example.com:7061" "$TARGET_FILE"

  elif
    [ "$1" = "getinfo" ] && [ "$2" = "supply" ] && [ "$3" = "iottracker" ] && [ "$4" = "peer0" ]
  then

    peerChannelGetInfo "supply" "cli.tracker.example.com" "peer0.tracker.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$2" = "config" ] && [ "$3" = "supply" ] && [ "$4" = "iottracker" ] && [ "$5" = "peer0" ]; then
    TARGET_FILE=${6:-"$channel-config.json"}

    peerChannelFetchConfig "supply" "cli.tracker.example.com" "$TARGET_FILE" "peer0.tracker.example.com:7081"

  elif [ "$1" = "fetch" ] && [ "$3" = "supply" ] && [ "$4" = "iottracker" ] && [ "$5" = "peer0" ]; then
    BLOCK_NAME=$2
    TARGET_FILE=${6:-"$BLOCK_NAME.block"}

    peerChannelFetchBlock "supply" "cli.tracker.example.com" "${BLOCK_NAME}" "peer0.tracker.example.com:7081" "$TARGET_FILE"

  else

    echo "$@"
    echo "$1, $2, $3, $4, $5, $6, $7, $#"
    printChannelsHelp
  fi

}

printChannelsHelp() {
  echo "Channel management commands:"
  echo ""

  echo "fablo channel list pharma peer0"
  echo -e "\t List channels on 'peer0' of 'Pharma'".
  echo ""

  echo "fablo channel list hospital peer0"
  echo -e "\t List channels on 'peer0' of 'Hospital'".
  echo ""

  echo "fablo channel list iottracker peer0"
  echo -e "\t List channels on 'peer0' of 'IoTTracker'".
  echo ""

  echo "fablo channel getinfo supply pharma peer0"
  echo -e "\t Get channel info on 'peer0' of 'Pharma'".
  echo ""
  echo "fablo channel fetch config supply pharma peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Pharma'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply pharma peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Pharma'".
  echo ""

  echo "fablo channel getinfo supply hospital peer0"
  echo -e "\t Get channel info on 'peer0' of 'Hospital'".
  echo ""
  echo "fablo channel fetch config supply hospital peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'Hospital'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply hospital peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'Hospital'".
  echo ""

  echo "fablo channel getinfo supply iottracker peer0"
  echo -e "\t Get channel info on 'peer0' of 'IoTTracker'".
  echo ""
  echo "fablo channel fetch config supply iottracker peer0 [file-name.json]"
  echo -e "\t Download latest config block and save it. Uses first peer 'peer0' of 'IoTTracker'".
  echo ""
  echo "fablo channel fetch <newest|oldest|block-number> supply iottracker peer0 [file name]"
  echo -e "\t Fetch a block with given number and save it. Uses first peer 'peer0' of 'IoTTracker'".
  echo ""

}
