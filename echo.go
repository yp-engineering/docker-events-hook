// Copyright 2015 YP LLC.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file
package main

import (
	"log"
	"net/rpc/jsonrpc"

	"github.com/fsouza/go-dockerclient"
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

func (Api) Start(container *docker.Container, response *string) error {
	log.Printf("got call for Start with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Attach(container *docker.Container, response *string) error { return nil }
func (Api) Create(container *docker.Container, response *string) error { return nil }
func (Api) Delete(event *docker.APIEvents, response *string) error     { return nil }
func (Api) Destroy(event *docker.APIEvents, response *string) error    { return nil }
func (Api) Die(container *docker.Container, response *string) error    { return nil }
func (Api) Resize(container *docker.Container, response *string) error { return nil }
func (Api) Untag(event *docker.APIEvents, response *string) error      { return nil }
