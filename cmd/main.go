package main

import (
	"context"
	"flag"
	"helm-api-version-update/pkg"
	"helm-api-version-update/pkg/cfg"
	"log"
	"os"
	"path/filepath"
)

const exitCode = 1

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	onlyFind := flag.Bool("only-find", false, "(optional) only display config maps")
	ns := flag.String("ns", "", "(optional) namespaces separated by comma")
	helm2 := flag.Bool("helm2", false, "run only for helm3 releases")
	helm3 := flag.Bool("helm3", false, "run only for helm2 releases")
	dryRun := flag.Bool("dry-run", false, "(optional) don't update")
	filter := flag.String("filter", "", "(optional) filter for releases")

	flag.Parse()

	ctx := context.Background()

	conf := cfg.Config{
		KubeCfg:  *kubeconfig,
		OnlyFind: *onlyFind,
		Ns:       *ns,
		Helm2:    *helm2,
		Helm3:    *helm3,
		DryRun:   *dryRun,
		Filter:   *filter,
	}

	if err := conf.Validate(); err != nil {
		log.Println(err)
		os.Exit(exitCode)
	}

	log.Printf("conf = %+v\n", conf)

	if err := pkg.Run(ctx, conf); err != nil {
		panic(err)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}

	return os.Getenv("USERPROFILE") // windows
}
