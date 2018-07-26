/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func hashSHA256File(filePath string) (string, error) {
	var hashValue string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error1\n")
		fmt.Printf("%s\n", filePath)
		fmt.Println(err)
		return hashValue, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Printf("error2")
		return hashValue, err
	}
	hashInBytes := hash.Sum(nil)
	hashValue = hex.EncodeToString(hashInBytes)
	return hashValue, nil

}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init......")
	_, args := stub.GetFunctionAndParameters()
	var A, B string       // Entities
	var Aval, Bval string // Asset holdings
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	A = args[0]
	Aval = "0"
	/*
		if err != nil {
			return shim.Error("Expecting integer value for asset holding")
		}
	*/
	B = args[1]
	Bval = "0"
	/*
		if err != nil {
			return shim.Error("Expecting integer value for asset holding")
		}
	*/
	fmt.Printf("Ahashval = %s, Bhashval = %s\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(Aval))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(Bval))
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Initiation Succeed!")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke......")
	function, args := stub.GetFunctionAndParameters()
	if function == "transfer" {
		// Make payment of X units from A to B
		return t.transfer(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	} else if function == "upload" {
		return t.upload(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\" \"upload\"")
}

func (t *SimpleChaincode) upload(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string        // Entities
	var filePath string // FilePath
	var Aval string
	var X string // Transaction value
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	A = args[0]
	filePath = args[1]
	fmt.Printf("%s is uploading %s ......\n", A, filePath)
	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval = string(Avalbytes)
	/*
		Bvalbytes, err := stub.GetState(B)
		if err != nil {
			return shim.Error("Failed to get state")
		}
		if Bvalbytes == nil {
			return shim.Error("Entity not found")
		}
		Bval, _ = strconv.Atoi(string(Bvalbytes))
	*/
	// Perform the execution
	X, err = hashSHA256File(filePath)
	if err != nil {
		return shim.Error("Invalid File Path, Please Check it.")
	}

	if X == Aval {
		return shim.Error("File have already existed in the system. Refuse to upload it again.")
	}
	Aval = X
	/*
		Aval = Aval - X
		Bval = Bval + X
	*/
	fmt.Printf("File successfully uploaded!\n")
	fmt.Printf(" %s, sha256 value: %s\n ", filePath, X)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(Aval))
	if err != nil {
		return shim.Error(err.Error())
	}
	/*
		err = stub.PutState(B, []byte(Bval))
		if err != nil {
			return shim.Error(err.Error())
		}
	*/
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string // Entities
	var filePath string
	var Aval, Bval string // Asset holdings
	var X string          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]
	filePath = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval = string(Avalbytes)

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval = string(Bvalbytes)

	// Perform the execution
	X, err = hashSHA256File(filePath)
	if err != nil {
		return shim.Error("Invalid File Path, Please Check it.")
	}

	if X != Aval {
		return shim.Error("File has not been uploaded to the system. Please Upload it first!")
	}
	//Aval = Aval - X
	Bval = X
	fmt.Printf("File has been successfully transferred from %s to %s\n", A, B)
	fmt.Printf("Bval = %s\n", Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(Aval))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(Bval))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Hash\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
