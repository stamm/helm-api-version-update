package common

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ErrNotRelease error for absent data.release in secret
var ErrNotRelease = errors.New("don't have data.release")

// ErrNothingUpdate error if not found old apiVersion
var ErrNothingUpdate = errors.New("don't have old apiVersion")

// ConvertSecret secret
func ConvertSecret(secret corev1.Secret) (corev1.Secret, error) {
	data, ok := secret.Data["release"]
	if !ok {
		return corev1.Secret{}, fmt.Errorf("secret %s.%s: %w", secret.Name, secret.Namespace, ErrNotRelease)
	}
	// fmt.Printf("string(data) = %+v\n", string(data))
	decoded, err := Decode(string(data))
	if err != nil {
		return corev1.Secret{}, fmt.Errorf("can't decode %s %s: %w", secret.Namespace, secret.Name, err)
	}

	secret.StringData = make(map[string]string, 1)
	isChanged := false

	for from, to := range GetRules() {
		ok = strings.Contains(decoded, from)
		if !ok {
			continue
		}

		fixed := strings.Replace(decoded, from, to, -1)

		enFixed, err := Encode(fixed)
		if err != nil {
			return corev1.Secret{}, fmt.Errorf("can't encode %s %s: %w", secret.Namespace, secret.Name, err)
		}

		secret.StringData["release"] = enFixed
		isChanged = true
	}

	if !isChanged {
		return corev1.Secret{}, fmt.Errorf("secret %s.%s: %w", secret.Name, secret.Namespace, ErrNotRelease)
	}

	return secret, nil
}

// ContainsRulesSecret secret
func ContainsRulesSecret(secret corev1.Secret) ([]string, error) {
	return containsRules(secret.Name, secret.Namespace, string(secret.Data["release"]))
}

// ContainsRulesConfigMap secret
func ContainsRulesConfigMap(configMap corev1.ConfigMap) ([]string, error) {
	return containsRules(configMap.Name, configMap.Namespace, configMap.Data["release"])
}

func containsRules(name, ns string, data string) ([]string, error) {
	if data == "" {
		return []string{}, fmt.Errorf("dont have release in data %s.%s: %w", name, ns, ErrNotRelease)
	}
	// fmt.Printf("string(data) = %+v\n", string(data))
	decoded, err := Decode(data)
	if err != nil {
		return []string{}, fmt.Errorf("can't decode %s %s: %w", ns, name, err)
	}
	// fmt.Printf("decoded = %s\n", decoded)

	result := make([]string, 0)

	for from := range GetRules() {
		ok := strings.Contains(decoded, from)
		if !ok {
			continue
		}

		result = append(result, from)
	}

	return result, nil
}

func GetRules() map[string]string {
	return map[string]string{
		"extensions/v1beta1\\nkind: Deployment": "apps/v1\\nkind: Deployment",
		"apps/v1beta1\\nkind: Deployment":       "apps/v1\\nkind: Deployment",
		"apps/v1beta2\\nkind: Deployment":       "apps/v1\\nkind: Deployment",
		"extensions/v1beta1\\nkind: Ingress":    "networking.k8s.io/v1beta1\\nkind: Ingress",
		"extensions/v1beta1\\nkind: DaemonSet":  "apps/v1\\nkind: DaemonSet",

		"extensions/v1beta1\nkind: Deployment": "apps/v1\nkind: Deployment",
		"apps/v1beta1\nkind: Deployment":       "apps/v1\nkind: Deployment",
		"apps/v1beta2\nkind: Deployment":       "apps/v1\nkind: Deployment",
		"extensions/v1beta1\nkind: Ingress":    "networking.k8s.io/v1beta1\nkind: Ingress",
		"extensions/v1beta1\nkind: DaemonSet":  "apps/v1\nkind: DaemonSet",
		// "apps/v1\nkind: Deployment": "apps/v1beta2\nkind: Deployment",
	}
}
