package membership

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/require"
)

func TestCluster(t *testing.T) {
	addr := "127.0.0.1:9000"
	config := Config{
		NodeName: "test",
		BindAddr: addr,
		Tags: map[string]string{
			"rpc_addr": addr,
			"purpose":  "test",
		},
		JoinAddrs: strings.Split("127.0.0.1:8991,127.0.0.1:8992", ","),
	}

	member, err := NewMembership(config, &handler{}, nil)
	if err != nil {
		log.Fatal("error:", err)
	}

	// SetTags replace tag via Update event
	member.Serf.SetTags(map[string]string{"rpc_addr": addr})

	// UserEvent: broadcast user event
	member.Serf.UserEvent("hello", []byte("world"), true)
	member.Serf.UserEvent("hello", []byte("world"), true)
	member.Serf.UserEvent("hello1", []byte("world1"), true)
	member.Serf.UserEvent("hello2", []byte("world2"), true)
	member.Serf.UserEvent("hello", []byte("world"), false)
	member.Serf.UserEvent("hello", []byte("world"), false)

	// check API
	t.Logf("LocalMember: %#v", member.Serf.LocalMember())
	t.Logf("Members: %#v", member.Serf.Members())
	t.Logf("State: %#v", member.Serf.State())
	t.Logf("Status: %#v", member.Serf.Stats())
	t.Logf("NumNodes: %#v", member.Serf.NumNodes())

	// Query:
	resp, err := member.Serf.Query("hello", []byte("world"), nil)
	require.NoError(t, err)
	r, ok := <-resp.ResponseCh()
	if ok {
		t.Logf("receive response from %s: %s", r.From, string(r.Payload))
	}
}

func TestMembership(t *testing.T) {
	m, handler := setupMember(t, nil)
	m, _ = setupMember(t, m)
	m, _ = setupMember(t, m)

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			0 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.NoError(t, m[2].Leave())

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			serf.StatusLeft == m[0].Members()[2].Status &&
			1 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.Equal(t, fmt.Sprintf("%d", 2), <-handler.leaves)
}

func setupMember(t *testing.T, members []*Membership) (
	[]*Membership, *handler,
) {
	id := len(members)
	port := 10000 + id
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", port)
	tags := map[string]string{
		"rpc_addr": addr,
	}
	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}
	h := &handler{}
	if len(members) == 0 {
		h.joins = make(chan map[string]string, 3)
		h.leaves = make(chan string, 3)
	} else {
		c.JoinAddrs = []string{
			members[0].BindAddr,
		}
	}
	m, err := NewMembership(c, h, nil)
	require.NoError(t, err)
	members = append(members, m)
	return members, h
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(member serf.Member) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   member.Name,
			"addr": member.Tags["rpc_addr"],
		}
	}
	return nil
}

func (h *handler) Leave(member serf.Member) error {
	if h.leaves != nil {
		h.leaves <- member.Name
	}
	return nil
}

func (h *handler) Update(member serf.Member) error {
	return nil
}

func (h *handler) Reap(member serf.Member) error {
	return nil
}

func (h *handler) User(ue serf.UserEvent) error {
	return nil
}

func (h *handler) Query(query *serf.Query) error {
	return nil
}
