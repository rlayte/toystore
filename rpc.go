package toystore

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
