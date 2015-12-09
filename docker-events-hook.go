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
  version: 1.21
`
var (
	config       Config
	configFile   string
	dockerClient *docker.Client
	plugins      []rpc.Client
)

func init() {
	configFileMessage := "override default config path of " + configFile
	flag.StringVar(&configFile, "config", configFile, configFileMessage)
	flag.Parse()

	parseConfig()
	dockerClient = newDockerClient()
	plugins = createPlugins()
}

func refute(err error, level string) {
	if err != nil {
		switch level {
		case "fatal":
			log.Fatal("fatal: ", err)
		case "warn":
			log.Print("warn: ", err)
		}
	}
}

func parseConfig() {
	var toLoad []byte
	if configFile != "" {
		var err error
		toLoad, err = ioutil.ReadFile(configFile)
		refute(err, "fatal")
	} else {
		toLoad = []byte(defaultConfig)
	}
	refute(yaml.Unmarshal(toLoad, &config), "fatal")
}

func newDockerClient() *docker.Client {
	client, err := docker.NewVersionedClient(config.Docker.Endpoint, config.Docker.Version)
	refute(err, "fatal")
	return client
}

func createPlugins() []rpc.Client {
	var plugins []rpc.Client
	for _, plugin := range config.Plugins {
		plug, err := pie.StartProviderCodec(jsonrpc.NewClientCodec, os.Stderr, plugin)
		refute(err, "fatal")
		plugins = append(plugins, *plug)
	}
	return plugins
}

func dockerInspect(eventID string) *docker.Container {
	inspect, err := dockerClient.InspectContainer(eventID)
	refute(err, "warn")
	return inspect
}

// Decision switch for what type of events we receive from daemon
func handleEvent(event *docker.APIEvents) {
	log.Print("event: " + event.Status)
	for _, plugin := range plugins {
		go func() {
			var result string
			var pluginError error
			switch event.Status {
			case "attach":
				pluginError = plugin.Call("Api.Attach", dockerInspect(event.ID), &result)
			case "create":
				pluginError = plugin.Call("Api.Create", dockerInspect(event.ID), &result)
			case "delete":
				pluginError = plugin.Call("Api.Delete", event, &result)
			case "destroy":
				pluginError = plugin.Call("Api.Destroy", event, &result)
			case "die":
				pluginError = plugin.Call("Api.Die", dockerInspect(event.ID), &result)
			case "resize":
				pluginError = plugin.Call("Api.Resize", dockerInspect(event.ID), &result)
			case "start":
				pluginError = plugin.Call("Api.Start", dockerInspect(event.ID), &result)
			case "untag":
				pluginError = plugin.Call("Api.Untag", event, &result)
			default:
				log.Printf("unknown event: %v", event)
			}
			if result != "" {
				log.Printf("result: %v", result)
			}
			refute(pluginError, "warn")
		}()
	}
}

func main() {
	log.SetPrefix("[server log] ")

	eventChannel := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventChannel)

	for {
		handleEvent(<-eventChannel)
	}
}
