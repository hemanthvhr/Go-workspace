package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"fmt"
	"encoding/json"
	"encoding/csv"

	"github.com/davecgh/go-spew/spew"
)

// Each block in the blockchain
type Block struct {
	Index     int
	Timestamp string
	Hash      string
	nonce	  int
	Data 	  Transaction
	PrevHash  string
}

// Block represents the transaction added by the taxi
type Transaction struct {
	NodeType  	int
	Index     	int 	`index of the transaction`
	Time 		string
	Latitude	float64
	Longitude	float64
	NodeId		int 	`ID of the taxi`
}

// Block represents the vote given to some transaction
type Vote struct {
	NodeType  	int
	Index     	int 	`index of the transaction to which we are voting`
	Decision	bool
	Weight		int
}

// Block represents confirmation of the added block by the miner
type Confirmation struct {
	NodeType  	int
	Index     	int 	`index of the transaction which is confirmed`
	NodeId		int 	`ID of the miner who confirmed or disconfirmed it`
	Decision	bool
	Nonce 		int 	`The nonce value computed by the miners`
}


// Node Properties and types
var nodeClass map[string]bool

// The csv file for the taxi data
const records_file = "landmark/sample_taxi_data.txt"

// The Blockchain
var Blockchain []Block

// The Last unconfirmed transaction, presently only supports one transaction
var pendingTrans Transaction
var pending bool

// Network properties
var networkPorts []int
var networkIP []string
var network, networkIn []net.Conn
var networkSize int
var seed int

// Node Properties
var myClass string
var myId int
var voteValue int

// The TCP server of the nodes
var server net.Listener

