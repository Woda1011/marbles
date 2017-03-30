package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}



var artefactIndexStr = "_artefactindex"
var deviceIndexStr = "_deviceIndex"

type Artefact struct{
	Version string `json:"artefactVersion"`
	Name 	string `json:"artefactName"`
	Hash 	string `json:"artefactHash"`
	//ManufacturerId string `json:"manufacturererId"`
	//Size string `json:"size"`
	//prevVersionBlockHash string `json:"prevVersionBlockHash"`
	//prevArtefactBlockHash string `json:"prevArtefactBlockHash"`
	ArtefactType string `json:"artefactType"`
	Timestamp string `json:"timestamp"`
}

type Device struct {
	DeviceId string `json:"deviceId"`
	CurrentArtifactHash string `json:"currentArtifactHash "`
}

// Maybe needed to keep track of all deployment transactions
type Deployment struct {
	DeviceId string
	CurrentArtefactHash string
	Timestamp int64
	TransactionType string
}


// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Artefact Deployment chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	//marshal an emtpy array of strings to clear the index for the artefacts and the devices
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the initial device")
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)
	err = stub.PutState(artefactIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(deviceIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return t.init_device(stub, args);
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "delete" {
		res, err := t.Delete(stub, args)
		return res, err
	} else if function == "init_artefact" {
		return t.init_artefact(stub, args)
	} else if function == "init_device" {
		return t.init_device(stub, args)
	} else if function == "deploy_artifact" {
		res, err := t.deploy_artifact(stub, args)
		return res, err
	}
	fmt.Println("invoke did not find func: " + function)
	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	//Read Device State
	//Get Artefact Information
	//Get Artefact
	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	fmt.Println("Trying to get state for: " + name)

	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//TODO Update Device index if a device should be deleted

	//get the artefact index
	artefactsAsBytes, err := stub.GetState(artefactIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get artefact index")
	}
	var artefactIndex []string
	json.Unmarshal(artefactsAsBytes, &artefactIndex)
	
	//remove artefact from index
	for i,val := range artefactIndex {
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{
			fmt.Println("found artefact")
			artefactIndex = append(artefactIndex[:i], artefactIndex[i+1:]...)
			for x:= range artefactIndex {
				fmt.Println(string(x) + " - " + artefactIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(artefactIndex)
	err = stub.PutState(artefactIndexStr, jsonAsBytes)
	return nil, nil
}


// ============================================================================================================================
// Init Artefact - create a new artefact, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_artefact(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error


	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init artefact")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}

	if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}

	version := args[0]
	name 	:= args[1]
	hash 	:= args[2]
	artefactType := args[3]
	artefact := args[4]
	timestamp := makeTimestamp()

	//check if already exists
	artefactAsBytes, err := stub.GetState(hash)
	if err != nil {
		return nil, errors.New("Failed to get Artefact hash")
	}
	res := Artefact{}
	json.Unmarshal(artefactAsBytes, &res)
	if res.Hash == hash{
		fmt.Println("This Artefact arleady exists: " + hash)
		fmt.Println(res);
		return nil, errors.New("This Artefact arleady exists")
	}

	//build the json string manually
	str := `{"artefactVersion": "` + version + `", "artefactType": "` + artefactType + `", "artefactName": "` + name + `", "hash": "` + hash + `", "artefact":` + artefact +`, "timestamp": "` + strconv.FormatInt(timestamp, 10) + `"}`
	err = stub.PutState(hash, []byte(str))
	if err != nil {
		return nil, err
	}
		
	//get the artefacts index
	artefactsAsBytes, err := stub.GetState(artefactIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get artefact index")
	}
	var artefactIndex []string

	//un stringify it aka JSON.parse()
	json.Unmarshal(artefactsAsBytes, &artefactIndex)
	
	//append
	artefactIndex = append(artefactIndex, hash)
	fmt.Println("! artefact index: ", artefactIndex)
	jsonAsBytes, _ := json.Marshal(artefactIndex)
	err = stub.PutState(artefactIndexStr, jsonAsBytes)

	fmt.Println("- end init artefact")
	return nil, nil
}

// ============================================================================================================================
// Init Device - create a new Device, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_device(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	fmt.Println("- start init device")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}

	deviceId := args[0]

	//check if device already exists
	deviceAsBytes, err := stub.GetState(deviceId)
	if err != nil {
		return nil, errors.New("Failed to get Device for deviceId")
	}

	res := Device{}
	json.Unmarshal(deviceAsBytes, &res)
	if res.DeviceId == deviceId {
		fmt.Println("This Device arleady exists: " + deviceId)
		fmt.Println(res);
		return nil, errors.New("This Device arleady exists")
	}

	//build the json string manually
	str := `{"deviceId": "` + deviceId + `", "currentArtifactHash": "" }`
	err = stub.PutState(deviceId, []byte(str))
	if err != nil {
		return nil, err
	}

	//get the device index
	devicesAsBytes, err := stub.GetState(deviceIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get device index")
	}
	var deviceIndex []string
	json.Unmarshal(devicesAsBytes, &deviceIndex)

	//append
	deviceIndex = append(deviceIndex, deviceId)
	fmt.Println("! device index: ", deviceIndex)
	jsonAsBytes, _ := json.Marshal(deviceIndex)
	err = stub.PutState(deviceIndexStr, jsonAsBytes)

	fmt.Println("- end init device")
	return nil, nil
}

// ============================================================================================================================
// Set current Artifact Hash of a Device
// ============================================================================================================================
func (t *SimpleChaincode) deploy_artifact(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}

	if len(args[1]) <= 0 {
		return nil, errors.New("2st argument must be a non-empty string")
	}

	fmt.Println("- deploy")
	fmt.Println(args[0] + " - " + args[1])

	deviceAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get device")
	}

	artifactAsBytes, err := stub.GetState(args[1])
	if err != nil {
		return artifactAsBytes, errors.New("Failed to get artifact")
	}

	res := Device{}
	json.Unmarshal(deviceAsBytes, &res)

	res.CurrentArtifactHash = args[1]

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)
	if err != nil {
		return nil, err
	}

	fmt.Println("- end deploy artefact")
	return nil, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

