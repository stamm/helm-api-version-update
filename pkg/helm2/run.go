package helm2

import (
	"context"
	"errors"
	"fmt"
	"helm-api-version-update/pkg/cfg"
	"helm-api-version-update/pkg/common"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

// Run convert api version
func Run(ctx context.Context, clientset *kubernetes.Clientset, namespaces []string, conf cfg.Config) error {
	configMaps, err := clientset.CoreV1().ConfigMaps("kube-system").
		List(metav1.ListOptions{LabelSelector: "OWNER=TILLER,STATUS=DEPLOYED"})
	if err != nil {
		return fmt.Errorf("can't get configmaps: %w", err)
	}

	log.Printf("len(configMaps) = %+v\n", len(configMaps.Items))

	for _, configMap := range configMaps.Items {
		if conf.Filter != "" && !strings.Contains(configMap.Name, conf.Filter) {
			continue
		}

		rls, err := getHelmRelease(configMap)
		if err != nil {
			log.Printf("configMap error getting release%s: %s", configMap.Name, common.ErrNotRelease)
			continue
		}

		if len(namespaces) > 1 && skip(rls, namespaces) {
			continue
		}

		if conf.OnlyFind {
			rules, err := common.ContainsRulesConfigMap(configMap)
			if err != nil {
				log.Printf("error on ContainsRules %s: %s", configMap.Name, err)
			}

			if len(rules) > 0 {
				log.Printf("%s have rules: %+v", configMap.Name, rules)
			}

			continue
		}

		newConfigMap, err := ConvertConfigMap(configMap, rls)
		if err != nil {
			if errors.Is(err, common.ErrNotRelease) || errors.Is(err, common.ErrNothingUpdate) {
				log.Printf("skip: %s", err)

				continue
			}

			return fmt.Errorf("can't decode %s: %w", configMap.Name, err)
		}

		if conf.DryRun {
			log.Printf("skip dry-run update configMap %s\n", newConfigMap.Name)

			continue
		}

		_, err = clientset.CoreV1().ConfigMaps("kube-system").Update(&newConfigMap)
		if err != nil {
			return fmt.Errorf("can't update configMap %s: %w", configMap.Name, err)
		}

		log.Printf("updated configMap %s\n", newConfigMap.Name)
	}

	return nil
}

// ConvertConfigMap secret
func ConvertConfigMap(configMap corev1.ConfigMap, rls rspb.Release) (corev1.ConfigMap, error) {
	isChanged := false

	for from, to := range common.GetRules() {
		ok := strings.Contains(rls.Manifest, from)
		ok2 := false

		for _, tmpl := range rls.Chart.Templates {
			if strings.Contains(string(tmpl.Data), from) {
				ok2 = true
				break
			}
		}

		if !ok && !ok2 {
			continue
		}

		log.Printf("%s %s", from, to)
		rls.Manifest = strings.Replace(rls.Manifest, from, to, -1)

		for i, tmpl := range rls.Chart.Templates {
			rls.Chart.Templates[i].Data = []byte(strings.Replace(string(tmpl.Data), from, to, -1))
		}

		b, err := proto.Marshal(&rls)
		if err != nil {
			return corev1.ConfigMap{}, err
		}

		enFixed, err := common.Encode(string(b))
		if err != nil {
			return corev1.ConfigMap{}, fmt.Errorf("can't encode %s: %w", configMap.Name, err)
		}

		configMap.Data["release"] = enFixed
		isChanged = true
	}

	if !isChanged {
		return corev1.ConfigMap{}, fmt.Errorf("configMap %s: %w", configMap.Name, common.ErrNothingUpdate)
	}

	return configMap, nil
}

func getHelmRelease(configMap corev1.ConfigMap) (rspb.Release, error) {
	data, ok := configMap.Data["release"]
	if !ok {
		return rspb.Release{}, fmt.Errorf("configMap %s: %w", configMap.Name, common.ErrNotRelease)
	}
	decoded, err := common.Decode(data)
	if err != nil {
		return rspb.Release{}, fmt.Errorf("can't decode ConfigMap %s: %w", configMap.Name, err)
	}

	var rls rspb.Release

	// unmarshal protobuf bytes
	if err := proto.Unmarshal([]byte(decoded), &rls); err != nil {
		return rspb.Release{}, fmt.Errorf("cant unmarshal %s: %w", configMap.Name, err)
	}

	return rls, nil
}

func skip(rls rspb.Release, namespaces []string) bool {
	for _, ns := range namespaces {
		if rls.Namespace == ns {
			return false
		}
	}

	return true
}
