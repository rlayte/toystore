package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rlayte/toystore"
	"github.com/rlayte/toystore/adapters/redis"
)

func main() {
	var seed string
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [port]", os.Args[0])
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic(err)
	}

	if port != 3000 {
		seed = ":3010"
	}

	t := toystore.New(port, redis.New("localhost:6379"), seed, toystore.ToystoreMetaData{RPCAddress: ":3020"})
	t.Serve()
}
