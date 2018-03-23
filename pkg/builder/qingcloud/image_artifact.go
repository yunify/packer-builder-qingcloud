package qingcloud

import (
	"fmt"
	"github.com/yunify/qingcloud-sdk-go/service"
)

type ImageArtifact struct {
	ImageID string
	ImageService *service.ImageService
}

func (artifact *ImageArtifact) BuilderId() string {
	return BuilderId
}

func (artifact *ImageArtifact) Files() []string {
	return []string{}
}

func (artifact *ImageArtifact) Id() string {
	return artifact.ImageID
}

func (artifact *ImageArtifact) String() string {
	return fmt.Sprintf("QingCloud image %s",artifact.ImageID)
}

func (artifact *ImageArtifact) State(name string) interface{} {
	artifact.ImageService.DescribeImages(&service.DescribeImagesInput{Images:[]*string{}})

}

func (artifact *ImageArtifact) Destroy() error {
	artifact.ImageService.DeleteImages(&service.DeleteImagesInput{})
}
