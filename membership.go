package toystore

import (
	"log"

	"github.com/charlesetc/dive"
)

func (t *Toystore) Adjacent(address string) bool {
	return t.Ring.Adjacent([]byte(t.rpcAddress()), []byte(address))
}

func (t *Toystore) handleJoin(address string) {
	log.Printf("Toystore joined: %s\n", address)
	t.Ring.AddString(address)

	if t.Adjacent(address) {
		log.Println("Adjacent.")
		t.Transfer(address)
	}
}

func (t *Toystore) handleFail(address string) {
	log.Printf("Toystore left: %s\n", address)
	if address != t.rpcAddress() {
		t.Ring.RemoveString(address) // this is causing a problem
	}
}

func (t *Toystore) serveAsync() {
	for {
		select {
		case event := <-t.dive.Events:
			switch event.Kind {
			case dive.Join:
				address := event.Data.(ToystoreMetaData).RPCAddress // might not be rpc..
				t.handleJoin(address)
			case dive.Fail:
				address := event.Data.(ToystoreMetaData).RPCAddress
				t.handleFail(address)
			}
		case key := <-t.requestAddress:
			t.receiveAddress <- t.Ring.KeyAddress(key)
		}
	}
}
