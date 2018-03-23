package qingcloud

import (
	"context"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/yunify/qingcloud-sdk-go/service"
	"io/ioutil"
)

type StepEnsureKeypair struct {
}



func (step *StepEnsureKeypair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config, _ := state.Get(BuilderConfig).(Config)
	ui, _ := state.Get(UI).(packer.Ui)
	qservice := config.GetQingCloudService()

	var loginKeyPairID string
	var privateKey string
	//security group not found, create one
	keypairService, err := qservice.KeyPair(config.Zone)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if len(config.KeypairID) == 0 || config.KeypairID == AllocateNewID {
		keypairOutput,err:=keypairService.CreateKeyPair(
			&service.CreateKeyPairInput{
				KeyPairName: service.String("packer" + config.PackerConfig.PackerBuildName),
				Mode: service.String("system"),
				EncryptMethod: service.String("ssh-rsa"),
			},
		)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		privateKey=*keypairOutput.PrivateKey
		loginKeyPairID = *keypairOutput.KeyPairID

	} else if config.KeypairID == LocalKey{
		publicKey,err:= loadFileContent(DefaultPublicKey)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		privateKey,err= loadFileContent(DefaultPrivateKey)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		keypairOutput, err := keypairService.CreateKeyPair(
			&service.CreateKeyPairInput{
				KeyPairName: service.String("packer" + config.PackerConfig.PackerBuildName),
				Mode: service.String("user"),
				PublicKey:service.String(publicKey),
				})
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		loginKeyPairID = *keypairOutput.KeyPairID
		privateKey = *keypairOutput.PrivateKey

	} else {
		loginKeyPairID=config.KeypairID
		privateKey=config.SSHPrivateKey
	}
	state.Put(LoginKeyPairID, loginKeyPairID)
	state.Put(PrivateKey,privateKey)

	return multistep.ActionHalt
}

func (step *StepEnsureKeypair) Cleanup(state multistep.StateBag) {
	config, _ := state.Get(BuilderConfig).(Config)
	ui, _ := state.Get(UI).(packer.Ui)
	keypairID, ok := state.Get(LoginKeyPairID).(string)
	if ok && keypairID != config.KeypairID {
		qservice := config.GetQingCloudService()
		keypairService, err := qservice.KeyPair(config.Zone)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		keypairService.DeleteKeyPairs(&service.DeleteKeyPairsInput{KeyPairs: []*string{service.String(keypairID)}})
	}
}

