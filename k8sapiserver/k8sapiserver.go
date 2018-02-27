package k8sapiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/api/meta"
	"bytes"
	"net/http"
	"podstatus/router"
	"strings"
)

type K8sApiserver struct {
	K8sApi *kubernetes.Clientset
	Host   string
}
type PodEvent struct {
	Reason  string
	Message string
}
type PodInfo struct {
	Namespace  string            `json:"namespace,omitempty"`
	CreateTime unversioned.Time  `json:"createtime,omitempty"`
	PodName    string            `json:"podname,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	NodeIp     string            `json:"hostip,omitempty"`
	PodIp      string            `json:"podip,omitempty"`
	Status     string            `json:"status,omitempty"`
	Version    string            `json:"status,omitempty"`
}

func NewClient(kubeconfig, host string) *K8sApiserver {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	k8sapiserver := K8sApiserver{K8sApi: clientset, Host: host}
	return &k8sapiserver
}
func (K8s *K8sApiserver) K8sGetVersion(image string) string {
	version := ""
	str := strings.Split(image, "/")
	for _, value := range str[2:] {
		version = version + value
	}
	return version
}
func (k8s *K8sApiserver) K8sGetPodList(namespace, servername string) ([]PodInfo, error) {

	podinfolist := make([]PodInfo, 0)
	opt := v1.ListOptions{LabelSelector: servername}
	podlist, err := k8s.K8sApi.CoreV1().Pods(namespace).List(opt)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if len(podlist.Items) == 0 {
		fmt.Println("podlist is nil")
	}
	for _, pod := range podlist.Items {
		podinfo := new(PodInfo)

		podinfo.Namespace = namespace
		podinfo.CreateTime = pod.ObjectMeta.CreationTimestamp
		podinfo.PodName = pod.ObjectMeta.Name
		podinfo.Labels = pod.ObjectMeta.Labels
		podinfo.NodeIp = pod.Status.HostIP
		podinfo.PodIp = pod.Status.PodIP
		podinfo.Version = k8s.K8sGetVersion(pod.Spec.Containers[0].Image)

		status := string(pod.Status.Phase)

		for _, containerstatus := range pod.Status.ContainerStatuses {

			if containerstatus.State.Waiting != nil {
				status = containerstatus.State.Waiting.Reason
				break
			}
			if containerstatus.State.Terminated != nil {
				status = containerstatus.State.Terminated.Reason
			}
			if containerstatus.Ready == false {
				status = "NotReady"
				break
			}
		}
		podinfo.Status = string(status)
		podinfolist = append(podinfolist, *podinfo)
	}
	return podinfolist, nil
}
func (k8s *K8sApiserver) K8sGetLog(namespace, podname, containername string) (string, error) {
	url := k8s.Host + "/api/v1/namespaces/" + namespace + "/pods/" + podname + "/log" + "?container=" + containername
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), err
}
func (k8s *K8sApiserver) K8sGetDescribe(namespace, podname string, servername string) (string, error) {
	opt := v1.ListOptions{LabelSelector: servername}
	listpodevent, err := k8s.K8sApi.CoreV1().Events(namespace).List(opt)

	if err != nil {
		return "", err
	}
	var message string
	for _, podevent := range listpodevent.Items {
		if podevent.InvolvedObject.Name == podname {
			message = message + "\n" + podevent.Message
		}
	}
	return message, nil
}
func (k8s *K8sApiserver) UpdateJettyOffline(namespace, podname string) error {
	pod, err := k8s.K8sApi.CoreV1().Pods(namespace).Get(podname)

	if err != nil {
		return err
	} else {
		if pod.ObjectMeta.Labels["app"] == "jetty" {

			url := k8s.Host + "/api/v1/namespaces/" + namespace + "/pods/" + podname
			client := &http.Client{}

			pod.ObjectMeta.Labels["offline"] = "1"
			data, _ := json.Marshal(pod)

			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))

			if err != nil {
				return err
			}
			resp, err := client.Do(req)

			defer resp.Body.Close()
			if err != nil {
				return err
			}
			/*var oldpod api.Pod
			data, _ := json.Marshal(pod)
			json.Unmarshal(data, &oldpod)
			accessor, err := meta.Accessor(&oldpod)
			if err != nil {
				return err
			}
			objLabels := accessor.GetLabels()

			fmt.Println(objLabels)

			if objLabels == nil {
				objLabels = make(map[string]string)
			}
			if pod.ObjectMeta.Labels["offline"] == "" {
				err := errors.New("not offline type... ")
				return err
			}
			objLabels["offline"] = "1"
			fmt.Println(objLabels)
			accessor.SetLabels(objLabels)*/
			return nil
		} else {
			err := errors.New("not jetty type...")
			return err
		}
	}
}

/*func (k8s *K8sApiserver) GetDeployments(namespace, servername string) ([]v1beta1.Deployment, error) {
	opt := v1.ListOptions{LabelSelector: servername}
	deployments, err := k8s.K8sApi.ExtensionsV1beta1().Deployments(namespace).List(opt)

	return deployments.Items, err
}*/
/*func (k8s *K8sApiserver) GetServices(namespace, servername string) ([]v1.Service, error) {
	opt := v1.ListOptions{LabelSelector: servername}
	listservices, err := k8s.K8sApi.CoreV1().Services(namespace).List(opt)

	return listservices, err
}*/
func (k8s *K8sApiserver) GetPods(podip, nodeip string) ([]PodInfo, error) {
	if podip == "" && nodeip == "" {
		err := fmt.Errorf("%s", "podip and nodeip is nil...")
		return nil, err
	}
	podinfolist := make([]PodInfo, 0)
	opt := v1.ListOptions{}
	pods, err := k8s.K8sApi.CoreV1().Pods("").List(opt)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if len(pods.Items) == 0 {
		fmt.Println("pods is nil ...")
	}

	for _, pod := range pods.Items {
		if podip != "" && podip != pod.Status.PodIP {
			continue
		} else if nodeip != "" && nodeip != pod.Status.HostIP {
			continue
		}

		podinfo := new(PodInfo)

		podinfo.Namespace = pod.ObjectMeta.Namespace
		podinfo.CreateTime = pod.ObjectMeta.CreationTimestamp
		podinfo.PodName = pod.ObjectMeta.Name
		podinfo.NodeIp = pod.Status.HostIP
		podinfo.PodIp = pod.Status.PodIP
		podinfo.Version = k8s.K8sGetVersion(pod.Spec.Containers[0].Image)

		status := string(pod.Status.Phase)

		for _, containerstatus := range pod.Status.ContainerStatuses {
			if containerstatus.State.Waiting != nil {
				status = containerstatus.State.Waiting.Reason
				break
			}
			if containerstatus.State.Terminated != nil {
				status = containerstatus.State.Terminated.Reason
			}
			if containerstatus.Ready == false {
				status = "NotReady"
				break
			}
		}
		podinfo.Status = string(status)
		podinfolist = append(podinfolist, *podinfo)
	}
	return podinfolist, nil
}
func (k8s *K8sApiserver) Registry() {
	ws := router.Get()
	ws.Route(ws.GET("/namespaces/{namespace}").To(k8s.getstatus))
	ws.Route(ws.GET("/namespaces/{namespace}/podname/{podname}/containername/{containername}/log").To(k8s.getlog_or_describe))
	ws.Route(ws.PUT("/namespaces/{namespace}/podname/{podname}/jettyoffline").To(k8s.jetty_offline))
	ws.Route(ws.GET("/pods/hostip/{hostip}").To(k8s.getpodstohostip))
	ws.Route(ws.GET("/pods/podip/{podip}").To(k8s.getpodstopodip))
	//	ws.Route(ws.GET("/namespaces/{namespace}/deployments").To(k8s.getdeployments))
	//	ws.Route(ws.GET("/namespaces/{namespace}/services").To(k8s.getservices))
}
func (k8s *K8sApiserver) getstatus(request *restful.Request, response *restful.Response) {

	namespace := request.PathParameter("namespace")

	servername := request.QueryParameter("servername")

	/*	if err != nil {
		response.WriteHeader(http.StatusCreated)
		response.WriteEntity(err)
	}*/
	status, err := k8s.K8sGetPodList(namespace, servername)
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.AddHeader("Content-Type", "application/json")
		//		podstatus, _ := json.Marshal(status)
		response.WriteEntity(status)
	}
}
func (k8s *K8sApiserver) getlog_or_describe(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	podname := request.PathParameter("podname")
	containername := request.PathParameter("containername")
	status := request.QueryParameter("status")
	servername := request.QueryParameter("servername")

	var containerLog string
	var err error
	if status == "Pending" || status == "NotReady" || status == "ContainerCreating" {
		containerLog, err = k8s.K8sGetDescribe(namespace, podname, servername)
	} else {
		containerLog, err = k8s.K8sGetLog(namespace, podname, containername)
	}
	if err != nil {
		response.AddHeader("content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.WriteEntity(containerLog)
	}
}
func (k8s *K8sApiserver) jetty_offline(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	podname := request.PathParameter("podname")

	err := k8s.UpdateJettyOffline(namespace, podname)
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.WriteEntity("success!")
	}

}
func (k8s *K8sApiserver) getpodstohostip(request *restful.Request, response *restful.Response) {
	HostIP := request.PathParameter("hostip")

	pods, err := k8s.GetPods("", HostIP)
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.AddHeader("Content-Type", "application/json")
		response.WriteEntity(pods)
	}

}
func (k8s K8sApiserver) getpodstopodip(request *restful.Request, response *restful.Response) {
	PodIP := request.PathParameter("podip")

	pods, err := k8s.GetPods(PodIP, "")
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.AddHeader("Content-TYpe", "application/json")
		response.WriteEntity(pods)
	}
}

/*func (k8s *K8sApiserver) getdeployments(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	servername := request.QueryParameter("servername")

	deployments, err := k8s.GetDeployments(namespace, servername)
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.AddHeader("Content-Type", "application/json")
		response.WriteEntity(deployments)
	}
}
func (k8s *K8sApiserver) getservices(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	servername := request.QueryParameter("servername")

	services, err := k8s.GetServices(namespace, servername)
	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusNotFound, err.Error())
	} else {
		response.AddHeader("Content-Type", "application/json")
		response.WriteEntity(services)
	}
}*/
