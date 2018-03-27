package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/yunify/packer-builder-qingcloud/pkg/builder/qingcloud"
)

func main() {
	server, err := plugin.Server()
	if err != nil {

	}
	server.RegisterBuilder(new(qingcloud.Builder))
	server.Serve()
}
