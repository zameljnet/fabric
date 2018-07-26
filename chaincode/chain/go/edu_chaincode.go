/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'
// peer chaincode invoke -C myc -n mycc -c '{"Args":["delete","51114214"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support marbles02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":["data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"data.docType\",\"data.owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index
//

// Index for docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
//
// Index definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.size\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Luxurioust/excelize"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type student struct {
	ObjectType                      string  `json:"docType"`     //docType is used to distinguish the various types of objects in state database
	StudentCode                     string  `json:"studentCode"` //the fieldtags are needed to keep case from bouncing around
	Gender                          string  `json:"gender"`
	CensusRegister                  string  `json:"censusRegister"`
	SeniorHighSchool                string  `json:"seniorHighSchool"`
	SchoolCode                      int     `json:"schoolCode"`
	CollegeEntranceExaminationScore float64 `json:"collegeEntranceExaminationScore"`
	Chinese                         float64 `json:"chinese"`
	Maths                           float64 `json:"maths"`
	English                         float64 `json:"english"`
	ComprehensiveTest               float64 `json:"comprehensiveTest"`
}

type marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function + "\n")

	// Handle different functions
	if function == "initStudent" { //create a new marble
		return t.initStudent(stub, args)
	} else if function == "updateStudent" { //change owner of a specific marble
		return t.updateStudent(stub, args)
	} else if function == "transferMarblesBasedOnColor" { //transfer all marbles of a certain color
		return t.transferMarblesBasedOnColor(stub, args)
	} else if function == "delete" { //delete a marble
		return t.delete(stub, args)
	} else if function == "readStudent" { //read a marble
		return t.readStudent(stub, args)
	} else if function == "queryMarblesByOwner" { //find marbles for owner X using rich query
		return t.queryMarblesByOwner(stub, args)
	} else if function == "queryMarbles" { //find marbles based on an ad hoc rich query
		return t.queryMarbles(stub, args)
	} else if function == "getHistoryForStudent" { //get history of values for a marble
		return t.getHistoryForStudent(stub, args)
	} else if function == "getMarblesByRange" { //get marbles based on range query
		return t.getMarblesByRange(stub, args)
	} else if function == "InitWithData" { //get marbles based on range query
		return t.InitWithData(stub)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

//   0       	1       2    		 3				4		5		6		7	8	9
// "51114214", "女", "京籍", "首都师范大学附属密云中学","28101","536.5","106.5","100"，"97"，"233"
//StudentCode                     string    `json:"studentCode"` //the fieldtags are needed to keep case from bouncing around
//Gender                          string `json:"gender"`
//CensusRegister                  string `json:"censusRegister"`
//SeniorHighSchool                string `json:"seniorHighSchool"`
//SchoolCode                      int    `json:"schoolCode"`
//CollegeEntranceExaminationScore int    `json:"collegeEntranceExaminationScore"`
//Chinese                         int    `json:"chinese"`
//Maths                           int    `json:"maths"`
//English                         int    `json:"english"`
//ComprehensiveTest               int    `json:"comprehensiveTest"`
func (t *SimpleChaincode) InitWithData(stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	xlsx, err := excelize.OpenFile("excel-1.xlsx")
	if err != nil {
		return shim.Error("Open File Error")
		//fmt.Println(err)
		//os.Exit(1)
	}
	rows := xlsx.GetRows("Sheet1")
	//fmt.Print(rows[0][0])
	for index, row := range rows {
		if index == 0 {
			continue
		}
		studentCode := row[0]
		gender := row[1]
		censusRegister := row[2]
		seniorHighSchool := row[3]
		schoolCode, err := strconv.Atoi(row[4])
		if err != nil {
			return shim.Error("5th argument must be a numeric string!!!")
		}
		collegeEntranceExaminationScore, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			return shim.Error("6th argument must be a numeric string")
		}
		chinese, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			return shim.Error("7th argument must be a numeric string")
		}
		maths, err := strconv.ParseFloat(row[7], 64)
		if err != nil {
			return shim.Error("8th argument must be a numeric string")
		}
		english, err := strconv.ParseFloat(row[8], 64)
		if err != nil {
			return shim.Error("9th argument must be a numeric string")
		}
		comprehensiveTest, err := strconv.ParseFloat(row[9], 64)
		if err != nil {
			return shim.Error("10th argument must be a numeric string")
		}

		fmt.Println("- start init student ", "StudentCode:", studentCode, "Gender: "+gender, "CensusRegister: "+censusRegister, "SeniorHighSchool: "+seniorHighSchool, "SchoolCode", schoolCode, "CollegeEntranceExaminationScore:", collegeEntranceExaminationScore, "Chinese:", chinese, "English:", english, "Maths:", maths, "ComprehensiveTest:", comprehensiveTest)
		studentAsBytes, err := stub.GetState(studentCode)
		if err != nil {
			fmt.Println("Failed to get student: " + studentCode + err.Error())
			continue
		} else if studentAsBytes != nil {
			fmt.Println("This student already exists: " + studentCode)
			continue
		}
		objectType := "student"
		student := &student{objectType, studentCode, gender, censusRegister, seniorHighSchool, schoolCode, collegeEntranceExaminationScore, chinese, maths, english, comprehensiveTest}
		studentJSONasBytes, err := json.Marshal(student)
		if err != nil {
			fmt.Println("Marshal Error" + err.Error())
		}
		err = stub.PutState(studentCode, studentJSONasBytes)
		if err != nil {
			fmt.Println("putState Error" + err.Error())
		}
		fmt.Println("- end init student (success) ", "StudentCode:", studentCode+"\n")
	}
	fmt.Println("- end InitWithData (success)" + "\n")
	return shim.Success(nil)
}

