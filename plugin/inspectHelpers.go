package plugin

import (
	"errors"
	"regexp"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

func RunningPort(info *docker.Container) (string, error) {
	var err error
	var port string
	switch info.HostConfig.NetworkMode {
	// `-p` and `-P`
	case "default":
		port, err = mappedTCPPort(info)
		if err == nil {
			return port, nil
		}
	// `--net host` obviously
	case "host":
		port, err = exposedTCPPort(info)
		if err == nil {
			return port, nil
		}
	}
	return "", err
}

// Returns the mapped tcp port from docker inspect. See runningPort to see why.
func mappedTCPPort(info *docker.Container) (string, error) {
	for key, val := range info.NetworkSettings.Ports {
		protocol := key.Proto()
		match, _ := regexp.MatchString("tcp", strings.ToLower(protocol))
		if match && len(val) > 0 {
			return val[0].HostPort, nil
		}
	}
	return "", errors.New("no mapped port found.")
}

// Returns the exposed tcp port from docker inspect. See runningPort to see why.
func exposedTCPPort(info *docker.Container) (string, error) {
	for key := range info.Config.ExposedPorts {
		protocol := key.Proto()
		match, _ := regexp.MatchString("tcp", strings.ToLower(protocol))
		if match {
			return key.Port(), nil
		}
	}
	return "", errors.New("no exposed port found.")
}
