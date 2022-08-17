package nats

import (
	"fmt"

	stan "github.com/nats-io/stan.go"
)

type Nats struct {
	connection   stan.Conn
	subscription stan.Subscription
}

func Create(clusterId, clientId, channelId string, messageChannel chan<- string) (*Nats, error) {
	connection, err := stan.Connect(clusterId, clientId)
	if err != nil {
		return nil, err
	}

	subscription, err := connection.Subscribe(channelId, func(m *stan.Msg) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		messageChannel <- string(m.Data)
	})
	if err != nil {
		return nil, err
	}

	return &Nats{connection, subscription}, nil
}

func (nats Nats) Destroy() {
	nats.subscription.Unsubscribe()
	nats.connection.Close()
}
