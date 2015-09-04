package toystore

import (
	"log"
	"net/rpc"
	"time"

	"github.com/rlayte/toystore/data"
)

// PeerClient defines the possible interactions between nodes in the cluster.
// Should be implemented with a specific transport client.
type PeerClient interface {
	Get(address string, key string) (value *data.Data, status bool)
	Put(address string, value *data.Data) (status bool)
	CoordinateGet(address string, key string) (value *data.Data, status bool)
	CoordinatePut(address string, value *data.Data) (status bool)
	HintPut(address string, hint string, value *data.Data) (status bool)
	Transfer(address string, data []*data.Data) (status bool)
}

// dial attempts to connect to a specified RPC server.
// It will retry every 1/3 seconds if connection fails.
// If it can't connect within 1 second it aborts and panics.
func dial(address string) *rpc.Client {
	var err error
	var conn *rpc.Client
	success := make(chan *rpc.Client)

	go func() {
		conn, err = rpc.Dial("tcp", address)

		for err != nil {
			conn, err = rpc.Dial("tcp", address)
			time.Sleep(time.Second / 3)
		}

		success <- conn
	}()

	select {
	case conn := <-success:
		return conn
	case <-time.After(time.Second):
		log.Printf("Failed to dial address: %s", address)
		return nil
	}
}

// call attempts to make an RPC.
// If the connection fails it returns false, otherwise true.
func call(address string, method string, args interface{}, reply interface{}) bool {
	if address == "" {
		return false
	}

	conn := dial(address)

	if conn == nil {
		return false
	}

	err := conn.Call(method, args, reply)
	conn.Close()

	if err != nil {
		return false
	}

	return true
}

// RpcClient implements PeerClient using Go's RPC package.
type RpcClient struct {
}

// Get makes an RPC to the address to find the specified key and returns
// the value and an existence bool.
func (r *RpcClient) Get(address string, key string) (*data.Data, bool) {
	log.Printf("GET request to %s for %s", address, key)
	args := &GetArgs{key}
	reply := &GetReply{}

	call(address, "RpcHandler.Get", args, reply)

	return reply.Value, reply.Ok
}

// Put makes an RPC to the address to add the Data value and returns a boolean
// representing the status of this operation.
func (r *RpcClient) Put(address string, value *data.Data) bool {
	log.Printf("PUT request to %s for %v", address, value)
	args := &PutArgs{value}
	reply := &PutReply{}

	call(address, "RpcHandler.Put", args, reply)

	return reply.Ok
}

// CoordinateGet forwards the key to the coordinating node so it can organize
// the Get operation.
func (r *RpcClient) CoordinateGet(address string, key string) (*data.Data, bool) {
	log.Printf("Forwarding GET request to %s for %s", address, key)
	args := &GetArgs{key}
	reply := &GetReply{}

	call(address, "RpcHandler.CoordinateGet", args, reply)

	return reply.Value, reply.Ok
}

// CoordinatePut forwards the Data value to the coordinating node so it can organize
// the Put operation.
func (r *RpcClient) CoordinatePut(address string, value *data.Data) bool {
	log.Printf("Forwarding PUT request to coordinator %s for %s", address, value.Key)

	args := &PutArgs{value}
	reply := &PutReply{}

	call(address, "RpcHandler.CoordinatePut", args, reply)

	return reply.Ok
}

// HintPut makes an RPC to add hint data to the specified node.
func (r *RpcClient) HintPut(address string, hint string, data *data.Data) bool {
	log.Printf("Sending hint to %s for %s (%s)", address, hint, data)

	args := &HintArgs{data, hint}
	reply := &HintReply{}

	call(address, "RpcHandler.HintPut", args, reply)

	return reply.Ok
}

// Transfer makes an RPC call to send a set of keys to the specified address.
func (r *RpcClient) Transfer(address string, data []*data.Data) bool {
	log.Printf("Transferring data to %s - %v", address, data)
	args := &TransferArgs{data}
	reply := &TransferReply{}

	call(address, "RpcHandler.Transfer", args, reply)

	return reply.Ok
}

// NewRpcClient returns a new RpcClient instance.
func NewRpcClient() *RpcClient {
	return &RpcClient{}
}
