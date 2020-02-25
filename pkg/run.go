package pkg

import (
	"context"
	"fmt"
	"helm-api-version-update/pkg/cfg"
	"helm-api-version-update/pkg/helm2"
	"helm-api-version-update/pkg/helm3"
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// init oidc plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Run convert api version
func Run(ctx context.Context, cfg cfg.Config) error {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeCfg)
	if err != nil {
		return fmt.Errorf("can't build config from file (%s): %w", cfg.KubeCfg, err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("can't build client: %w", err)
	}

	namespaces, err := getNamespaces(clientset, cfg.Ns)
	if err != nil {
		return fmt.Errorf("can't get namespaces: %w", err)
	}

	log.Printf("namespaces = %+v\n", namespaces)

	if cfg.Helm2 {
		return helm2.Run(ctx, clientset, namespaces, cfg)
	}

	if cfg.Helm3 {
		return helm3.Run(ctx, clientset, namespaces, cfg)
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
