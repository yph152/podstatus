package k8sapiserver

import (
	"fmt"
	"testing"
)

/*func Test_K8sGetStatus(t *testing.T) {
	k8s := NewClient("/root/.kube/kubeconfig")
	fmt.Println(k8s.K8sGetPodList("default", ""))

}*/
func Test_K8sGetLog(t *testing.T) {
	k8s := NewClient("/root/.kube/kubeconfig", "172.16.189.22")
	version := k8s.K8sGetVersion("registry.meizu.com/codis/test:v1")
	fmt.Println(version)
	/*	podlist, err := k8s.GetPods("", "172.16.189.30")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(podlist)
		}
	*/
}
