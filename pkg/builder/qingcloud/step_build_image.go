package qingcloud

import (
	"github.com/hashicorp/packer/helper/multistep"
	"context"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"
	"github.com/hashicorp/packer/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
	"github.com/yunify/qingcloud-sdk-go/client"
)

type StepBuildImage struct {

}

func (step *StepBuildImage) Run(ctx context.Context,state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	instanceID := state.Get(InstanceID).(string)
	qservice:=config.GetQingCloudService()
	imageService,err:= qservice.Image(config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	imageOutput,err:=imageService.CaptureInstance(&service.CaptureInstanceInput{Instance:&instanceID,ImageName:&config.ImageArtifactName})
	if err  != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	jobService,err:= qservice.Job(config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err= client.WaitJob(jobService,*imageOutput.JobID,DefaultTimeout,DefaultInterval)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (step *StepBuildImage) Cleanup(multistep.StateBag) {

}

