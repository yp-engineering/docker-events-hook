// Copyright 2015 YP LLC.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file
package main

import (
	"log"
	"net/rpc/jsonrpc"

	"github.com/fsouza/go-dockerclient"
	"github.com/natefinch/pie"
	"github.com/yp-engineering/docker-events-hook/plugin"
)

func main() {
	log.SetPrefix("[echo log] ")

	p := pie.NewProvider()
	if err := p.Register(Api{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

type Api struct{}

func (Api) Attach(container *docker.Container, response *string) error {
	log.Printf("got call for Attach with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Create(container *docker.Container, response *string) error {
	log.Printf("got call for Create with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Delete(event *docker.APIEvents, response *string) error {
	log.Printf("got call for Delete with event %#v", event)

	*response = "done"
	return nil
}

func (Api) Destroy(event *docker.APIEvents, response *string) error {
	log.Printf("got call for Destroy with event %#v", event)

	*response = "done"
	return nil
}

func (Api) Die(container *docker.Container, response *string) error {
	log.Printf("got call for Die with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Kill(container *docker.Container, response *string) error {
	log.Printf("got call for Kill with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Resize(container *docker.Container, response *string) error {
	log.Printf("got call for Resize with container %#v", container)

	*response = "done"
	return nil
}

func (Api) Start(container *docker.Container, response *string) error {
	log.Printf("got call for Start with container %#v", container)
	rp, err := plugin.RunningPort(container)
	if err != nil {
		log.Printf("Error: %s", err)
	} else {
		log.Printf("Running port: %#v", rp)
	}

	localIP, err := plugin.LocalIPAddress()
	if err != nil {
		log.Printf("%#v", err)
	} else {
		log.Printf("IP Address: %#v", localIP)
	}

	*response = "done"
	return nil
}

func (Api) Untag(event *docker.APIEvents, response *string) error {
	log.Printf("got call for Untag with event %#v", event)

	*response = "done"
	return nil
}
