package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type GroupServer struct {
	Group  string  `yaml:"group"`
	Server *Server `yaml:"server"`
}

type Server struct {
	Type          string `yaml:"type,omitempty"`
	Name          string `yaml:"name,omitempty"`
	Address       string `yaml:"address,omitempty"`
	Port          int    `yaml:"port,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Network       string `yaml:"network,omitempty"`
	SNI           string `yaml:"sni,omitempty"`
	Path          string `yaml:"path,omitempty"`
	AllowInsecure bool   `yaml:"allow_insecure"`
}

func main() {
	subscriptionURL := "https://yfjc.cfd/api/v1/client/subscribe?token=edcabf66a10d757f9a0a724f15d44d84"
	resp, err := http.Get(subscriptionURL)
	if err != nil {
		fmt.Printf("request err: %s\n", err.Error())
		os.Exit(-1)
	}

	base64Bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read body err: %s\n", err.Error())
		os.Exit(-1)
	}

	serversBytes, err := base64.StdEncoding.DecodeString(string(base64Bytes))
	if err != nil {
		fmt.Printf("decode err: %s\n", err.Error())
		os.Exit(-1)
	}

	var servers []*GroupServer
	for _, serverLine := range strings.Split(string(serversBytes), "\n") {
		u, err := url.Parse(serverLine)
		if err != nil {
			fmt.Printf("parse url err: %s\n", err.Error())
			continue
		}

		if u.Scheme != "trojan" {
			continue
		}

		server := &Server{
			Type:    u.Scheme,
			Address: u.Hostname(),
		}

		server.Name, _ = url.QueryUnescape(u.Fragment)
		server.Port, _ = strconv.Atoi(u.Port())
		server.Path = u.Query().Get("path")
		server.Password = u.User.Username()

		if u.Query().Has("type") {
			server.Network = u.Query().Get("type")
		} else {
			server.Network = "tcp"
		}

		server.SNI = u.Query().Get("sni")
		if u.Query().Get("allowInsecure") == "0" {
			server.AllowInsecure = false
		} else {
			server.AllowInsecure = true
		}

		servers = append(servers, &GroupServer{
			Group:  "yifenjichang.com",
			Server: server,
		})
	}

	bb, err := yaml.Marshal(servers)
	if err != nil {
		fmt.Printf("yaml marshal err: %s\n", err.Error())
		os.Exit(-1)
	}
	fmt.Println(string(bb))
}
