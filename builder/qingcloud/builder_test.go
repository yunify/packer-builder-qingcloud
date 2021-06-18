package qingcloud

import (
	"context"
	"fmt"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"testing"
	"time"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"api_key":      "your api key",
		"api_secret":   "your secret",
		"ssh_username": "root",
		"ssh_password": "Abcd@1234",
	}
	_, warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

type qingContext struct {
	context.Context
}

func (s *qingContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (s *qingContext) Done() <-chan struct{} {
	return nil
}

func (s *qingContext) Err() error {
	return nil
}

func TestBuilder_Run(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"api_key":      "your api key",
		"api_secret":   "your secret",
		"ssh_username": "root",
		"ssh_password": "Abcd@1234",
		"zone":         "pek3a",
		"image_id":     "ubuntu xenial3x64",
		"vxnet_id":     "vxnet-0",
		"log_level":    "info",
	}
	_, _, _ = b.Prepare(c)
	var ctx context.Context = &qingContext{}
	var ui packersdk.Ui = &packersdk.MockUi{}
	var hook packersdk.Hook = &packersdk.MockHook{RunFunc: nil,

		RunCalled: true,
		RunComm: &packersdk.MockCommunicator{

		},
		RunData: c,
		RunName: "name",
		RunUi: &packersdk.MockUi{

		}}
	artifact, err := b.Run(ctx, ui, hook)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}
	fmt.Println(artifact)
}
