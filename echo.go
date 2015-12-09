// Copyright 2015 YP LLC.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file
package main

import (
	"log"
	"net/rpc/jsonrpc"

	"github.com/natefinch/pie"
)

func main() {
	log.SetPrefix("[plugin log] ")

	p := pie.NewProvider()
	if err := p.Register(Api{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

type Api struct{}

func (Api) Start(eventId string, response *string) error {
	log.Printf("got call for Start with eventId %q", eventId)

	*response = "Hi " + eventId
	return nil
}
