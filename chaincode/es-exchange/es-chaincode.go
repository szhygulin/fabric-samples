package main

import (
    "fmt"
    "strconv"
    "github.com/hyperledger/fabric/common/util"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type Exchange struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Exchange) Init(stub shim.ChaincodeStubInterface) peer.Response {
    return shim.Success(nil)
}

// Invoke
func (am *Exchange) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
    actionName, args := stub.GetFunctionAndParameters()
// Params - usd sender address, usd amount, ene sender address, ene amount
    if actionName == "exchange" {
        usdCCGetArgs := util.ToChaincodeArgs("get", args[0])
        usdResponse := stub.InvokeChaincode("USDAsset", usdCCGetArgs, "myc")
        if usdResponse.Status != shim.OK {
            return shim.Error(usdResponse.Message)
        }
        usd_value, err1 := strconv.ParseInt(args[1], 10, 64)
        if err1 != nil {
            return shim.Error(fmt.Sprintf("Failed to convert USD amount to be sent: %s with error: %s", args[1], err1))
        }
        usd_sender_value, err1 := strconv.ParseInt(string(usdResponse.Payload), 10, 64)
        if err1 != nil {
            return shim.Error(fmt.Sprintf("Failed to convert asset: %s with error: %s", usdResponse.Payload, err1))
        }
        eneCCGetArgs := util.ToChaincodeArgs("get", args[2])
        eneResponse := stub.InvokeChaincode("EnergyAsset", eneCCGetArgs, "myc")
        if eneResponse.Status != shim.OK {
            return shim.Error(eneResponse.Message)
        }
        ene_value, err1 := strconv.ParseInt(args[3], 10, 64) 
        if err1 != nil {
            return shim.Error(fmt.Sprintf("Failed to convert Ene amount to be sent: %s with error: %s", args[3], err1))
        }
        ene_sender_value, err1 := strconv.ParseInt(string(eneResponse.Payload), 10, 64) 
        if err1 != nil {
            return shim.Error(fmt.Sprintf("Failed to convert asset: %s with error: %s", eneResponse.Payload, err1))
        }
        if usd_sender_value < usd_value {
            return shim.Error(fmt.Sprintf("Failed USD transfer - sender has %s, but inteded to send %s", usdResponse.Payload, args[1]))
        }
        if ene_sender_value < ene_value {
            return shim.Error(fmt.Sprintf("Failed energy transfer - sender has %s, but inteded to send %s", eneResponse.Payload, args[3]))
        }
        fmt.Printf("thus far successful, %s, %s", args[1], args[3])
        usdSendGetArgs := util.ToChaincodeArgs("send", args[0], args[2], args[1])
        usdSendResponse := stub.InvokeChaincode("USDAsset", usdSendGetArgs, "myc")
        if usdSendResponse.Status != shim.OK {
            return shim.Error(usdSendResponse.Message)
        }
        eneSendGetArgs := util.ArrayToChaincodeArgs("send", args[2], args[0], args[3])
        eneSendResponse := stub.InvokeChaincode("EnergyAsset", eneSendGetArgs, "myc")
        if eneSendResponse.Status != shim.OK {
            return shim.Error(eneSendResponse.Message)
        }

        return shim.Success(nil)
    }

    // NOTE: This is an example, hence assuming only valid call is to call another chaincode
    return shim.Error(fmt.Sprintf("[ERROR] No <%s> action defined", actionName))
}

// main function starts up the chaincode in the container during instantiate
func main() {
    if err := shim.Start(new(Exchange)); err != nil {
            fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
    }
}
