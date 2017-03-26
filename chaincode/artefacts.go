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
	} else if function == "set_user" {
		res, err := t.set_user(stub, args)
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
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	marbleAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res)										//un stringify it aka JSON.parse()
	res.User = args[1]														//change the user
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set user")
	return nil, nil
}

// ============================================================================================================================
// Open Trade - create an open trade for a marble you want with marbles you have 
// ============================================================================================================================
func (t *SimpleChaincode) open_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var will_size int
	var trade_away Description
	
	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}
	if len(args)%2 == 0{
		return nil, errors.New("Incorrect number of arguments. Expecting an odd number")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	open := AnOpenTrade{}
	open.User = args[0]
	open.Timestamp = makeTimestamp()											//use timestamp as an ID
	open.Want.Color = args[1]
	open.Want.Size =  size1
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	for i:=3; i < len(args); i++ {												//create and append each willing trade
		will_size, err = strconv.Atoi(args[i + 1])
		if err != nil {
			msg := "is not a numeric string " + args[i + 1]
			fmt.Println(msg)
			return nil, errors.New(msg)
		}
		
		trade_away = Description{}
		trade_away.Color = args[i]
		trade_away.Size =  will_size
		fmt.Println("! created trade_away: " + args[i])
		jsonAsBytes, _ = json.Marshal(trade_away)
		err = stub.PutState("_debug2", jsonAsBytes)
		
		open.Willing = append(open.Willing, trade_away)
		fmt.Println("! appended willing to open")
		i++;
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)										//un stringify it aka JSON.parse()
	
	trades.OpenTrades = append(trades.OpenTrades, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState(openTradesStr, jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}

// ============================================================================================================================
// Perform Trade - close an open trade and move ownership
// ============================================================================================================================
func (t *SimpleChaincode) perform_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0		1					2					3				4					5
	//[data.id, data.closer.user, data.closer.name, data.opener.user, data.opener.color, data.opener.size]
	if len(args) < 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}
	
	fmt.Println("- start close trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	size, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, errors.New("6th argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)															//un stringify it aka JSON.parse()
	
	for i := range trades.OpenTrades{																//look for the trade
		fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			
			
			marbleAsBytes, err := stub.GetState(args[2])
			if err != nil {
				return nil, errors.New("Failed to get thing")
			}
			closersMarble := Marble{}
			json.Unmarshal(marbleAsBytes, &closersMarble)											//un stringify it aka JSON.parse()
			
			//verify if marble meets trade requirements
			if closersMarble.Color != trades.OpenTrades[i].Want.Color || closersMarble.Size != trades.OpenTrades[i].Want.Size {
				msg := "marble in input does not meet trade requriements"
				fmt.Println(msg)
				return nil, errors.New(msg)
			}
			
			marble, e := findMarble4Trade(stub, trades.OpenTrades[i].User, args[4], size)			//find a marble that is suitable from opener
			if(e == nil){
				fmt.Println("! no errors, proceeding")

				t.set_user(stub, []string{args[2], trades.OpenTrades[i].User})						//change owner of selected marble, closer -> opener
				t.set_user(stub, []string{marble.Name, args[1]})									//change owner of selected marble, opener -> closer
			
				trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)		//remove trade
				jsonAsBytes, _ := json.Marshal(trades)
				err = stub.PutState(openTradesStr, jsonAsBytes)										//rewrite open orders
				if err != nil {
					return nil, err
				}
			}
		}
	}
	fmt.Println("- end close trade")
	return nil, nil
}

