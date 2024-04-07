package copyhomeplugin

import (
	"github.com/walnuts1018/code-server-operator/internal/initplugins/common"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type copyHomePlugin struct {
	Image      string `required:"true" json:"image"`
	VolumeName string `required:"true" json:"volumeName"`
}

func New(params map[string]string) (common.PluginInterface, error) {
	var plugin copyHomePlugin
	err := common.Parse(&plugin, params)
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (p *copyHomePlugin) GenerateInitContainerApplyConfiguration() *corev1apply.ContainerApplyConfiguration {

	// /home/coderをImageからコピーする（Volumeに存在しなければ）（.local, .configは除外）
	command := `
		sudo apt install -y rsync
		sudo rsync -ahv --progress --exclude=".local" --exclude=".config" /home/coder/ /persistent/
	`

	initcontainer := corev1apply.Container().
		WithName("copy-home").
		WithImage(p.Image).
		WithCommand("sh", "-c", command).
		WithVolumeMounts(corev1apply.VolumeMount().
			WithName(p.VolumeName).
			WithMountPath("/persistent"))

	return initcontainer
}
