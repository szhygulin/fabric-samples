package main

import (
    "fmt"
    "strconv"
    "time"
    "encoding/json"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type USDAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *USDAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
    // Get the args from the transaction proposal
    args := stub.GetStringArgs()
    if len(args) != 0 {
            return shim.Error("Incorrect arguments. Expecting nothing")
    }

    // Set up any variables or assets here by calling stub.PutState()

    // We store the key and the value on the ledger
    //err := stub.PutState(args[0], []byte(args[1]))
    //if err != nil {
    //        return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
    //}
    return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *USDAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
    // Extract the function and args from the transaction proposal
    fn, args := stub.GetFunctionAndParameters()

    var result string
    var err error
    if fn == "set" {
            result, err = set(stub, args)
    } else if fn == "send" {
            result, err = send(stub, args)
    } else if fn == "keys" {
            return t.keys(stub, args)
    } else { // assume 'get' even if fn is nil
            result, err = get(stub, args)
    }
    if err != nil {
            return shim.Error(err.Error())
    }

    // Return the result as success payload
    return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
    if len(args) != 2 {
            return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
    }

    err := stub.PutState(args[0], []byte(args[1]))
    if err != nil {
            return "", fmt.Errorf("Failed to set asset: %s", args[0])
    }
    return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
    if len(args) != 1 {
            return "", fmt.Errorf("Incorrect arguments. Expecting a key")
    }

    value, err := stub.GetState(args[0])
    if err != nil {
            return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
    }
    if value == nil {
            return "", fmt.Errorf("Asset not found: %s", args[0])
    }
    return string(value), nil
}

// Send transfer the asset from sender key to recipient key on the ledger. If the sender key does not exists,
// it will rise an error. If recipient key does not exist, it will create one.
func send(stub shim.ChaincodeStubInterface, args []string) (string, error) {
    if len(args) != 3 {
            return "", fmt.Errorf("Incorrect arguments. Expecting a sender key, recipient key, and a value")
    }
    fmt.Printf("Args received: %s, %s, %s", args[0], args[1], args[2])
    sender_value, err1 := get(stub, []string{args[0]})
    if err1 != nil {
            return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err1)
    }
    sender_value_int, err := strconv.ParseInt(sender_value, 10, 64)
    if err != nil {
            return "", fmt.Errorf("Failed to convert asset to int64: %s with error: %s", sender_value, err)
    }
    amount_int, err := strconv.ParseInt(args[2], 10, 64)
    if err != nil {
            return "", fmt.Errorf("Failed to convert amount to int64: %s with error: %s", args[2], err)
    }
    if sender_value_int < amount_int {
            return "", fmt.Errorf("Sender's asset value on %s is less than intended to transfer", args[0])
    }
    
    new_sender_balance_int := sender_value_int - amount_int
    new_sender_balance := strconv.FormatInt(new_sender_balance_int, 10)
    result, err2 := set(stub, []string{args[0], new_sender_balance})

    recipient_value, err3 := get(stub, []string{args[1]})
    if err3 == nil {
            recipient_value_int, err := strconv.ParseInt(recipient_value, 10, 64)
            if err != nil {
                    return "", fmt.Errorf("Failed to convert asset to int64: %s with error: %s", recipient_value, err)
            }
            new_recipient_balance_int := recipient_value_int + amount_int
            new_recipient_balance := strconv.FormatInt(new_recipient_balance_int,10)
            result, err2 = set(stub, []string{args[1], new_recipient_balance})
    } else {
            result, err2 = set(stub, []string{args[1], args[2]})
    }

    return result, err2
}

func (t *USDAsset) keys(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) > 0 {
		return shim.Error("keys operation must not include any arguments")
	}
	startKey := ""
	endKey := ""

	//sleep needed to test peer's timeout behavior when using iterators
	stime := 100
	//if len(args) > 2 {
	//	stime, _ = strconv.Atoi(args[2])
	//}

	keysIter, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
	}
	defer keysIter.Close()

	var keys []string
	for keysIter.HasNext() {
		//if sleeptime is specied, take a nap
		if stime > 0 {
			time.Sleep(time.Duration(stime) * time.Millisecond)
		}

		response, iterErr := keysIter.Next()
		if iterErr != nil {
			return shim.Error(fmt.Sprintf("keys operation failed. Error accessing state: %s", err))
		}
		keys = append(keys, response.Key)
	}

	for key, value := range keys {
		fmt.Printf("key %d contains %s\n", key, value)
	}

	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		return shim.Error(fmt.Sprintf("keys operation failed. Error marshaling JSON: %s", err))
	}

	return shim.Success(jsonKeys)
}


// main function starts up the chaincode in the container during instantiate
func main() {
    if err := shim.Start(new(USDAsset)); err != nil {
            fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
    }
}
