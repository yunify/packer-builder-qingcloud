package qingcloud

import (
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	gossh "golang.org/x/crypto/ssh"
	validator "gopkg.in/asaskevich/govalidator.v8"
)

const BuilderId = "yunify.qingcloud"

const (
	BuilderConfig   = "config"
	UI              = "ui"
	InstanceID      = "instancd_id"
	ImageID         = "image_id"
	SecurityGroupID = "security_group_id"
	EIPID           = "eip_id"
	PublicIP        = "public_ip"
	PrivateIP       = "private_ip"
	LoginKeyPairID  = "keypair_id"
	PrivateKey      = "private_key_content"
)

const (
	AllocateNewID = "new"
)

const (
	DefaultPublicKey  = "~/.ssh/id_rsa.pub"
	DefaultPrivateKey = "~/.ssh/id_rsa"
	LocalKey          = "local"
)

var DefaultTimeout = time.Second * 300
var DefaultInterval = time.Second * 5

//Builder qingcloud builder
type Builder struct {
	runner multistep.BasicRunner
	config Config
}

func (builder *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	builder.config = *c
	return nil, nil
}

func (builder *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Setup
	state := new(multistep.BasicStateBag)
	state.Put(BuilderConfig, builder.config)
	state.Put(UI, ui)
	state.Put("hook", hook)

	// Run
	steps := []multistep.Step{
		new(StepEnsureSecurityGroup),
		new(StepEnsureKeypair),
		new(StepCreateVM),
		new(StepEnsureIP),
		&communicator.StepConnect{
			Config:    &builder.config.Config,
			Host:      builder.getHost,
			SSHConfig: builder.getSSHConfig,
		},
		new(common.StepProvision),
		new(StepShutDownVM),
		new(StepBuildImage),
	}
	builder.runner = multistep.BasicRunner{Steps: steps}
	builder.runner.Run(state)
	imageID, ok := state.GetOk(ImageID)
	if  !ok {
		return nil, fmt.Errorf("Failed to get image id:%v",imageID)
	}

	imageService ,_:= builder.config.GetQingCloudService().Image(builder.config.Zone)
	artifact := &ImageArtifact{
		ImageID: imageID.(string),
		ImageService: imageService,
	}
	return artifact, nil
}

func (builder *Builder) Cancel() {
	builder.runner.Cancel()
}

func (builder *Builder) getHost(state multistep.StateBag) (string, error) {
	publicIP, ok := state.Get(PublicIP).(string)
	if ok && validator.IsIP(publicIP) {
		return publicIP, nil
	}
	privateIP, ok := state.Get(PrivateIP).(string)
	if ok && validator.IsIP(privateIP) {
		return privateIP, nil
	}
	return "", fmt.Errorf("neither public ip nor private ip is valid")
}

func (builder *Builder) getSSHConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get(BuilderConfig).(Config)
	privateKey := state.Get(PrivateKey).(string)
	signer, err := gossh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to set up ssh config：%v", err)
	}
	return &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}, nil
}