// for calculating the hash value of a block
func calculateHash(block Block) string {
	data, _ := json.Marshal(block.Data)
	record := string(block.Index) + block.Timestamp + strconv.Itoa(block.nonce) + string(data) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// This sets up the different node classes available.
func setNodeClasses() {
	nodeClass = make(map[string]bool)
	nodeClass["Spectator"] = true	// formal voting rights (vote value = 0)
	nodeClass["Taxi"] = true	// spectator + transaction making capabilities
	nodeClass["Regulator"] = true	// spectator + skewed voting rights
	nodeClass["Safety_Inspector"] = true // spectator + skewed voting rights can also add invalid transactions
	nodeClass["Miner"] = true	// spectator + Miner
	nodeClass["Player"] = true	// spectator + voting rights
}

// This function takes the id of the node and returns the port number.
func encodeId(id int) int {
	return id+(seed%4000)+2000
}

// This function takes the port number and returns the decoded node ID. 
func decodePort(port int) int {
	return port-(seed%4000)-2000
}

// Find the port no from net.Conn object
func getPort(conn net.Conn) int {
	// The local address of the process initiating the connection
	localAddr := conn.LocalAddr().String()
	fmt.Println("The addr string is : " + localAddr)
	var port string
	for i:=len(localAddr)-1; i>=0; i--	{
		if localAddr[i]==':' {
			port = localAddr[i+1:]
			break
		}
	}
	pno, err := strconv.Atoi(port)
	if err!=nil {
		fmt.Println("Unknown Addr string")
		return -1
	}
	return pno
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, Data Transaction, nonce int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.nonce = nonce
	newBlock.Data = Data
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// make sure the chain we're checking is longer than the current blockchain
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

// Prints the block in readable fashion
func printBlock(block Block) {
	fmt.Printf("Block - %v, Transaction num - %v, lat - %v, lon - %v\n",
				 block.Index, block.Data.Index, block.Data.Latitude, block.Data.Longitude)
}

// Prints the whole blockchain
func printBlockchain(blockchain []Block) {
	fmt.Printf("\n\n")
	for _, v := range blockchain {
		printBlock(v)
	}
	fmt.Printf("\n\n")
}

// Handles the connection from other nodes in the network
func handleSetup() {
	for {
		conn, err := server.Accept()
		if err!=nil {
			if conn!=nil {
				conn.Close()
			}
			continue
		}
		networkIn = append(networkIn, conn)
		fmt.Println("oh..")
	}
}

// Process the received data and take appropriate action
func receiveData(data string) {
	fmt.Println("Something")
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(data), &dat); err!=nil {
		fmt.Println("In Valid Data received")
		return
	}
	fmt.Println("switching")
	switch dat["NodeType"].(int) {
		case 1 : 
			// This is a transaction
			new_transaction := Transaction{}
			json.Unmarshal([]byte(data), &new_transaction)
			// If there is a pending transaction dont accept another
			if pending {
				return
			}
			pendingTrans = new_transaction
			pending = true
			// Add it to the blockchain
			new_block, _ := generateBlock(Blockchain[len(Blockchain)-1], new_transaction, 0)
			if isBlockValid(Blockchain[len(Blockchain)-1], new_block) {
				Blockchain = append(Blockchain, new_block)
			}
			pending = false

			fmt.Println(new_transaction)

			// Printing the Blockchain
			printBlockchain(Blockchain)

		case 2 : 
			// This is a vote
			// Every one except miners will ignore this
		case 3 : 
			// This is a confirmation
		default : 
			fmt.Println("Noisy Data Received")
			return
	}
}

// Broadcast some message to all nodes available
func broadCast(msg string, senderId int) {
	for k,conn := range network {
		fmt.Println("k - "+strconv.Itoa(k) + ", ")
		if (k+1)==senderId || conn==nil {
			continue
		}
		fmt.Println("Addr - "+conn.RemoteAddr().String())
		x, err := io.WriteString(conn, msg)
		if err!=nil {
			fmt.Println("Error -" + err.Error()+" - "+strconv.Itoa(x))
		}
		//fmt.Printf("Message is %v\n",msg)
	}
}

// It performs the opearations of a taxi node
// Referred - https://stackoverflow.com/a/24999351
func taxiNode() {

	f, err := os.Open(records_file)
	if err!=nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	count := 1

    for {
        row, err := csvr.Read()
        if err != nil {
            if err == io.EOF {
                err = nil
            }
            log.Fatal(err)
        }

        t := Transaction{}
        if t.Latitude, err = strconv.ParseFloat(row[3], 64); err != nil {
            log.Fatal(err)
        }
        if t.Longitude, err = strconv.ParseFloat(row[2], 64); err != nil {
            log.Fatal(err)
        }
        t.Time = row[0]+":"+row[1]
        t.NodeType = 1
        t.Index = count
        t.NodeId = myId

        // Broad cast this block
        msg, _ := json.Marshal(t)
        fmt.Println(string(msg))
        go broadCast(string(msg), 0)

        // Wait for some time before next Iteration
        time.Sleep(5*time.Second)
        count++
    }

}


// Discover all the nodes in the discovery and the ID self
func networkDiscovery() {

	n := networkSize
	networkPorts = make([]int, n)
	networkIP = make([]string, n)
	network = make([]net.Conn, n)

	for i:=0 ;i<n; i++ {
		networkPorts[i] = encodeId(i+1)
		networkIP[i] = "localhost"
		network[i] = nil
		fmt.Printf("Port[%v] - %v\n", i+1, networkPorts[i])
	}
}

func main() {

	// initial setup
	pending = false

	// Implement the class system
	argsW := os.Args
	if len(argsW)!=4 {
		log.Fatal("Insufficient arguments")
	}
	myId, _= strconv.Atoi(argsW[1])
	networkSize, _ = strconv.Atoi(argsW[2])
	seed, _ = strconv.Atoi(argsW[3])

	fmt.Printf("Node number - %v online\n", myId)

	// Discover the network
	networkDiscovery()

	// create genesis block
	t := time.Now()
	firstTrans := Transaction{1, 0, "", 0.0, 0.0, 0}
	genesisBlock := Block{0, t.String(), "", 0, firstTrans, ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	// start TCP server and putting it in listening state
	some_conn, err := net.Listen("tcp", ":"+strconv.Itoa(networkPorts[myId-1]))
	if err != nil {
		log.Fatal(err)
		fmt.Println("listening")
	}
	server = some_conn
	defer server.Close()

	// It accepts the connections from other nodes and adds it to the network
	go handleSetup()

	// Establishing connections with the remaining network
	for i:=0; i<networkSize; i++ {

		// Each connection setup is run in a go routine
		go func(id int) {
			for {
				conn, err := net.Dial("tcp", networkIP[id] + ":" + strconv.Itoa(networkPorts[id]))
				if err != nil {
					if conn!= nil {
						conn.Close()
					}
					continue
				}

				fmt.Printf("Connected to Node - %v, with port - %v\n", id+1, networkPorts[id])
				network[id] = conn
				return
			}
		}(i)
	}

	// Wait until all the neighbours are connected
	for {
		flag := true
		for i:=1; i<=networkSize; i++ {
			if network[i-1]==nil {
				flag = false
				//fmt.Println("id - " + strconv.Itoa(i))
			}
		}
		if flag {
			break
		}
		time.Sleep(4*time.Second)
 	}

	fmt.Println("Connections Established")

	if myId==1 {
		go taxiNode()
	}

	// Blockchain stuff
	// Loop infintely and check status of each connection

	for {
		for i:=1; i<=networkSize; i++ {
			if networkIn[i-1]==nil {
				continue
			}
			dataRecv, err := bufio.NewReader(networkIn[i-1]).ReadString('\n')
			fmt.Println("Not ok")
            if err != nil {
                fmt.Println(err)
                continue
            }

            fmt.Println("Dest Addr - "+ networkIn[i-1].RemoteAddr().String())

            // If the connection is closed on the other end, the len(dataRecv) will be 0 bytes,
            // else the EOF(or empty string) will be printed when as long as the connection is idle.
            if dataRecv == "" {
            	continue
            }
            
            // Else handle the received data
            go receiveData(dataRecv)
        }
	}

}
