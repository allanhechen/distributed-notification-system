package testutil

import (
	"context"
	"errors"
	"fmt"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	rabbitmqnotifications "github.com/allanhechen/distributed-notification-system/utils/rabbitmq_notifications"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	toxiproxyTc "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
)

// RabbitMqContainer is a struct representing a testcontainers/rabbitmq
// instances intended to be used for testing.
type RabbitMqContainer struct {
	rabbitmqContainer  testcontainers.Container
	toxiproxyContainer testcontainers.Container
	ConnString         string
	proxy              *toxiproxy.Proxy
	nw                 *testcontainers.DockerNetwork
}

// GetRabbitMqContainer initializes a testcontainer for RabbitMQ and
// toxiproxy in the form of a RabbitMqContainer.
func GetRabbitMqContainer(ctx context.Context) (*RabbitMqContainer, error) {
	const (
		user     = "guest"
		password = "guest"
	)
	nw, err := network.New(ctx)
	if err != nil {
		return nil, err
	}

	rmc, err := rabbitmq.Run(ctx, "rabbitmq:4.1.4-alpine",
		rabbitmq.WithAdminUsername(user),
		rabbitmq.WithAdminPassword(password),
		network.WithNetwork([]string{"rabbitmq"}, nw),
	)
	if err != nil {
		return nil, err
	}

	tpc, err := toxiproxyTc.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		network.WithNetwork([]string{"toxiproxy"}, nw),
		toxiproxyTc.WithProxy("rabbitmq-proxy", "rabbitmq:5672"),
	)
	if err != nil {
		return nil, err
	}

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
		proxy:              proxy,
		nw:                 nw,
	}, nil
}

// Disconnect disconnects the network interface of the RabbitMQ container
// using toxiproxy.
func (r *RabbitMqContainer) Disconnect() error {
	return r.proxy.Disable()
}

// Reconnect reconnects the network interface of the RabbitMQ container
// using toxiproxy.
func (r *RabbitMqContainer) Reconnect() error {
	return r.proxy.Enable()
}

// Close closes all containers associated with a RabbitMqContainer, and
// removes the network interface.
func (r *RabbitMqContainer) Close(ctx context.Context) error {
	var errs []error
	if err := r.toxiproxyContainer.Terminate(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := r.rabbitmqContainer.Terminate(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := r.nw.Remove(ctx); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// DeclareEntities uses the utility function to declare RabbitMQ entities
// for the associated testcontainer instance.
func (r *RabbitMqContainer) DeclareEntities(ctx context.Context) error {
	return rabbitmqnotifications.DeclareEntities(ctx, r.ConnString)
}
