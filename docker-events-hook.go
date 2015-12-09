// Copyright 2015 YP LLC.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/natefinch/pie"
	"gopkg.in/yaml.v2"
)

type DockerConfig struct {
	Endpoint string
	Version  string
}

type Config struct {
	Plugins []string
	Docker  DockerConfig
}

var defaultConfig = `
plugins:
  - ./echo
docker:
  endpoint: unix:///var/run/docker.sock
  version: 1.9
`

var (
	configFile string
	config     Config
)

func init() {
	configFileMessage := "override default config path of " + configFile
	flag.StringVar(&configFile, "config", configFile, configFileMessage)
	flag.Parse()

	parseConfig()
}

func refute(err error) {
	if err != nil {
		log.Fatal("fatal: ", err)
	}
}

func parseConfig() {
	var toLoad []byte
	if configFile != "" {
		var err error
		toLoad, err = ioutil.ReadFile(configFile)
		refute(err)
	} else {
		toLoad = []byte(defaultConfig)
	}
	refute(yaml.Unmarshal(toLoad, &config))
}

func newDockerClient() *docker.Client {
	client, err := docker.NewVersionedClient(config.Docker.Endpoint, config.Docker.Version)
	refute(err)
	return client
}

// Decision switch for what type of events we receive from daemon
func handleEvent(event *docker.APIEvents, plugins []rpc.Client) {
	for _, plugin := range plugins {
		go func() {
			var result string
			switch event.Status {
			case "start":
				refute(plugin.Call("Api.Start", event.ID, &result))
				log.Printf("res: %v", result)
			}
		}()
	}
}

func createPlugins() []rpc.Client {
	var plugins []rpc.Client
	for _, plugin := range config.Plugins {
		plug, err := pie.StartProviderCodec(jsonrpc.NewClientCodec, os.Stderr, plugin)
		refute(err)
		plugins = append(plugins, *plug)
	}
	return plugins
}

func main() {
	dockerClient := newDockerClient()

	eventChannel := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventChannel)

	plugins := createPlugins()

	for {
		handleEvent(<-eventChannel, plugins)
	}
}
