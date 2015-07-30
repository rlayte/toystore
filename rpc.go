package toystore

import (
	"net"
	"net/rpc"
	"time"
)

type ToystoreRPC struct {
	store *Toystore
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Value string
	Ok    bool
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
	Ok bool
}

func (t *ToystoreRPC) Get(args *GetArgs, reply *GetReply) error {
	reply.Value, reply.Ok = t.store.data.Get(args.Key)
	return nil
}

func (t *ToystoreRPC) Put(args *PutArgs, reply *PutReply) error {
	t.store.data.Put(args.Key, args.Value)
	reply.Ok = true
	return nil
}

func ServeRPC(store *Toystore) {
	rpcs := rpc.NewServer()
	s := &ToystoreRPC{store}
	rpcs.Register(s)

	l, err := net.Listen("tcp", store.rpcAddress())

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

func GetCall(address string, key string) (string, bool) {
	args := &GetArgs{key}
	reply := &GetReply{}

	call(address, "ToystoreRPC.Get", args, reply)

	return reply.Value, reply.Ok
}

func PutCall(address string, key string, value string) bool {
	args := &PutArgs{key, value}
	reply := &PutReply{}

	call(address, "ToystoreRPC.Put", args, reply)

	return reply.Ok
}