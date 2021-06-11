package qingcloud

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	qingcloudconfig "github.com/yunify/qingcloud-sdk-go/config"
	"github.com/yunify/qingcloud-sdk-go/service"
)

const (
	QingCloudAPIKey    = "QINGCLOUD_API_KEY"
	QingCloudAPISecret = "QINGCLOUD_API_SECRET"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	communicator.Config `mapstructure:",squash"`
	ApiKey              string `mapstructure:"api_key"`
	ApiSecret           string `mapstructure:"api_secret"`
	Zone                string `mapstructure:"zone"`
	Protocol            string `mapstructure:"protocol"`
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	Uri                 string `mapstructure:"uri"`
	LogLevel            string `mapstructure:"log_level"`
	VxnetID             string `mapstructure:"vxnet_id"`
	EIPID               string `mapstructure:"eip_id"`
	SecurityGroupID     string `mapstructure:"securitygroup_id"`
	KeypairID           string `mapstructure:"keypair_id"`
	BaseImageID         string `mapstructure:"image_id"`
	ImageArtifactName   string `mapstructure:"image_name"`
	CPU                 int    `mapstructure:"cpu"`
	Memory              int    `mapstructure:"memory"`
	InstanceClass       int    `mapstructure:"instance_class"`
	ctx                 interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	warnings := []string{}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, warnings, err
	}

	var ok bool
	// 如果 APIKey / APISecret 为空，则从环境变量中获取
	if c.ApiKey == "" {
		c.ApiKey, ok = os.LookupEnv(QingCloudAPIKey)
		if !ok {
			return nil, warnings, fmt.Errorf("%s is empty", QingCloudAPIKey)
		}
		warnings = append(warnings, "Got API key from env")
	}
	if c.ApiSecret == "" {
		c.ApiSecret, ok = os.LookupEnv(QingCloudAPISecret)
		if !ok {
			return nil, warnings, fmt.Errorf("%s is empty", QingCloudAPISecret)
		}
		warnings = append(warnings, "Got API secret from env")

	}
	if c.Zone == "" {
		c.Zone = "pek3a"
		warnings = append(warnings, "Set zone to default(pek3a)")
	}
	if c.BaseImageID == "" {
		c.BaseImageID = "xenial3x64"
		warnings = append(warnings, "Set base image to default(ubuntu xenial3x64)")
	}
	if c.Host == "" {
		c.Host = "api.qingcloud.com"
	}
	if c.Port == 0 {
		c.Port = 443
	}
	if c.Protocol == "" {
		c.Protocol = "https"
	}
	if c.Uri == "" {
		c.Uri = "/iaas"
	}
	if c.VxnetID == "" {
		c.VxnetID = "vxnet-0"
		warnings = append(warnings, "Set vxnet to default(vxnet-0)")
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.PackerConfig.PackerDebug {
		c.LogLevel = "debug"
	}

	if c.CPU ==0 {
		c.CPU =1
	}

	if c.Memory == 0 {
		c.Memory =1024
	}
	err = c.validate()
	if err != nil {
		return nil,warnings,err
	}
	errs := c.Config.Prepare(&c.ctx)
	if len(errs) >0 {
		return nil,warnings,errs[0]
	}
	return c, warnings, nil
}

func (config *Config) validate() error {
	qservice := config.GetQingCloudService()

	//validate apikey and api secret
	describeZoneOutput, err := qservice.DescribeZones(&service.DescribeZonesInput{})
	if err != nil {
		return err
	}
	if *describeZoneOutput.RetCode != 0 {
		return fmt.Errorf("describe zone failed: return code is %d", *describeZoneOutput.RetCode)
	}

	//validate base image
	describeImageInput := &service.DescribeImagesInput{
		Images: []*string{&config.BaseImageID},
	}
	imageService, _ := qservice.Image(config.Zone)
	describeImageOutput, err := imageService.DescribeImages(describeImageInput)
	if err != nil {
		return err
	}
	if *describeImageOutput.TotalCount != 1 {
		return fmt.Errorf("image is not found")
	}

	//validate VxnetID
	if len(config.VxnetID) > 0 {
		vxnetService, err := qservice.VxNet(config.Zone)
		if err != nil {
			return err
		}
		describeVxnetsOutput, err := vxnetService.DescribeVxNets(&service.DescribeVxNetsInput{VxNets: []*string{&config.VxnetID}})
		if err != nil {
			return err
		}
		if *describeVxnetsOutput.RetCode != 0 || *describeVxnetsOutput.TotalCount != 1 {
			return fmt.Errorf("vxnet is not found, %s", *describeVxnetsOutput.Message)
		}
	}

	//validate security group
	if len(config.SecurityGroupID) > 0 && config.SecurityGroupID != AllocateNewID {
		securityGroupService, err := qservice.SecurityGroup(config.Zone)
		if err != nil {
			return err
		}
		describeSecurityGroupOutput, err := securityGroupService.DescribeSecurityGroups(
			&service.DescribeSecurityGroupsInput{SecurityGroups: []*string{&config.SecurityGroupID}},
		)
		if err != nil {
			return err
		}
		if *describeSecurityGroupOutput.RetCode != 0 || *describeSecurityGroupOutput.TotalCount != 1 {
			return fmt.Errorf("security group is not found, %s", *describeSecurityGroupOutput.Message)
		}
	}

	if len(config.EIPID) > 0 && config.EIPID != AllocateNewID {
		eipService, err := qservice.EIP(config.Zone)
		if err != nil {
			return err
		}
		describeEIPoutput, err := eipService.DescribeEIPs(&service.DescribeEIPsInput{EIPs: []*string{&config.EIPID}})
		if err != nil {
			return err
		}
		if *describeEIPoutput.RetCode != 0 || *describeEIPoutput.TotalCount != 1 {
			return fmt.Errorf("EIP is not found, %s", *describeEIPoutput.Message)
		}
	}

	//validate KeypairID
	if len(config.KeypairID) > 0 && config.KeypairID != AllocateNewID && config.KeypairID != LocalKey {
		keypairService, err := qservice.KeyPair(config.Zone)
		if err != nil {
			return err
		}
		describeKeypairOutput, err := keypairService.DescribeKeyPairs(
			&service.DescribeKeyPairsInput{
				KeyPairs: []*string{&config.KeypairID},
			})
		if err != nil {
			return err
		}
		if *describeKeypairOutput.RetCode != 0 || *describeKeypairOutput.TotalCount != 1 {
			return fmt.Errorf("keypair is not found,%s", *describeKeypairOutput.Message)
		}
	}
	if len(config.ImageArtifactName) == 0 {
		config.ImageArtifactName = "packer" + config.PackerBuildName
	}

	return nil
}

func (config *Config) GetQingCloudService() *service.QingCloudService {
	qconfig, _ := qingcloudconfig.NewDefault()
	qconfig.AccessKeyID = config.ApiKey
	qconfig.SecretAccessKey = config.ApiSecret
	qconfig.Zone = config.Zone
	qconfig.Protocol = config.Protocol
	qconfig.Host = config.Host
	qconfig.Port = config.Port
	qconfig.URI = config.Uri
	qconfig.LogLevel = config.LogLevel
	qservice, _ := service.Init(qconfig)
	return qservice
}
