package toystore

import (
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

	for {
		conn, err := l.Accept()

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
	reply.Value, reply.Ok = r.store.Data.Get(args.Key)
	return nil
}

func (r *RpcHandler) Put(args *PutArgs, reply *PutReply) error {
	r.store.Data.Put(args.Value)
	reply.Ok = true
	return nil
}

func (r *RpcHandler) CoordinateGet(args *GetArgs, reply *GetReply) error {
	reply.Value, reply.Ok = r.store.CoordinateGet(args.Key)
	return nil
}

func (r *RpcHandler) CoordinatePut(args *PutArgs, reply *PutReply) error {
	reply.Ok = r.store.CoordinatePut(args.Value)
	return nil
}

func (r *RpcHandler) HintPut(args *HintArgs, reply *HintReply) error {
	r.store.Hints.Put(args.Data, args.Hint)
	reply.Ok = true
	return nil
}

func (r *RpcHandler) Transfer(args *TransferArgs, reply *TransferReply) error {
	ok := true

	for _, item := range args.Data {
		r.store.Data.Put(item)
	}

	reply.Ok = ok
	return nil
}

func NewRpcHandler(store *Toystore) *RpcHandler {
	rpcs := rpc.NewServer()
	s := &RpcHandler{store}
	rpcs.Register(s)
	go serve(store.rpcAddress(), rpcs)

	return s
}
