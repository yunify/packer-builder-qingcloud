package qingcloud

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/yunify/qingcloud-sdk-go/client"
	"github.com/yunify/qingcloud-sdk-go/service"
	"github.com/yunify/qingcloud-sdk-go/utils"

)

type StepEnsureIP struct {
}

func (step *StepEnsureIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(BuilderConfig).(Config)
	ui := state.Get(UI).(packer.Ui)
	ui.Message("Create eip if needed")

	instanceID := state.Get(InstanceID).(string)
	if config.VxnetID == "vxnet-0" || len(config.EIPID) > 0 {
		qservice := config.GetQingCloudService()
		eipService, err := qservice.EIP(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		var eipid string
		//allocate eip
		if len(config.EIPID) == 0 || config.EIPID == AllocateNewID {

			allocateEipOutput, err := eipService.AllocateEIPs(&service.AllocateEIPsInput{Bandwidth: service.Int(5)})
			if err != nil {
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			if *allocateEipOutput.RetCode != 0 {
				ui.Error("Failed to allocate eip when vxnet-0 is chosen, packer may failed to connect to this machine")
				return multistep.ActionHalt
			}
			eipid = *allocateEipOutput.EIPs[0]
		} else {
			eipid = config.EIPID
		}

		// attach eip to machine
		associateJobOutput, err := eipService.AssociateEIP(
			&service.AssociateEIPInput{EIP: service.String(eipid), Instance: service.String(instanceID)},
		)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if *associateJobOutput.RetCode != 0 {
			ui.Error(*associateJobOutput.Message)
			return multistep.ActionHalt
		}

		jobService, err := qservice.Job(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		err = client.WaitJob(jobService, *associateJobOutput.JobID, DefaultTimeout, DefaultInterval)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		instanceService, _ := qservice.Instance(config.Zone)
		_, err = client.WaitInstanceStatus(instanceService, instanceID, client.InstanceStatusRunning, DefaultTimeout, DefaultInterval)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		describeEipoutput, err := eipService.DescribeEIPs(&service.DescribeEIPsInput{EIPs: []*string{service.String(eipid)}})
		if err != nil {
			if *describeEipoutput.RetCode != 0 && *describeEipoutput.TotalCount != 1 {
				ui.Error("Failed to describe eip")
				return multistep.ActionHalt
			}
		}
		state.Put(EIPID, eipid)
		state.Put(PublicIP, *describeEipoutput.EIPSet[0].EIPAddr)
	}

	return multistep.ActionContinue
}

func (step *StepEnsureIP) Cleanup(state multistep.StateBag) {
	ui := state.Get(UI).(packer.Ui)
	ui.Message("clean up eip if needed")
	value, ok := state.GetOk(EIPID)
	eip := value.(string)
	if ok {
		config := state.Get(BuilderConfig).(Config)
		if config.EIPID != eip {
			qservice := config.GetQingCloudService()
			eipService, err := qservice.EIP(config.Zone)
			if err != nil {
				ui.Error(err.Error())
			}
			disassociateEIPoutput, err := eipService.DissociateEIPs(
				&service.DissociateEIPsInput{EIPs: []*string{service.String(eip)}})
			if disassociateEIPoutput.Message != nil {
				ui.Message(*disassociateEIPoutput.Message)
			}
			jobService, err := qservice.Job(config.Zone)
			if err != nil {
				ui.Error(err.Error())
			}
			client.WaitJob(jobService, *disassociateEIPoutput.JobID, DefaultTimeout, DefaultInterval)
			ui.Message("Disassociate eip with machine")
			errorTimes := 0
			err = utils.WaitForSpecificOrError(func() (bool, error) {
				_, err = eipService.ReleaseEIPs(&service.ReleaseEIPsInput{EIPs: []*string{service.String(eip)}})
				if err != nil {
					errorTimes++
					if errorTimes > 3 {
						return false, err
					}
					return false, nil
				}
				return true, nil
			}, DefaultTimeout, DefaultInterval)
			if err != nil {
				ui.Error(err.Error())
			}
			ui.Message("Released eip")
		}
	}
}
