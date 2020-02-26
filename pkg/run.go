package pkg

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/stamm/helm-api-version-update/pkg/cfg"
	"github.com/stamm/helm-api-version-update/pkg/helm2"
	"github.com/stamm/helm-api-version-update/pkg/helm3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// init oidc plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Run convert api version
func Run(ctx context.Context, conf cfg.Config) error {
	config, err := clientcmd.BuildConfigFromFlags("", conf.KubeCfg)
	if err != nil {
		return fmt.Errorf("can't build config from file (%s): %w", conf.KubeCfg, err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("can't build client: %w", err)
	}

	namespaces, err := getNamespaces(clientset, conf.Ns)
	if err != nil {
		return fmt.Errorf("can't get namespaces: %w", err)
	}

	log.Printf("namespaces = %+v\n", namespaces)

	if conf.Helm2 {
		return helm2.Run(ctx, clientset, namespaces, conf)
	}

	if conf.Helm3 {
		return helm3.Run(ctx, clientset, namespaces, conf)
	}

	return nil
}

func getNamespaces(cl kubernetes.Interface, cfgNamespaces string) ([]string, error) {
	if cfgNamespaces != "" {
		return strings.Split(cfgNamespaces, ","), nil
	}

	namespaces, err := cl.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return []string{}, fmt.Errorf("can't get namespaces: %w", err)
	}

	nss := make([]string, 0, len(namespaces.Items))

	for _, namespace := range namespaces.Items {
		nss = append(nss, namespace.Name)
	}

	return nss, nil
}
