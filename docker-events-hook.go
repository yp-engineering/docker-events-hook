// Copyright 2015 YP LLC.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"
	"github.com/natefinch/pie"
	"gopkg.in/yaml.v2"
)

const (
	_ = iota
	verbose
	debug
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
	plugins      map[*rpc.Client]string
)

func init() {
	flag.StringVar(&configFile, "config", configFile, "path to overiding config.yml file")
	flag.Parse()

	parseConfig()
	dockerClient = newDockerClient()
	plugins = createPlugins()
}

func fatal(err error) {
	if err != nil {
		glog.Fatalf("fatal: %s", err)
	}
}

func parseConfig() {
	var toLoad []byte
	if configFile != "" {
		var err error
		toLoad, err = ioutil.ReadFile(configFile)
		fatal(err)
	} else {
		toLoad = []byte(defaultConfig)
	}

	fatal(yaml.Unmarshal(toLoad, &config))

	if glog.V(debug) {
		glog.Infof("config: %#v", config)
	}
}

func newDockerClient() *docker.Client {
	client, err := docker.NewVersionedClient(config.Docker.Endpoint, config.Docker.Version)
	fatal(err)
	if glog.V(debug) {
		glog.Infof("docker client: %#v", client)
	}
	return client
}

func createPlugins() map[*rpc.Client]string {
	plugins = make(map[*rpc.Client]string)
	for _, plugin := range config.Plugins {
		plug, err := pie.StartProviderCodec(jsonrpc.NewClientCodec, os.Stderr, plugin)
		fatal(err)
		plugins[plug] = plugin
	}
	if glog.V(debug) {
		glog.Infof("plugins: %#v", plugins)
	}
	return plugins
}

func dockerInspect(event *docker.APIEvents) (*docker.Container, error) {
	status := event.Status
	if status == "delete" || status == "destroy" || status == "untag" {
		// No error, just can't inspect them
		return nil, nil
	} else {
		return dockerClient.InspectContainer(event.ID)
	}
}

func marshalJson(d interface{}) []byte {
	data, err := json.Marshal(d)
	if err != nil {
		glog.Errorf("couldn't marshal to JSON: %#v", data)
		return nil
	}
	return data
}

func eventInfo(event *docker.APIEvents, name string, inspect *docker.Container, result string, call string) {
	if glog.V(debug) {
		logMessage := `{"Event":%s,"Plugin":%#v,"Container":%s,"Result":%#v,"Plugin Call":%#v}`
		glog.Infof(logMessage, marshalJson(event), name, marshalJson(inspect), result, call)
	} else {
		var image string
		if inspect == nil {
			image = ""
		} else {
			image = inspect.Config.Image
		}
		if glog.V(verbose) {
			logMessage := `{"Event":{"Id":%#v,"Status":%#v},"Plugin":%#v,"Container":%#v,"Result":%#v,"Plugin Call":%#v}`
			glog.Infof(logMessage, event.ID, event.Status, name, image, result, call)
		} else {
			logMessage := `{"Event":{"Status":%#v},"Plugin":%#v,"Container":%#v}`
			glog.Infof(logMessage, event.Status, name, image)
		}
	}
}

// Decision switch for what type of events we receive from daemon
func handleEvent(event *docker.APIEvents) {
	for plugin, name := range plugins {
		go func() {
			var result string
			var pluginError error

			inspect, err := dockerInspect(event)
			if err != nil {
				glog.Errorf("couldn't inspect event: %#v", event)
				return
			}

			call := "Api." + strings.Title(event.Status)
			switch event.Status {
			case "attach":
				pluginError = plugin.Call(call, inspect, &result)
			case "create":
				pluginError = plugin.Call(call, inspect, &result)
			case "delete":
				pluginError = plugin.Call(call, event, &result)
			case "destroy":
				pluginError = plugin.Call(call, event, &result)
			case "die":
				pluginError = plugin.Call(call, inspect, &result)
			case "resize":
				pluginError = plugin.Call(call, inspect, &result)
			case "start":
				pluginError = plugin.Call(call, inspect, &result)
			case "untag":
				pluginError = plugin.Call(call, event, &result)
			default:
				glog.Errorf("unknown event: %v", event)
			}
			eventInfo(event, name, inspect, result, call)

			if pluginError != nil {
				glog.Errorf("plugin error: %#v", pluginError)
			}
		}()
	}
}

func main() {
	eventChannel := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventChannel)
	for {
		handleEvent(<-eventChannel)
	}
}
