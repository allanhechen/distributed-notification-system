package testutil

import (
	"context"
	"fmt"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	toxiproxyTc "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
)

type RabbitMqContainer struct {
	rabbitmqContainer  testcontainers.Container
	toxiproxyContainer testcontainers.Container
	ConnString         string
	containerId        string
	proxy              *toxiproxy.Proxy
}

func GetRabbitMqContainer(ctx context.Context) (*RabbitMqContainer, error) {
	const (
		user           = "guest"
		password       = "guest"
		containerAlias = "rabbitmq"
	)
	nw, err := network.New(ctx)

	rmc, err := rabbitmq.Run(ctx, "rabbitmq:4.1.4-alpine",
		rabbitmq.WithAdminUsername(user),
		rabbitmq.WithAdminPassword(password),
		network.WithNetwork([]string{"rabbitmq"}, nw),
	)
	if err != nil {
		return nil, err
	}
	containerId := rmc.GetContainerID()

	tpc, err := toxiproxyTc.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		network.WithNetwork([]string{"toxiproxy"}, nw),
		toxiproxyTc.WithProxy("rabbitmq-proxy", "rabbitmq:5672"),
	)
	toxiURI, err := tpc.URI(ctx)
	if err != nil {
		return nil, err
	}
	cli := toxiproxy.NewClient(toxiURI)
	proxy, err := cli.Proxy("rabbitmq-proxy")
	if err != nil {
		return nil, err
	}

	host, port, err := tpc.ProxiedEndpoint(8666)
	if err != nil {
		return nil, err
	}
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)

	return &RabbitMqContainer{
		rabbitmqContainer:  rmc,
		toxiproxyContainer: tpc,
		ConnString:         connString,
		containerId:        containerId,
		proxy:              proxy,
	}, nil
}

func (r *RabbitMqContainer) Disconnect() error {
	return r.proxy.Disable()
}

func (r *RabbitMqContainer) Reconnect() error {
	return r.proxy.Enable()
}

func (r *RabbitMqContainer) Close(ctx context.Context) error {
	err := r.toxiproxyContainer.Terminate(ctx)
	if err != nil {
		return err
	}
	err = r.rabbitmqContainer.Terminate(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMqContainer) DeclareEntities(ctx context.Context) error {
	return rabbitmqnotifications.DeclareEntities(ctx, r.ConnString)
}
