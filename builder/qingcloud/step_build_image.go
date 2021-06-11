package qingcloud

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"context"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
	"github.com/yunify/qingcloud-sdk-go/client"
)

type StepBuildImage struct {

}

func (step *StepBuildImage) Run(ctx context.Context,state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	instanceID := state.Get(InstanceID).(string)
	ui.Message("Start to capture image")
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
	state.Put(ImageID,*imageOutput.ImageID)
	return multistep.ActionContinue
}

func (step *StepBuildImage) Cleanup(multistep.StateBag) {

}

