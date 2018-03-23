package qingcloud

import (
	"github.com/hashicorp/packer/helper/multistep"
	"context"
	"github.com/hashicorp/packer/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
	"github.com/yunify/qingcloud-sdk-go/client"
)

type StepEnsureIP struct {
	eipID string
}

func (step *StepEnsureIP) Run(ctx context.Context,state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	instanceID := state.Get(InstanceID).(string)
	if config.VxnetID == "vxnet-0" || len(config.EIPID) >0 {
		qservice := config.GetQingCloudService()
		eipService, err := qservice.EIP(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}

		//allocate eip
		if len(config.EIPID) == 0 || config.EIPID == AllocateNewID{

			allocateEipOutput,err:=eipService.AllocateEIPs(&service.AllocateEIPsInput{Bandwidth:service.Int(5)})
			if err != nil {
				ui.Error(err.Error())
				return multistep.ActionContinue
			}
			if *allocateEipOutput.RetCode != 0 {
				ui.Error("Failed to allocate eip when vxnet-0 is chosen, packer may failed to connect to this machine")
				return multistep.ActionContinue
			}
			step.eipID = *allocateEipOutput.EIPs[0]
		} else {
			step.eipID = config.EIPID
		}

		// attach eip to machine
		associateJobOutput,err:=eipService.AssociateEIP(
			&service.AssociateEIPInput{EIP:service.String(step.eipID),Instance:service.String(instanceID)},
			)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}
		if *associateJobOutput.RetCode != 0 {
			ui.Error(*associateJobOutput.Message)
			return multistep.ActionContinue
		}

		jobService,err := qservice.Job(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}
		err =client.WaitJob(jobService,*associateJobOutput.JobID,DefaultTimeout,DefaultInterval)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}

		describeEipoutput,err:= eipService.DescribeEIPs(&service.DescribeEIPsInput{EIPs:[]*string{service.String(step.eipID)}})
		if err != nil {
			if *describeEipoutput.RetCode != 0 && *describeEipoutput.TotalCount != 1 {
				ui.Error("Failed to describe eip")
				return multistep.ActionContinue
			}
		}
		state.Put(PublicIP,*describeEipoutput.EIPSet[0].EIPAddr)
	}

	return multistep.ActionContinue
}

func (step *StepEnsureIP) Cleanup(state multistep.StateBag) {
	ui := state.Get(UI).(packer.Ui)
	if len(step.eipID) > 0 {
		config := state.Get(BuilderConfig).(Config)
		if config.EIPID != step.eipID {
			qservice := config.GetQingCloudService()
			eipService,err := qservice.EIP(config.Zone)
			if err != nil {
				ui.Error(err.Error())
			}
			disassociateEIPoutput,err:=eipService.DissociateEIPs(
					&service.DissociateEIPsInput{EIPs:[]*string{service.String(step.eipID)}})
			ui.Message(*disassociateEIPoutput.Message)
			jobService,err := qservice.Job(config.Zone)
			if err != nil {
				ui.Error(err.Error())
			}
			client.WaitJob(jobService,*disassociateEIPoutput.JobID,DefaultTimeout,DefaultInterval)
			eipService.ReleaseEIPs(&service.ReleaseEIPsInput{EIPs:[]*string{service.String(step.eipID)}})
		}
	}
}
