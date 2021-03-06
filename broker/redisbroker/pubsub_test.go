package redisbroker

import (
	"sync"
	"testing"
	"time"

	"github.com/mna/juggler/message"
	"github.com/mna/redisc/redistest"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	cmd, port := redistest.StartServer(t, nil, "")
	defer cmd.Process.Kill()

	pool := redistest.NewPool(t, ":"+port)
	brk := &Broker{
		Pool:    pool,
		Dial:    pool.Dial,
		LogFunc: logIfVerbose,
	}

	// list results on this conn UUID
	psc, err := brk.NewPubSubConn()
	require.NoError(t, err, "get PubSub connection")

	// keep track of received events
	wg := sync.WaitGroup{}
	wg.Add(1)
	var uuids []uuid.UUID
	go func() {
		defer wg.Done()
		for ep := range psc.Events() {
			uuids = append(uuids, ep.MsgUUID)
		}
	}()

	// subscribe to some channels
	require.NoError(t, psc.Subscribe("a", false), "Subscribe a")
	require.NoError(t, psc.Subscribe("b", false), "Subscribe b")

	cases := []struct {
		ch   string
		pp   *message.PubPayload
		exp  bool
		unsb string
	}{
		{"a", &message.PubPayload{MsgUUID: uuid.NewRandom()}, true, ""},
		{"b", &message.PubPayload{MsgUUID: uuid.NewRandom()}, true, ""},
		{"c", &message.PubPayload{MsgUUID: uuid.NewRandom()}, false, "a"},
		{"a", &message.PubPayload{MsgUUID: uuid.NewRandom()}, false, ""},
		{"b", &message.PubPayload{MsgUUID: uuid.NewRandom()}, true, "b"},
		{"b", &message.PubPayload{MsgUUID: uuid.NewRandom()}, false, ""},
	}
	var expected []uuid.UUID
	for i, c := range cases {
		if c.exp {
			expected = append(expected, c.pp.MsgUUID)
		}
		require.NoError(t, brk.Publish(c.ch, c.pp), "Publish %d", i)
		if c.unsb != "" {
			require.NoError(t, psc.Unsubscribe(c.unsb, false), "Unsubscribe %d", i)
		}
	}

	time.Sleep(10 * time.Millisecond) // ensure time to pop the last message :(
	require.NoError(t, psc.Close(), "close pubsub connection")
	wg.Wait()
	if assert.Error(t, psc.EventsErr(), "EventsErr returns the error") {
		assert.Contains(t, psc.EventsErr().Error(), "use of closed", "EventsErr is the expected error")
	}
	assert.Equal(t, expected, uuids, "got expected UUIDs")
}
