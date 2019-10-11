package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mataharimall/mesos-consul/consul"
	"github.com/mataharimall/mesos-consul/mesos"
)

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

	m := mesos.NewClient(c, *mesosPtr)
	m.OnEvent = func(e mesos.EventStatusUpdate) {
		fmt.Println(e.TimeStamp, e.TaskStatus, e.AppID, ",", e.TaskID, "(", e.Host, ":", e.Ports, ")")
	}

	apps, err := m.List()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %v apps\n", len(apps))
	for _, app := range apps {
		css := mesos2consul(app)
		for _, cs := range css {
			err := con.Register(cs)
			if err != nil {
				log.Printf("error registering %v (%v): %v\n", cs.Name, cs.ID, err)
			}
		}
	}

	if err := m.Subscribe(); err != nil {
		log.Fatal(err)
	}
}

func mesos2consul(a mesos.App) []consul.Service {
	ss := make([]consul.Service, 0)
	for _, t := range a.Tasks {
		if t.State != mesos.StateRunning {
			log.Println(t.State)
			continue
		}

		if len(t.Ports) == 0 {
			fmt.Println("No port:", a.ID, t.ID)
			continue
		}

		ss = append(ss, consul.Service{
			ID:      t.ID,
			Name:    strings.Replace(strings.Trim(a.ID, "/"), "/", ".", -1),
			Tags:    []string{"mesos"},
			Address: t.Host,
			Port:    t.Ports[0],
		})
	}
	return ss
}
