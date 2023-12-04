# TraceChain

## Command to invoke smart contract

Create a raw material
```bash
docker exec -e CORE_PEER_ADDRESS="peer0.supplier.example.com:7041" "cli.supplier.example.com" peer chaincode invoke -o orderer0.group1.orderer.example.com:7030 -C supply -n rawMaterialTokenContract --peerAddresses peer0.supplier.example.com:7041 --peerAddresses peer1.supplier.example.com:7042 --peerAddresses peer0.manufacturer.example.com:7061 -c '{"Args":["CreateRawMaterial", "Cocoa"]}'
```

Create a raw material, but fail due to manufacturer trying
```bash
docker exec -e CORE_PEER_ADDRESS="peer0.manufacturer.example.com:7061" "cli.manufacturer.example.com" peer chaincode invoke -o orderer0.group1.orderer.example.com:7030 -C supply -n rawMaterialTokenContract --peerAddresses peer0.supplier.example.com:7041 --peerAddresses peer1.supplier.example.com:7042 --peerAddresses peer0.manufacturer.example.com:7061 -c '{"Args":["CreateRawMaterial", "Milk"]}'
```

Initiate a transfer of raw material
```bash
docker exec -e CORE_PEER_ADDRESS="peer0.supplier.example.com:7041" "cli.supplier.example.com" peer chaincode invoke -o orderer0.group1.orderer.example.com:7030 -C supply -n rawMaterialTokenContract --peerAddresses peer0.supplier.example.com:7041 --peerAddresses peer1.supplier.example.com:7042 --peerAddresses peer0.manufacturer.example.com:7061 -c '{"Args":["InitiateTransferToManufacturer", "1", "ManufacturerMSP"]}'
```

Accept a transfer of raw material
```bash
docker exec -e CORE_PEER_ADDRESS="peer0.manufacturer.example.com:7061" "cli.manufacturer.example.com" peer chaincode invoke -o orderer0.group1.orderer.example.com:7030 -C supply -n rawMaterialTokenContract --peerAddresses peer0.supplier.example.com:7041 --peerAddresses peer1.supplier.example.com:7042 --peerAddresses peer0.manufacturer.example.com:7061 -c '{"Args":["CompleteTransferToManufacturer", "1", "SupplierMSP"]}'
```



Burn a raw material to generate product
```bash
docker exec -e CORE_PEER_ADDRESS="peer0.manufacturer.example.com:7061" "cli.manufacturer.example.com" peer chaincode invoke -o orderer0.group1.orderer.example.com:7030 -C distribution -n productTokenContract --peerAddresses peer0.supplier.example.com:7041 --peerAddresses peer1.supplier.example.com:7042 --peerAddresses peer0.manufacturer.example.com:7061 -c '{"Args":["CreateProduct", "Cocoa", "Chocolate"]}'
```