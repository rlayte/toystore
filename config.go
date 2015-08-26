package toystore

type Config struct {
	ReplicationLevel int
	W                int
	R                int
	ClientPort       int
	RPCPort          int
	GossipPort       int
	Host             string
	SeedAddress      string
	Store            Store
}
