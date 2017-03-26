package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}



var artefactIndexStr = "_artefactindex"
var deviceIndexStr = "_deviceindex"

type Artefact struct{
	Version string `json:"version"`					//the fieldtags are needed to keep case from bouncing around
	Name 	string `json:"name"`
	Hash 	string `json:"hash"`
	//ManufacturerId string `json:"manufacturererId"`
	//Size string `json:"size"`
	//prevVersionBlockHash string `json:"prevVersionBlockHash"`
	//prevArtefactBlockHash string `json:"prevArtefactBlockHash"`
	ArtefactType string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type Device struct {
	DeviceId string
	CurrentArtefactName string
	CurrentArtefactVerision string
}

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

	//marshal an emtpy array of strings to clear the index for the artefacts
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)
	err = stub.PutState(artefactIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)


	//create Artefact - done
	//delete Artefact - done
	//deploy Artefact (transfer to Device) Artefact Stores a List of Devices which installed the particular artefact
	//undeploy Artefact (delete Device from Artefakt) Removing an Device from an Artefacts Device List

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "delete" {
		res, err := t.Delete(stub, args)
		cleanTrades(stub)
		return res, err
	} else if function == "write" {
		return t.Write(stub, args)
	} else if function == "init_artefact" {
		return t.init_marble(stub, args)
	} else if function == "deploy_artefact" {
		res, err := t.deploy_artefact(stub, args)
		cleanTrades(stub)
		return res, err
	} else if function == "open_trade" {
		return t.open_trade(stub, args)
	} else if function == "perform_trade" {
		res, err := t.perform_trade(stub, args)
		cleanTrades(stub)
		return res, err
	} else if function == "remove_trade" {
		return t.remove_trade(stub, args)
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
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the artefact index
	artefactsAsBytes, err := stub.GetState(artefactIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get artefact index")
	}
	var artefactIndex []string
	json.Unmarshal(artefactsAsBytes, &artefactIndex)								//un stringify it aka JSON.parse()
	
	//remove artefact from index
	for i,val := range artefactIndex {
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct marble
			fmt.Println("found artefact")
			artefactIndex = append(artefactIndex[:i], artefactIndex[i+1:]...)			//remove it
			for x:= range artefactIndex {											//debug prints...
				fmt.Println(string(x) + " - " + artefactIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(artefactIndex)									//save new index
	err = stub.PutState(artefactIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Artefact - create a new artefact, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_artefact(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error


	if len(args) != 4 {
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

	version := args[0]
	name 	:= args[1]
	hash 	:= args[2]
	artefactType := args[3]
	timestamp := makeTimestamp()

	//check if already exists
	artefactAsBytes, err := stub.GetState(name + version)
	if err != nil {
		return nil, errors.New("Failed to get Artefact name + version")
	}
	res := Artefact{}
	json.Unmarshal(artefactAsBytes, &res)
	if res.Name == name && res.Version == version {
		fmt.Println("This Artefact arleady exists: " + name + version)
		fmt.Println(res);
		return nil, errors.New("This Artefact arleady exists")
	}

	//build the json string manually
	str := `{"artefactVersion": "` + version + `", + "artefactType": "` + artefactType + `", "artefactName": "` + name + `", "hash": "` + hash + `", "timestamp": "` + timestamp + `"}`
	err = stub.PutState(name + version, []byte(str))
	if err != nil {
		return nil, err
	}
		
	//get the artefacts index
	artefactsAsBytes, err := stub.GetState(artefactIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get artefact index")
	}
	var artefactIndex []string
	json.Unmarshal(artefactsAsBytes, &artefactIndex)							//un stringify it aka JSON.parse()
	
	//append
	artefactIndex = append(artefactIndex, name + version)									//add marble name to index list
	fmt.Println("! artefact index: ", artefactIndex)
	jsonAsBytes, _ := json.Marshal(artefactIndex)
	err = stub.PutState(artefactIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init artefact")
	return nil, nil
}

// ============================================================================================================================
// Init Device - create a new Device, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_device(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
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

	deviceId := args[0]
	currentArtefactVersion 	:= args[1]
	currentArtefactName 	:= args[2]
	timestamp := makeTimestamp()

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
	str := `{"deviceId": "` + deviceId + `", + "currentArtefactVersion": "` + currentArtefactVersion + `", "artefactName": "` + currentArtefactName + `", "timestamp": "` + timestamp + `"}`
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
	json.Unmarshal(devicesAsBytes, &deviceIndex)							//un stringify it aka JSON.parse()

	//append
	deviceIndex = append(deviceIndex, deviceId)									//add marble name to index list
	fmt.Println("! device index: ", deviceIndex)
	jsonAsBytes, _ := json.Marshal(deviceIndex)
	err = stub.PutState(deviceIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init device")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Marble
// ============================================================================================================================
func (t *SimpleChaincode) deploy_artefact(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error


	//deviceId requests an artefact. deviceId needs to be mapped on artefact UUID (name+version)

	//   0       1
	// "deviceId", "artefactName", "artefactVersion"
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	
	fmt.Println("- deploy")
	fmt.Println(args[0] + " - " + args[1])
	deviceAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get device")
	}

	artefactAsBytes, err := stub.GetState(args[1] + args[2])
	if err != nil {
		return nil, errors.New("Failed to get artefact")
	}

	res := Device{}
	json.Unmarshal(deviceAsBytes, &res)
	res.CurrentArtefactName = args[1]
	res.CurrentArtefactVerision = args[2]
	
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

