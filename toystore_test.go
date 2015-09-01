package toystore

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

var numTests int = 100
var cmds []*exec.Cmd = []*exec.Cmd{}
var m *sync.Mutex = &sync.Mutex{}
var hosts []string = []string{
	"127.0.0.2",
	"127.0.0.3",
	"127.0.0.4",
	"127.0.0.5",
	"127.0.0.6",
}

func host() string {
	return fmt.Sprintf("http://%s:3000", hosts[rand.Intn(len(hosts))])
}

func startNode(host string) {
	log.Println("Starting node", host)
	args := []string{"run", "api/http.go", host}
	cmd := exec.Command("go", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go cmd.Run()

	m.Lock()
	cmds = append(cmds, cmd)
	m.Unlock()
}

func startCluster() {
	log.Println("Starting cluster")

	startNode(hosts[0])

	time.Sleep(time.Second)

	for _, host := range hosts[1:] {
		startNode(host)
	}

	log.Println("Waiting for cluster")
	time.Sleep(time.Second * 10)
	log.Println("Cluster running")
}

func stopCluster() {
	for _, cmd := range cmds {
		log.Println("Killing", cmd.Process.Pid)
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15) // note the minus sign
		} else {
			log.Println("Failed to kill", cmd.Process.Pid)
		}
	}

}

func randomset(t *testing.T, i int) {
	key := fmt.Sprintf("basic-%d", i)
	value := fmt.Sprintf("basic-value-%d", i)
	h := host()
	data := url.Values{"key": {key}, "value": {value}}

	_, err := http.PostForm(h, data)

	if err != nil {
		t.Error(err)
	}
}

func randomget(t *testing.T, i int) {
	h := host()
	key := fmt.Sprintf("basic-%d", i)
	value := fmt.Sprintf("basic-value-%d", i)
	resp, err := http.Get(h + "/" + key)

	if err != nil {
		t.Fatal("Error", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if string(body) != value {
		t.Errorf("%s/%s %s != %s", h, key, string(body), value)
	}
}

func TestIntegration__BasicData(t *testing.T) {
	var i int

	startCluster()
	defer stopCluster()

	for i = 0; i < numTests; i++ {
		go randomset(t, i)
	}

	time.Sleep(time.Second)

	for i = 0; i < numTests; i++ {
		go randomget(t, i)
	}

	time.Sleep(time.Second)
}
