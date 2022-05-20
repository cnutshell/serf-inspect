package main

import (
	"flag"
	"log"
	"strings"

	"github.com/cnutshell/membership"
	"github.com/hashicorp/serf/serf"
)

var (
	address   string
	nodename  string
	bootstrap string
)

func init() {
	flag.StringVar(&address, "address", "", "IP address")
	flag.StringVar(&nodename, "nodename", "", "node name")
	flag.StringVar(&bootstrap, "bootstrap", "", "command seperated know address")
}

func main() {
	flag.Parse()

	if address == "" {
		log.Fatal("address required")
	}
	if nodename == "" {
		log.Fatal("nodename required")
	}
	if bootstrap == "" {
		log.Fatal("bootstrap required")
	}

	config := membership.Config{
		NodeName: nodename,
		BindAddr: address,
		Tags: map[string]string{
			"rpc_addr": address,
		},
		JoinAddrs: strings.Split(bootstrap, ","),
	}

	_, err := membership.NewMembership(config, &handler{}, nil)
	if err != nil {
		log.Fatal("error:", err)
	}

	ch := make(chan bool)
	<-ch
}

type handler struct {
}

func (h *handler) Join(member serf.Member) error {
	log.Printf("====>Join event with member: %#v", member)
	return nil
}
func (h *handler) Leave(member serf.Member) error {
	log.Printf("====>Leave event with member: %#v", member)
	return nil
}
func (h *handler) Update(member serf.Member) error {
	log.Printf("====>Update event with member: %#v", member)
	return nil
}
func (h *handler) Reap(member serf.Member) error {
	log.Printf("====>Reap event with member: %#v", member)
	return nil
}
func (h *handler) User(ue serf.UserEvent) error {
	log.Printf("====>User event with member: %#v", ue)
	return nil
}
func (h *handler) Query(query *serf.Query) error {
	log.Printf("====>Query event with member: %#v", query)
	query.Respond([]byte("world"))
	return nil
}
