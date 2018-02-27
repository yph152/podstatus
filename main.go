package main

import (
	"flag"
	"podstatus/k8sapiserver"
	"podstatus/router"
)

var Prefix string = "/podlist/v1"

var kube_path string
var host string

func init() {
	flag.StringVar(&kube_path, "kube_path", "/root/.kube/kubeconfig", "k8s apiserver kubeconfig path..")
	flag.StringVar(&host, "host", "http://127.0.0.1:8080", "k8s apiserver host apiserver address and port")
}
func main() {
	flag.Parse()
	router.InitRouterService(Prefix)

	k8sclient := k8sapiserver.NewClient(kube_path, host)
	k8sclient.Registry()

	router.Run()
}
