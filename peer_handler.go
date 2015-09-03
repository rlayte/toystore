package toystore

import (
	"encoding/gob"
	"net"
	"net/rpc"

	"github.com/rlayte/toystore/data"
)

// PeerHandler defines the methods that a node should expose to other
// nodes in the cluster.
type PeerHandler interface {
	Get(args *GetArgs, reply *GetReply) error
	Put(args *PutArgs, reply *PutReply) error
	CoordinateGet(args *GetArgs, reply *GetReply) error
	CoordinatePut(args *PutArgs, reply *PutReply) error
}

// serve starts listening for RPC calls, and creates a new thread for
// each incoming connection.
func serve(address string, rpcs *rpc.Server) {
	l, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			panic(err)
		}

		go rpcs.ServeConn(conn)
	}
}

// RpcHandler implements PeerHandler using Go's RPC package.
type RpcHandler struct {
	store *Toystore
}

// Get looks up and item from Toystore's underlying Store data.
func (r *RpcHandler) Get(args *GetArgs, reply *GetReply) error {
	reply.Value, reply.Ok = r.store.Data.Get(args.Key)
	return nil
}

// Put adds a value directly to Toystore's underlying Store data.
func (r *RpcHandler) Put(args *PutArgs, reply *PutReply) error {
	r.store.Data.Put(args.Value)
	reply.Ok = true
	return nil
}

// CoordinateGet kicks off the coordination process from a
// non-coordinator node.
func (r *RpcHandler) CoordinateGet(args *GetArgs, reply *GetReply) error {
	reply.Value, reply.Ok = r.store.CoordinateGet(args.Key)
	return nil
}

// CoordinatePut kicks off the coordination process from a
// non-coordinator node.
func (r *RpcHandler) CoordinatePut(args *PutArgs, reply *PutReply) error {
	reply.Ok = r.store.CoordinatePut(args.Value)
	return nil
}

// HintPut adds a new data hint to the node's HintedHandoff list.
func (r *RpcHandler) HintPut(args *HintArgs, reply *HintReply) error {
	r.store.Hints.Put(args.Data, args.Hint)
	reply.Ok = true
	return nil
}

// Transfer adds a set of data to the node.
func (r *RpcHandler) Transfer(args *TransferArgs, reply *TransferReply) error {
	ok := true

	// TODO: Should use Merge here to take the latest value.
	for _, item := range args.Data {
		r.store.Data.Put(item)
	}

	reply.Ok = ok
	return nil
}

// NewRpcHandler returns a new RpcHandler instance and starts serving requests.
func NewRpcHandler(store *Toystore) *RpcHandler {
	gob.Register(data.Data{})
	rpcs := rpc.NewServer()
	s := &RpcHandler{store}
	rpcs.Register(s)
	go serve(store.rpcAddress(), rpcs)

	return s
}
