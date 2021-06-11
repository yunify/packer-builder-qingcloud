package main

import (
	"fmt"
	"github.com/yunify/packer-plugin-qingcloud/builder/qingcloud"
	qingcloudVersion "github.com/yunify/packer-plugin-qingcloud/version"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("qingcloud", new(qingcloud.Builder))
	pps.SetVersion(qingcloudVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