// ============================================================================================================================
// findMarble4Trade - look for a matching marble that this user owns and return it
// ============================================================================================================================
func findMarble4Trade(stub shim.ChaincodeStubInterface, user string, color string, size int )(m Marble, err error){
	var fail Marble;
	fmt.Println("- start find marble 4 trade")
	fmt.Println("looking for " + user + ", " + color + ", " + strconv.Itoa(size));

	//get the marble index
	marblesAsBytes, err := stub.GetState(artefactIndexStr)
	if err != nil {
		return fail, errors.New("Failed to get marble index")
	}
	var marbleIndex []string
	json.Unmarshal(marblesAsBytes, &marbleIndex)								//un stringify it aka JSON.parse()
	
	for i:= range marbleIndex{													//iter through all the marbles
		//fmt.Println("looking @ marble name: " + marbleIndex[i]);

		marbleAsBytes, err := stub.GetState(marbleIndex[i])						//grab this marble
		if err != nil {
			return fail, errors.New("Failed to get marble")
		}
		res := Marble{}
		json.Unmarshal(marbleAsBytes, &res)										//un stringify it aka JSON.parse()
		//fmt.Println("looking @ " + res.User + ", " + res.Color + ", " + strconv.Itoa(res.Size));
		
		//check for user && color && size
		if strings.ToLower(res.User) == strings.ToLower(user) && strings.ToLower(res.Color) == strings.ToLower(color) && res.Size == size{
			fmt.Println("found a marble: " + res.Name)
			fmt.Println("! end find marble 4 trade")
			return res, nil
		}
	}
	
	fmt.Println("- end find marble 4 trade - error")
	return fail, errors.New("Did not find marble to use in this trade")
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// ============================================================================================================================
// Remove Open Trade - close an open trade
// ============================================================================================================================
func (t *SimpleChaincode) remove_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	
	//	0
	//[data.id]
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return nil, errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																//un stringify it aka JSON.parse()
	
	for i := range trades.OpenTrades{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)				//remove this trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(openTradesStr, jsonAsBytes)												//rewrite open orders
			if err != nil {
				return nil, err
			}
			break
		}
	}
	
	fmt.Println("- end remove trade")
	return nil, nil
}

// ============================================================================================================================
// Clean Up Open Trades - make sure open trades are still possible, remove choices that are no longer possible, remove trades that have no valid choices
// ============================================================================================================================
func cleanTrades(stub shim.ChaincodeStubInterface)(err error){
	var didWork = false
	fmt.Println("- start clean trades")
	
	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return errors.New("Failed to get opentrades")
	}
	var trades AllTrades
	json.Unmarshal(tradesAsBytes, &trades)																		//un stringify it aka JSON.parse()
	
	fmt.Println("# trades " + strconv.Itoa(len(trades.OpenTrades)))
	for i:=0; i<len(trades.OpenTrades); {																		//iter over all the known open trades
		fmt.Println(strconv.Itoa(i) + ": looking at trade " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10))
		
		fmt.Println("# options " + strconv.Itoa(len(trades.OpenTrades[i].Willing)))
		for x:=0; x<len(trades.OpenTrades[i].Willing); {														//find a marble that is suitable
			fmt.Println("! on next option " + strconv.Itoa(i) + ":" + strconv.Itoa(x))
			_, e := findMarble4Trade(stub, trades.OpenTrades[i].User, trades.OpenTrades[i].Willing[x].Color, trades.OpenTrades[i].Willing[x].Size)
			if(e != nil){
				fmt.Println("! errors with this option, removing option")
				didWork = true
				trades.OpenTrades[i].Willing = append(trades.OpenTrades[i].Willing[:x], trades.OpenTrades[i].Willing[x+1:]...)	//remove this option
				x--;
			}else{
				fmt.Println("! this option is fine")
			}
			
			x++
			fmt.Println("! x:" + strconv.Itoa(x))
			if x >= len(trades.OpenTrades[i].Willing) {														//things might have shifted, recalcuate
				break
			}
		}
		
		if len(trades.OpenTrades[i].Willing) == 0 {
			fmt.Println("! no more options for this trade, removing trade")
			didWork = true
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)					//remove this trade
			i--;
		}
		
		i++
		fmt.Println("! i:" + strconv.Itoa(i))
		if i >= len(trades.OpenTrades) {																	//things might have shifted, recalcuate
			break
		}
	}

	if(didWork){
		fmt.Println("! saving open trade changes")
		jsonAsBytes, _ := json.Marshal(trades)
		err = stub.PutState(openTradesStr, jsonAsBytes)														//rewrite open orders
		if err != nil {
			return err
		}
	}else{
		fmt.Println("! all open trades are fine")
	}

	fmt.Println("- end clean trades")
	return nil
}