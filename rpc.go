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
	Data *data.Data
	Hint string
}

type HintReply struct {
	Ok bool
}

type TransferArgs struct {
	Data []*data.Data
}

type TransferReply struct {
	Ok bool
}
