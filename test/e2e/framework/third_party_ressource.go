package framework

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/pkg/util/yaml"
	"time"
)

func CreateAndWaitForThirdPartyRessource(kubeClient kubernetes.Interface, relativePath string, apiPath string) error {
	tpr, err := parseTPRYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.Extensions().ThirdPartyResources().Create(tpr)
	if err != nil {
		return err
	}

	if err := WaitForThridPartyRessource(kubeClient, apiPath); err != nil {
		return err
	}

	return nil
}

func parseTPRYaml(relativePath string) (*v1beta1.ThirdPartyResource, error) {
	manifest, err := PathToOSFile(relativePath)
	if err != nil {
		return nil, err
	}

	appVersion := v1beta1.ThirdPartyResource{}
	if err := yaml.NewYAMLOrJSONDecoder(manifest, 100).Decode(&appVersion); err != nil {
		return nil, err
	}

	return &appVersion, nil
}

func WaitForThridPartyRessource(kubeClient kubernetes.Interface, apiPath string) error {
	return wait.Poll(time.Second, time.Minute, func() (bool, error) {
		res := kubeClient.CoreV1().RESTClient().Get().AbsPath(apiPath).Do()

		if res.Error() != nil {
			return false, nil
		}

		return true, nil
	})
}

func DeleteThirdPartyResource(kubeClient kubernetes.Interface, relativePath string) error {
	tpr, err := parseTPRYaml(relativePath)
	if err != nil {
		return err
	}

	if err := kubeClient.Extensions().ThirdPartyResources().Delete(tpr.Name, &v1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
