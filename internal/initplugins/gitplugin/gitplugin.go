package gitplugin

import (
	"fmt"
	"net/url"

	"github.com/walnuts1018/code-server-operator/internal/initplugins/common"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type gitPlugin struct {
	Repourl     string `required:"true" json:"repourl"`
	Branch      string `json:"branch"`
	VolumeName  string `required:"true" json:"volumeName"`
	InitCommand string `json:"initCommand"`
}

func New(params map[string]string) (common.PluginInterface, error) {
	var gitplugin gitPlugin
	err := common.Parse(&gitplugin, params)
	if err != nil {
		return nil, err
	}

	repoURL, err := url.Parse(gitplugin.Repourl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repourl: %w", err)
	}
	repoURL.Scheme = "https"
	gitplugin.Repourl = repoURL.String()

	return &gitplugin, nil
}

func (g *gitPlugin) GenerateInitContainerApplyConfiguration() *corev1apply.ContainerApplyConfiguration {
	initCommand := fmt.Sprintf("cd /persistent/work && %v", g.InitCommand)

	command := fmt.Sprintf(`
		if [ ! -d /persistent/work ]; then
			mkdir -p /persistent/work;
			git clone -b `+g.Branch+` `+g.Repourl+` /persistent/work;
			%v
		else
			echo 'work directory already exists';
		fi
	`, initCommand)

	initcontainer := corev1apply.Container().
		WithName("git").
		WithImage("alpine/git").
		WithCommand("sh", "-c", command).
		WithVolumeMounts(corev1apply.VolumeMount().
			WithName(g.VolumeName).
			WithMountPath("/persistent"))

	return initcontainer
}
