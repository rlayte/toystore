package toystore

import "github.com/rlayte/toystore/data"

// GetArgs is used to request data from other nodes.
type GetArgs struct {
	Key string
}

// GetReply is used to send data to other nodes.
type GetReply struct {
	Value *data.Data
	Ok    bool
}

// PutArgs is used to write data on other nodes.
type PutArgs struct {
	Value *data.Data
}

// PutReply is used to send write status to other nodes.
type PutReply struct {
	Ok bool
}

// HintArgs is used to store hinted data temporarily on other nodes.
type HintArgs struct {
	Data *data.Data
	// Address where the data should be stored.
	Hint string
}

// HintReply is used to return hint status to other nodes.
type HintReply struct {
	Ok bool
}

// TransferArgs is used to send chunks of data to other nodes.
type TransferArgs struct {
	Data []*data.Data
}

// TransferReply is used to send transfer status to other nodes.
type TransferReply struct {
	Ok bool
}
