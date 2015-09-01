package toystore

import "github.com/rlayte/toystore/data"

type ToystoreRPC struct {
	store *Toystore
}

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
