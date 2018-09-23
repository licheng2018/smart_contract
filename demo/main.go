// Smart contract used in Hyperledger Fabric
// FileName: main.go
// Copyright (C) 2018 Licheng Zheng <lichengz@usc.edu or musicsir@outlook.com>
//
// Distributed under terms of the GPL license.

package main

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func ToChaincodergs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

type Chaincode struct {
}

//{"Args":["attr", "name"]}'
func (t *Chaincode) attr(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("parametes's number is wrong")
	}
	fmt.Println("get attr: ", args[0])
	value, ok, err := cid.GetAttributeValue(stub, args[0])
	if err != nil {
		return shim.Error("get attr error: " + err.Error())
	}

	if ok == false {
		value = "not found"
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return shim.Error("json marshal error: " + err.Error())
	}
	return shim.Success(bytes)
}

//{"Args":["creator2"]}'
func (t *Chaincode) creator2(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var cinfo struct {
		ID   string
		ORG  string
		CERT *x509.Certificate
	}

	fmt.Println("creator2: ", args)

	id, err := cid.GetID(stub)
	if err != nil {
		return shim.Error("getid error: " + err.Error())
	}

	id_readable, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return shim.Error("base64 decode error: " + err.Error())
	}
	cinfo.ID = string(id_readable)

	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error("getmspid error: " + err.Error())
	}
	cinfo.ORG = mspid

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return shim.Error("getX509Cert error: " + err.Error())
	}
	cinfo.CERT = cert

	bytes, err := json.Marshal(cinfo)
	if err != nil {
		return shim.Error("json marshal error: " + err.Error())
	}
	return shim.Success(bytes)
}

//{"Args":["creator"]}'
func (t *Chaincode) creator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("creator: ", args)
	bytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("get creator error: " + err.Error())
	}

	// TODO: status: 500, message: unmarshal creator error: invalid character '\x19'
	//looking for beginning of value, bytes:
	/*
		var creator msp.SerializedIdentity
		if err := json.Unmarshal(bytes, &creator); err != nil {
			return shim.Error("unmarshal creator error: " + err.Error())
		}
	*/
	return shim.Success(bytes)
}

//{"Args":["call","chaincode","method"...]}'
func (t *Chaincode) call(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("call: ", args)
	sub_args := args[1:]
	return stub.InvokeChaincode(args[0], ToChaincodergs(sub_args...), stub.GetChannelID())
}

//{"Args":["append","key", ...]}'
func (t *Chaincode) append(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	key := args[0]
	value := args[1:]
	var data []string

	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("query " + key + " fail: " + err.Error())
	}

	if bytes != nil {
		if err := json.Unmarshal(bytes, &data); err != nil {
			return shim.Error(err.Error())
		}
	}

	data = append(data, value...)
	new_bytes, err := json.Marshal(data)
	if err != nil {
	}

	if err := stub.PutState(key, new_bytes); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//{"Args":["query_chaincode","chaincode","key"]}'
func (t *Chaincode) query_chaincode(stub shim.ChaincodeStubInterface, chaincode, key string) pb.Response {
	fmt.Printf("query %s in %s\n", key, chaincode)
	return stub.InvokeChaincode(chaincode, ToChaincodergs("query", key), stub.GetChannelID())
}

//{"Args":["write_chaincode","chaincode","key","value"]}'
func (t *Chaincode) write_chaincode(stub shim.ChaincodeStubInterface, chaincode, key, value string) pb.Response {
	fmt.Printf("write %s to %s, value is %s\n", key, chaincode, value)
	return stub.InvokeChaincode(chaincode, ToChaincodergs("write", key, value), stub.GetChannelID())
}

//{"Args":["query","key"]}'
func (t *Chaincode) query(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("query %s\n", key)
	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}
	return shim.Success(bytes)
}

//{"Args":["history","key"]}'
func (t *Chaincode) history(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("history %s\n", key)
	iter, err := stub.GetHistoryForKey(key)
	defer iter.Close()
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}

	values := make(map[string]string)

	for iter.HasNext() {
		fmt.Printf("next\n")
		if kv, err := iter.Next(); err == nil {
			fmt.Printf("value: %s\n", kv.Value)
			values[kv.TxId] = string(kv.Value)
		}
		if err != nil {
			return shim.Error("iterator history fail: " + err.Error())
		}
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		return shim.Error("json marshal fail: " + err.Error())
	}

	return shim.Success(bytes)
}

//{"Args":["write","key","value"]}'
func (t *Chaincode) write(stub shim.ChaincodeStubInterface, key, value string) pb.Response {
	fmt.Printf("write %s, value is %s\n", key, value)
	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error("write fail " + err.Error())
	}
	return shim.Success(nil)
}

//{"Args":["init"]}
func (t *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode Chaincode")
	return shim.Success(nil)
}

func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	// return the user's information
	case "creator":
		return t.creator(stub, args)
	
	case "creator2":
		return t.creator2(stub, args)
	
	case "call":
		return t.call(stub, args)
	//append key
	case "append":
		return t.append(stub, args)
	//read the value
	case "attr":
		return t.attr(stub, args)
	//read current key
	case "query":
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.query(stub, args[0])
	//read history
	case "history":
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.history(stub, args[0])
	//create key
	case "write": //写入
		if len(args) != 2 {
			return shim.Error("parametes's number is wrong")
		}
		return t.write(stub, args[0], args[1])
	//
	case "query_chaincode":
		if len(args) != 2 {
			return shim.Error("parametes's number is wrong")
		}
		return t.query_chaincode(stub, args[0], args[1])
	//
	case "write_chaincode":
		if len(args) != 3 {
			return shim.Error("parametes's number is wrong")
		}
		return t.write_chaincode(stub, args[0], args[1], args[2])
	default:
		return shim.Error("Invalid invoke function name.")
	}
}

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting Chaincode chaincode: %s", err)
	}
}
