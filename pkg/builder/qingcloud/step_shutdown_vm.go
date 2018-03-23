package qingcloud

import (
	"github.com/hashicorp/packer/helper/multistep"
	"context"
	"github.com/hashicorp/packer/packer"
	"github.com/yunify/qingcloud-sdk-go/client"
	"github.com/yunify/qingcloud-sdk-go/service"
)

type StepShutDownVM struct {

}

func (step *StepShutDownVM) Run(ctx context.Context,state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	instanceID := state.Get(InstanceID).(string)
	qservice:=config.GetQingCloudService()
	instanceService,err:=qservice.Instance(config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	_,err=instanceService.StopInstances(&service.StopInstancesInput{Instances:[]*string{service.String(instanceID)}})
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_,err=client.WaitInstanceStatus(instanceService,instanceID,client.InstanceStatusStopped,DefaultTimeout,DefaultInterval)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (step *StepShutDownVM) Cleanup(state multistep.StateBag) {
	return
}

