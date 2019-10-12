package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mataharimall/mesos-consul/consul"
	"github.com/mataharimall/mesos-consul/mesos"
)

// TODO: Marathon label into consul tag
// TODO: Handle mesos.Subscribe() error
func main() {
	consulPtr := flag.String("consul", "http://127.0.0.1:8500", "Consul address")
	mesosPtr := flag.String("mesos", "http://127.0.0.1:8080", "Mesos address")
	helpPtr := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *helpPtr {
		fmt.Println("Usage: mesos-consul [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -consul=addr   Consul address (default: http://127.0.0.1:8500)")
		fmt.Println("  -mesos=addr    Mesos address (default: http://127.0.0.1:8080)")
		os.Exit(0)
	}

	c := &http.Client{}

	con := consul.NewClient(c, *consulPtr)

	m, err := mesos.NewClient(c, *mesosPtr)
	if err != nil {
		log.Fatal("Unable to connect to Marathon:", err)
	}

	m.OnEvent = func(t mesos.Task) {
		switch t.State {
		case mesos.StateStaging, mesos.StateStarting:
			// Do nothing

		case mesos.StateRunning:
			cs := mesosTaskToConsulService(t)
			err := con.Register(cs)
			if err != nil {
				log.Printf("error registering %v (%v): %v\n", cs.Name, cs.ID, err)
			}

		case mesos.StateFinished, mesos.StateFailed, mesos.StateKilling, mesos.StateKilled, mesos.StateLost:
			err := con.DeRegister(t.ID)
			if err != nil {
				log.Println("unable to deregister " + t.ID)
			}
		}
	}

	// Deregister non existing service in consul list
	mTasks := m.Tasks()
	log.Printf("%v services found in Marathon\n", len(mTasks))
	r := 0 // services that already exists on consul
	l, err := con.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range l {
		if _, ok := mTasks[s]; ok {
			r++
			delete(mTasks, s)
			continue
		}

		err := con.DeRegister(s)
		if err != nil {
			log.Println("unable to deregister " + s)
		}
	}
	if len(l) > 0 {
		log.Printf("%v services exists on Consul, %v are cleaned up, %v stays\n", len(l), len(l)-r, r)
	}

	// Register service not yet exists in consul
	for _, task := range mTasks {
		cs := mesosTaskToConsulService(task)
		err := con.Register(cs)
		if err != nil {
			log.Printf("error registering %v (%v): %v\n", cs.Name, cs.ID, err)
		}
	}
	if len(mTasks) > 0 {
		log.Printf("%v services registered on Consul\n", len(mTasks))
	}

	if err := m.Subscribe(); err != nil {
		log.Fatal(err)
	}
}

func mesosTaskToConsulService(task mesos.Task) consul.Service {
	return consul.Service{
		ID:      task.ID,
		Name:    strings.Replace(strings.Trim(task.AppID, "/"), "/", ".", -1),
		Tags:    []string{"mesos"},
		Address: task.Host,
		Port:    task.Ports[0],
		Check: consul.Check{
			HTTP:     "http://" + task.Host + ":" + strconv.Itoa(task.Ports[task.HealthCheck.PortIndex]) + task.HealthCheck.Path,
			Interval: strconv.Itoa(task.HealthCheck.Interval) + "s",
			Timeout:  strconv.Itoa(task.HealthCheck.TimeOut) + "s",
		},
	}
}
