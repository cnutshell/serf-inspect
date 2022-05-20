package membership

import (
	"net"

	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Handler interface {
	Join(serf.Member) error
	Leave(serf.Member) error
	Update(serf.Member) error
	Reap(serf.Member) error
	User(serf.UserEvent) error
	Query(*serf.Query) error
}

type Membership struct {
	Config
	handler Handler
	Serf    *serf.Serf
	eventCh chan serf.Event
	logger  *zap.Logger
}

type Config struct {
	NodeName  string
	BindAddr  string
	Tags      map[string]string
	JoinAddrs []string
}

func NewMembership(config Config, handler Handler, logger *zap.Logger) (*Membership, error) {
	if logger == nil {
		logger = zap.L().Named("membership")
	}

	c := &Membership{
		handler: handler,
		Config:  config,
		logger:  logger,
		eventCh: make(chan serf.Event),
	}

	err := c.setupSerf()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// public
func (m *Membership) Members() []serf.Member {
	return m.Serf.Members()
}

func (m *Membership) Leave() error {
	return m.Serf.Leave()
}

// private
func (m *Membership) setupSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}

	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	config.EventCh = m.eventCh
	config.Tags = m.Tags
	config.NodeName = m.Config.NodeName

	m.Serf, err = serf.Create(config)
	if err != nil {
		return err
	}

	go m.eventHandler()

	if m.JoinAddrs != nil {
		_, err = m.Serf.Join(m.JoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Membership) eventHandler() {
	for e := range m.eventCh {
		switch e.EventType() {
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberUpdate:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleUpdate(member)
			}
		case serf.EventMemberReap:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleReap(member)
			}
		case serf.EventUser:
			m.handleUser(e.(serf.UserEvent))
		case serf.EventQuery:
			m.handleQuery(e.(*serf.Query))
		default:
			panic("unknown serf event type")
		}
	}
}

func (m *Membership) handleUpdate(member serf.Member) {
	if err := m.handler.Update(member); err != nil {
		m.logError(err, "failed to update", member.Name)
	}
}

// TODO
func (m *Membership) handleReap(member serf.Member) {
	if err := m.handler.Reap(member); err != nil {
		m.logError(err, "failed to reap", member.Name)
	}
}

// TODO
func (m *Membership) handleUser(ue serf.UserEvent) {
	if err := m.handler.User(ue); err != nil {
		m.logError(err, "failed to user", ue.Name)
	}
}

// TODO
func (m *Membership) handleQuery(query *serf.Query) {
	if err := m.handler.Query(query); err != nil {
		m.logError(err, "failed to query", query.Name)
	}
}

func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(member); err != nil {
		m.logError(err, "failed to join", member.Name)
	}
}

func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(member); err != nil {
		m.logError(err, "failed to leave", member.Name)
	}
}

func (m *Membership) isLocal(member serf.Member) bool {
	return m.Serf.LocalMember().Name == member.Name
}

func (m *Membership) logError(err error, msg string, name string) {
	m.logger.Error(
		msg,
		zap.Error(err),
		zap.String("name", name),
	)
}
