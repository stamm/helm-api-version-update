package cfg

import "fmt"

// Config for program
type Config struct {
	KubeCfg  string
	Ns       string
	Filter   string
	OnlyFind bool
	Helm2    bool
	Helm3    bool
	DryRun   bool
}

// Validate config
func (c Config) Validate() error {
	if c.Helm2 && c.Helm3 {
		return fmt.Errorf("can't use both helm 2 and 3")
	}

	if !c.Helm2 && !c.Helm3 {
		return fmt.Errorf("must use one of helm 2 or 3")
	}

	return nil
}
