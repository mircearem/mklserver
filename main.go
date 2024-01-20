package main

import (
	"context"
	"time"

	_ "github.com/mircearem/mklserver/log"
	network "github.com/mircearem/mklserver/server"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := network.NewServerConfig().
		WithListenAddr(":3000").
		WithHandshakeTimeout(2 * time.Second).
		WithKeepAlive(15 * time.Second)
	ctx := context.Background()

	s, err := network.NewServer(cfg, ctx)
	if err != nil {
		logrus.Fatalln(err)
	}
	s.Run()
}
