package toystore

import "github.com/rlayte/toystore/data"

type GetArgs struct {
	Key string
}

type GetReply struct {
	Value *data.Data
	Ok    bool
}

type PutArgs struct {
	Value *data.Data
}

type PutReply struct {
	Ok bool
}

type HintArgs struct {
	Key   string
	Value string
	Hint  string
}

type HintReply struct {
	Ok bool
}

type TransferArgs struct {
	Data []*Data
}

type TransferReply struct {
	Ok bool
}
