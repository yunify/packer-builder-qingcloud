package qingcloud

import (
	"github.com/hashicorp/packer/helper/multistep"
	"context"
	"github.com/hashicorp/packer/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
	"github.com/yunify/qingcloud-sdk-go/client"
)

type StepCreateVM struct {

}

func (step *StepCreateVM) Run(ctx context.Context,state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	qservice:=config.GetQingCloudService()
	instanceService,err:=qservice.Instance(config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	securityGroupID := state.Get(SecurityGroupID).(string)
	loginKeyPairID := state.Get(LoginKeyPairID).(string)
	instanceJobOutput,err:=instanceService.RunInstances(&service.RunInstancesInput{
		CPU: service.Int(config.CPU),
		Memory: service.Int(config.Memory),
		VxNets:[]*string{service.String(config.VxnetID)},
		InstanceName: service.String("packer"+config.PackerBuildName),
		ImageID:service.String(config.BaseImageID),
		InstanceClass:&config.InstanceClass,
		SecurityGroup:service.String(securityGroupID),
		LoginKeyPair: service.String(loginKeyPairID),
	})
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if *instanceJobOutput.RetCode != 0 || len(instanceJobOutput.Instances) < 1 {
		ui.Error("Failed to create instance:"+*instanceJobOutput.Message)
		return multistep.ActionHalt
	}
	instance,err:=client.WaitInstanceNetwork(instanceService,*instanceJobOutput.Instances[0],DefaultTimeout,DefaultInterval)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put(InstanceID,*instance.InstanceID)

	if len (instance.VxNets) <= 0 {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	vxnet:= instance.VxNets[0]
	state.Put(PrivateIP,*vxnet.PrivateIP)
	return multistep.ActionContinue

}

func (step *StepCreateVM) Cleanup(state multistep.StateBag) {
	instanceID,ok:=state.Get(InstanceID).(string)
	if ok {
		config := state.Get(BuilderConfig).(Config)
		ui := state.Get(UI).(packer.Ui)
		qservice:=config.GetQingCloudService()
		instanceService,err:=qservice.Instance(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		instanceService.TerminateInstances(&service.TerminateInstancesInput{Instances:[]*string{service.String(instanceID)}})
	}
}
