package copydefaultconfig

import (
	"github.com/walnuts1018/code-server-operator/internal/initplugins/common"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type copyDefaultConfigPlugin struct {
	Image      string `required:"true" json:"image"`
	VolumeName string `required:"true" json:"volumeName"`
}

func New(params map[string]string) (common.PluginInterface, error) {
	var plugin copyDefaultConfigPlugin
	err := common.Parse(&plugin, params)
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (p *copyDefaultConfigPlugin) GenerateInitContainerApplyConfiguration() *corev1apply.ContainerApplyConfiguration {

	// /home/coder/.localをImageからコピーする（Volumeに存在しなければ）
	command := `
		if [ ! -d /persistent/.local ]; then
			cp -r /home/coder/.local /persistent/;
			echo 'copied .local';
		else
			echo '.local directory already exists';
		fi
	`

	initcontainer := corev1apply.Container().
		WithName("copy-default-config").
		WithImage(p.Image).
		WithCommand("sh", "-c", command).
		WithVolumeMounts(corev1apply.VolumeMount().
			WithName(p.VolumeName).
			WithMountPath("/persistent"))

	return initcontainer
}
