package toystore

import (
	"log"
	"net/rpc"
	"time"

	"github.com/rlayte/toystore/data"
)

type PeerClient interface {
	Get(address string, key string) (value string, status bool)
	Put(address string, key string, value string) (status bool)
	CoordinateGet(address string, key string) (value string, status bool)
	CoordinatePut(address string, key string, value string) (status bool)
	Transfer(address string, data []*data.Data) (status bool)
}

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
		panic(err)
		return nil
	}
}

func call(address string, method string, args interface{}, reply interface{}) bool {
	conn := dial(address)

	err := conn.Call(method, args, reply)
	conn.Close()

	if err != nil {
		return false
	}

	return true
}

type RPCPeerClient struct {
}

func (r *RPCPeerClient) Get(address string, key string) (string, bool) {
	args := &GetArgs{key}
	reply := &GetReply{}

	call(address, "RpcHandler.Get", args, reply)

	return reply.Value, reply.Ok
}

func (r *RPCPeerClient) Put(address string, key string, value string) bool {
	args := &PutArgs{key, value}
	reply := &PutReply{}

	call(address, "RpcHandler.Put", args, reply)

	return reply.Ok
}

func (r *RPCPeerClient) CoordinateGet(address string, key string) (string, bool) {
	log.Printf("Forwarding GET request to %s for %s", address, key)
	args := &GetArgs{key}
	reply := &GetReply{}

	call(address, "RpcHandler.CoordinateGet", args, reply)

	return reply.Value, reply.Ok
}

func (r *RPCPeerClient) CoordinatePut(address string, key string, value string) bool {
	log.Printf("Forwarding PUT request to coordinator %s for %s", address, key)

	args := &PutArgs{key, value}
	reply := &PutReply{}

	call(address, "RpcHandler.CoordinatePut", args, reply)

	return reply.Ok
}

func (r *RPCPeerClient) Transfer(address string, data []*data.Data) bool {
	log.Printf("Transferring data to %s - %v", address, data)
	return true
}

func NewRpcClient() *RPCPeerClient {
	return &RPCPeerClient{}
}
