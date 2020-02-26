package helm3

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/stamm/helm-api-version-update/pkg/cfg"
	"github.com/stamm/helm-api-version-update/pkg/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// init oidc plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Run convert api version
func Run(ctx context.Context, clientset *kubernetes.Clientset, namespaces []string, conf cfg.Config) error {
	for _, ns := range namespaces {
		list, err := clientset.CoreV1().Secrets(ns).List(metav1.ListOptions{LabelSelector: "owner=helm,status=deployed"})
		if err != nil {
			return fmt.Errorf("can't get secret in ns %s: %w", ns, err)
		}

		log.Printf("ns: %s, len(list) = %+v\n", ns, len(list.Items))

		for i := range list.Items {
			secret := list.Items[i]
			if conf.Filter != "" && !strings.Contains(secret.Name, conf.Filter) {
				continue
			}

			if conf.OnlyFind {
				rules, err := common.ContainsRulesSecret(&secret)
				if err != nil {
					log.Printf("error on ContainsRuls %s.%s: %s", secret.Name, secret.Namespace, err)
				}

				if len(rules) > 0 {
					log.Printf("%s.%s have rules: %+v", secret.Name, secret.Namespace, rules)
				}

				continue
			}

			newSecret, err := common.ConvertSecret(&secret)
			if err != nil {
				if errors.Is(err, common.ErrNotRelease) || errors.Is(err, common.ErrNothingUpdate) {
					log.Printf("skip: %s", err)
					continue
				}

				return fmt.Errorf("can't decode %s %s: %w", ns, secret.Name, err)
			}

			if conf.DryRun {
				log.Printf("skip dry-run update secret %s\n", secret.Name)
				continue
			}

			if _, err = clientset.CoreV1().Secrets(ns).Update(&newSecret); err != nil {
				return fmt.Errorf("can't update secret %s: %w", ns, err)
			}

			log.Printf("updated ns %s, secret %s\n", ns, newSecret.Name)
		}
	}

	return nil
}
