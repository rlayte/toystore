package toystore

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func host() string {
	hosts := []string{
		"http://127.0.0.2:3000",
		"http://127.0.0.3:3000",
		"http://127.0.0.4:3000",
		"http://127.0.0.5:3000",
		"http://127.0.0.6:3000",
	}

	return hosts[rand.Intn(len(hosts))]
}

func TestBasicData(t *testing.T) {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("basic-%d", i)
		value := fmt.Sprintf("basic-value-%d", i)
		data := url.Values{"key": {key}, "value": {value}}

		_, err := http.PostForm(host(), data)

		if err != nil {
			t.Error(err)
		}
	}

	time.Sleep(time.Second)

	for i := 0; i < 100; i++ {
		h := host()
		key := fmt.Sprintf("basic-%d", i)
		value := fmt.Sprintf("basic-value-%d", i)
		resp, err := http.Get(h + "/" + key)
		defer resp.Body.Close()

		if err != nil {
			t.Error(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if string(body) != value {
			t.Errorf("%s: %s != %s", h, string(body), value)
		}
	}
}
