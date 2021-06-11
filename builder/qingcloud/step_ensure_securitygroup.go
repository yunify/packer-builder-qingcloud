package qingcloud

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
)

type StepEnsureSecurityGroup struct {
}

func (step *StepEnsureSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config, _ := state.Get(BuilderConfig).(Config)
	ui, _ := state.Get(UI).(packer.Ui)
	ui.Message("Create firewall if needed")

	if len(config.SecurityGroupID) == 0 || config.SecurityGroupID == AllocateNewID {
		qservice := config.GetQingCloudService()
		//security group not found, create one
		securityGroupService, err := qservice.SecurityGroup(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}
		securityGroupOutput, err := securityGroupService.CreateSecurityGroup(&service.CreateSecurityGroupInput{SecurityGroupName: service.String("packer" + config.PackerConfig.PackerBuildName)})
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}
		securityGroupID := *securityGroupOutput.SecurityGroupID

		//add rules:enable all
		_, err = securityGroupService.AddSecurityGroupRules(&service.AddSecurityGroupRulesInput{
			Rules: []*service.SecurityGroupRule{
				{
					Protocol: service.String("tcp"),
					Action:   service.String("accept"),
					Priority: service.Int(0),
					Val1:     service.String("1"),
					Val2:     service.String("65534"),
				},
			},
			SecurityGroup: service.String(securityGroupID),
		})
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionContinue
		}

		//apply rule
		securityGroupService.ApplySecurityGroup(&service.ApplySecurityGroupInput{SecurityGroup: service.String(securityGroupID)})

		state.Put(SecurityGroupID, securityGroupID)
	} else {
		state.Put(SecurityGroupID, config.SecurityGroupID)
	}

	return multistep.ActionContinue
}

func (step *StepEnsureSecurityGroup) Cleanup(state multistep.StateBag) {
	config, _ := state.Get(BuilderConfig).(Config)
	ui, _ := state.Get(UI).(packer.Ui)
	ui.Message("Clean up firewall if needed")
	securityGroupID, ok := state.Get(SecurityGroupID).(string)
	if ok && securityGroupID != config.SecurityGroupID {
		qservice := config.GetQingCloudService()
		securityGroupService, err := qservice.SecurityGroup(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		securityGroupService.DeleteSecurityGroups(&service.DeleteSecurityGroupsInput{SecurityGroups: []*string{service.String(securityGroupID)}})
	}
}
