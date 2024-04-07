package common

import (
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type CommonFields struct {
	Image      string `required:"true" json:"image"`
	VolumeName string `required:"true" json:"volumeName"`
}

type PluginInterface interface {
	GenerateInitContainerApplyConfiguration() *corev1apply.ContainerApplyConfiguration
}