// ============================================================
// initMarble - create a new marble, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       	1       2    		 3				4		5		6		7	8	9
	// "51114214", "女", "京籍", "首都师范大学附属密云中学","28101","536.5","106.5","100"，"97"，"233"
	//StudentCode                     string    `json:"studentCode"` //the fieldtags are needed to keep case from bouncing around
	//Gender                          string `json:"gender"`
	//CensusRegister                  string `json:"censusRegister"`
	//SeniorHighSchool                string `json:"seniorHighSchool"`
	//SchoolCode                      int    `json:"schoolCode"`
	//CollegeEntranceExaminationScore int    `json:"collegeEntranceExaminationScore"`
	//Chinese                         int    `json:"chinese"`
	//Maths                           int    `json:"maths"`
	//English                         int    `json:"english"`
	//ComprehensiveTest               int    `json:"comprehensiveTest"`
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}
	if len(args[9]) <= 0 {
		return shim.Error("10th argument must be a non-empty string")
	}

	studentCode := args[0]
	gender := args[1]
	censusRegister := args[2]
	seniorHighSchool := args[3]
	schoolCode, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("5th argument must be a numeric string!!!")
	}
	collegeEntranceExaminationScore, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error("6th argument must be a numeric string")
	}
	chinese, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return shim.Error("7th argument must be a numeric string")
	}
	maths, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error("8th argument must be a numeric string")
	}
	english, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error("9th argument must be a numeric string")
	}
	comprehensiveTest, err := strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error("10th argument must be a numeric string")
	}

	fmt.Println("- start init student ", "StudentCode:", studentCode, "Gender: "+gender, "CensusRegister: "+censusRegister, "SeniorHighSchool: "+seniorHighSchool, "SchoolCode", schoolCode, "CollegeEntranceExaminationScore:", collegeEntranceExaminationScore, "Chinese:", chinese, "English:", english, "Maths:", maths, "ComprehensiveTest:", comprehensiveTest)

	// ==== Check if student already exists ====
	studentAsBytes, err := stub.GetState(studentCode)
	if err != nil {
		return shim.Error("Failed to get student: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("This student already exists: " + studentCode)
		return shim.Error("This student already exists: " + studentCode)
	}
	objectType := "student"
	student := &student{objectType, studentCode, gender, censusRegister, seniorHighSchool, schoolCode, collegeEntranceExaminationScore, chinese, maths, english, comprehensiveTest}
	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save student to state ===
	err = stub.PutState(studentCode, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	/*
		//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
		//  An 'index' is a normal key/value entry in state.
		//  The key is a composite key, with the elements that you want to range query on listed first.
		//  In our case, the composite key is based on indexName~color~name.
		//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
		indexName := "color~name"
		colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(colorNameIndexKey, value)
	*/

	// ==== Student saved and indexed. Return success ====

	fmt.Println("- end init student (success)" + "\n")
	return shim.Success(nil)
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (t *SimpleChaincode) readStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var code, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting code of the student to query")
	}

	code = args[0]
	valAsbytes, err := stub.GetState(code) //get the student from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + code + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Student does not exist: " + code + "\"}"
		return shim.Error(jsonResp)
	}

	studentToTransfer := student{}
	err = json.Unmarshal(valAsbytes, &studentToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\n \"StudentCode\":")
	buffer.WriteString("\"")
	buffer.WriteString(studentToTransfer.StudentCode)
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"Gender\":")
	buffer.WriteString("\"")
	buffer.WriteString(studentToTransfer.Gender)
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"CensusRegister\":")
	buffer.WriteString("\"")
	buffer.WriteString(studentToTransfer.CensusRegister)
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"SeniorHighSchool\":")
	buffer.WriteString("\"")
	buffer.WriteString(studentToTransfer.SeniorHighSchool)
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"SchoolCode\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatInt(int64(studentToTransfer.SchoolCode), 10))
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"CollegeEntranceExaminationScore\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatFloat(studentToTransfer.CollegeEntranceExaminationScore, 'f', -1, 64))
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"Chinese\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatFloat(studentToTransfer.Chinese, 'f', -1, 64))
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"Maths\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatFloat(studentToTransfer.Maths, 'f', -1, 64))
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"English\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatFloat(studentToTransfer.English, 'f', -1, 64))
	buffer.WriteString("\"")
	buffer.WriteString(",\n \"ComprehensiveTest\":")
	buffer.WriteString("\"")
	buffer.WriteString(strconv.FormatFloat(studentToTransfer.ComprehensiveTest, 'f', -1, 64))
	buffer.WriteString("\"\n")
	buffer.WriteString("}")
	fmt.Printf("- readStudent returning:\n%s\n", buffer.String())

	fmt.Println("- end readStudent (success)" + "\n")
	//return shim.Success(buffer.Bytes())

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a student key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var studentJSON student
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	studentCode := args[0]

	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetState(studentCode) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + studentCode + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Student does not exist: " + studentCode + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &studentJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + studentCode + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(studentCode) //remove the marble from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	/*
			// maintain the index
			indexName := "color~name"
			colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleJSON.Color, marbleJSON.Name})
			if err != nil {
				return shim.Error(err.Error())
			}


		//  Delete index entry to state.
		err = stub.DelState(colorNameIndexKey)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}
	*/
	fmt.Println("- end delete (success)" + "\n")
	return shim.Success(nil)
}

