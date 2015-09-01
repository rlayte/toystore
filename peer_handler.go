package toystore

import (
	"log"
	"net"
	"net/rpc"
)

type PeerHandler interface {
	Get(args *GetArgs, reply *GetReply) error
	Put(args *PutArgs, reply *PutReply) error
	CoordinateGet(args *GetArgs, reply *GetReply) error
	CoordinatePut(args *PutArgs, reply *PutReply) error
}

func serve(address string, rpcs *rpc.Server) {
	l, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	log.Println("Serving RPC requests", address)

	for {
		log.Println(address, "Waiting for RPC connection")
		conn, err := l.Accept()
		log.Println(address, "Received connection")

		if err != nil {
			panic(err)
		}

		go rpcs.ServeConn(conn)
	}
}

type RpcHandler struct {
	store *Toystore
}

func (r *RpcHandler) Get(args *GetArgs, reply *GetReply) error {
	log.Println("RPC Get")
	reply.Value, reply.Ok = r.store.Data.Get(args.Key)
	return nil
}

func (r *RpcHandler) Put(args *PutArgs, reply *PutReply) error {
	log.Println("RPC Put")
	r.store.Data.Put(args.Key, args.Value)
	reply.Ok = true
	return nil
}

func (r *RpcHandler) CoordinateGet(args *GetArgs, reply *GetReply) error {
	log.Println("RPC CoordinateGet")
	reply.Value, reply.Ok = r.store.CoordinateGet(args.Key)
	return nil
}

func (r *RpcHandler) CoordinatePut(args *PutArgs, reply *PutReply) error {
	log.Println("RPC CoordinatePut")
	reply.Ok = r.store.CoordinatePut(args.Key, args.Value)
	return nil
}

func NewRpcHandler(store *Toystore) *RpcHandler {
	rpcs := rpc.NewServer()
	s := &RpcHandler{store}
	rpcs.Register(s)
	go serve(store.rpcAddress(), rpcs)

	return s
}
