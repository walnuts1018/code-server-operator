package initplugins

import (
	"errors"

	"github.com/walnuts1018/code-server-operator/internal/initplugins/common"
	"github.com/walnuts1018/code-server-operator/internal/initplugins/copydefaultconfig"
	"github.com/walnuts1018/code-server-operator/internal/initplugins/copyhomeplugin"
	"github.com/walnuts1018/code-server-operator/internal/initplugins/gitplugin"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

var ErrNotFound = errors.New("plugin not found")

var plugins = map[string]func(params map[string]string) (common.PluginInterface, error){
	"git": func(params map[string]string) (common.PluginInterface, error) {
		return gitplugin.New(params)
	},
	"copyDefaultConfig": func(params map[string]string) (common.PluginInterface, error) {
		return copydefaultconfig.New(params)
	},
	"copyHome": func(params map[string]string) (common.PluginInterface, error) {
		return copyhomeplugin.New(params)
	},
}

func CreatePlugin(initpluginConfig map[string]map[string]string, commonParams common.CommonFields) ([]*corev1apply.ContainerApplyConfiguration, error) {
	containers := make([]*corev1apply.ContainerApplyConfiguration, 0, len(initpluginConfig))

	for name, parameters := range initpluginConfig {
		if parameters == nil {
			parameters = make(map[string]string)
		}
		parameters["image"] = commonParams.Image
		parameters["volumeName"] = commonParams.VolumeName

		plugin := plugins[name]
		if plugin == nil {
			return []*corev1apply.ContainerApplyConfiguration{}, ErrNotFound
		}
		p, err := plugin(parameters)
		if err != nil {
			return []*corev1apply.ContainerApplyConfiguration{}, err
		}
		containers = append(containers, p.GenerateInitContainerApplyConfiguration())
	}
	return containers, nil
}