// ===========================================================
// transfer a marble by setting a new owner name on the marble
// ===========================================================
func (t *SimpleChaincode) updateStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// ！=nil且不相等 则赋值
	//   0       	1       2    		 3				4		5		6		7	8	9
	// "51114214", "女", "京籍", "首都师范大学附属密云中学","28101","536.5","106.5","100"，"97"，"233"
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}
	if len(args[9]) <= 0 {
		return shim.Error("10th argument must be a non-empty string")
	}

	studentCode := args[0]
	gender := args[1]
	censusRegister := args[2]
	seniorHighSchool := args[3]
	schoolCode, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("5th argument must be a numeric string!!!")
	}
	collegeEntranceExaminationScore, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error("6th argument must be a numeric string")
	}
	chinese, err := strconv.ParseFloat(args[6], 64)
	if err != nil {
		return shim.Error("7th argument must be a numeric string")
	}
	maths, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error("8th argument must be a numeric string")
	}
	english, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error("9th argument must be a numeric string")
	}
	comprehensiveTest, err := strconv.ParseFloat(args[9], 64)
	if err != nil {
		return shim.Error("10th argument must be a numeric string")
	}
	if chinese+maths+english+comprehensiveTest != collegeEntranceExaminationScore {
		return shim.Error("Score error ")
	}

	// ==== Check if student already exists ====
	studentAsBytes, err := stub.GetState(studentCode)
	if err != nil {
		return shim.Error("Failed to get student: " + err.Error())
	} else if studentAsBytes == nil {
		fmt.Println("This student does not exists: " + studentCode)
		return shim.Error("This student does not exists: " + studentCode)
	}

	fmt.Println("- start updateStudent ", studentCode)

	studentCodeAsBytes, err := stub.GetState(studentCode)
	if err != nil {
		return shim.Error("Failed to get studentCode:" + err.Error())
	} else if studentCodeAsBytes == nil {
		return shim.Error("studentCode does not exist")
	}

	studentCodeToTransfer := student{}
	err = json.Unmarshal(studentCodeAsBytes, &studentCodeToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	studentCodeToTransfer.Gender = gender
	studentCodeToTransfer.CensusRegister = censusRegister
	studentCodeToTransfer.SeniorHighSchool = seniorHighSchool
	studentCodeToTransfer.SchoolCode = schoolCode
	studentCodeToTransfer.CollegeEntranceExaminationScore = collegeEntranceExaminationScore
	studentCodeToTransfer.Chinese = chinese
	studentCodeToTransfer.Maths = maths
	studentCodeToTransfer.English = english
	studentCodeToTransfer.ComprehensiveTest = comprehensiveTest

	studentJSONasBytes, _ := json.Marshal(studentCodeToTransfer)
	err = stub.PutState(studentCode, studentJSONasBytes) //rewrite the student
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end updateStudent (success)" + "\n")
	return shim.Success(nil)
}

// ===========================================================================================
// getMarblesByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getMarblesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferMarblesBasedOnColor will transfer marbles of a given color to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferMarblesBasedOnColor(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "color", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	color := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)

	// Query the color~name index by color
	// This will execute a key range query on all keys starting with 'color'
	coloredMarbleResultsIterator, err := stub.GetStateByPartialCompositeKey("color~name", []string{color})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredMarbleResultsIterator.Close()

	// Iterate through result set and for each marble found, transfer to newOwner
	var i int
	for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		responseRange, err := coloredMarbleResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedColor := compositeKeyParts[0]
		returnedMarbleName := compositeKeyParts[1]
		fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedMarbleName)

		// Now call the transfer function for the found marble.
		// Re-use the same function that is used to transfer individual marbles
		response := t.updateStudent(stub, []string{returnedMarbleName, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s marbles to %s", i, color, newOwner)
	fmt.Println("- end transferMarblesBasedOnColor: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryMarblesByOwner queries for marbles based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarblesByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarbles(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	studentCode := args[0]

	fmt.Printf("- start getHistoryForStudent: %s\n", studentCode)

	resultsIterator, err := stub.GetHistoryForKey(studentCode)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(" {\n  \"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)

		buffer.WriteString("\"")

		buffer.WriteString(",\n  \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(",\n  \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(",\n  \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"\n")

		buffer.WriteString(" }\n")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]\n")

	fmt.Printf("- getHistoryForStudent returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
