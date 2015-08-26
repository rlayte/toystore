package toystore

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
)

func TestBasicData(t *testing.T) {
	hosts := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:3003",
		"http://localhost:3004",
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("basic-%d", i)
		value := fmt.Sprintf("basic-value-%d", i)
		host := hosts[rand.Intn(len(hosts))]
		data := url.Values{"key": {key}, "value": {value}}

		log.Printf("Test: %s - %s/%s", host, key, value)

		_, err := http.PostForm(host, data)

		if err != nil {
			t.Error(err)
		}

		resp, err := http.Get(host + "/" + key)
		defer resp.Body.Close()

		if err != nil {
			t.Error(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if string(body) != value {
			t.Errorf("%s should equal %s", string(body), value)
		}
	}
}
