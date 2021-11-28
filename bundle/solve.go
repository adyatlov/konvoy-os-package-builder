package bundle

import (
	"fmt"
	"strings"
)

func CheckAndFixPackage(p *Package, b *Bundle) (*Package, *FixResult, error) {
	fixLog := make([]string, 0)
	m := b.Manager
	r, err := m.CheckInstall(p)
	fixLog = append(fixLog, "Simulate installation")
	if err != nil {
		return nil, nil, err
	}
	switch r.Result {
	case ResultOk:
		fixLog = append(fixLog, "Simulated installation was successful. No changes needed to the package.")
		return p, fixLog, nil
	case ResultUnmetDependencies:
		fixLog = append(fixLog, "Cannot install the package in it's current state. Reason: "+
			"the following dependencies were not met:\n"+printDependencyList(r))
		return WhenDependencyNotMet(p, b)
	}
	return nil, nil, nil
}

type FixResult struct {
	Log     []string
	Success bool
}

func WhenDependencyNotMet(p *Package, b *Bundle) (*Package, FixResult, error) {
	if !p.VersionEssential {

	}
}

func printDependencyList(r InstallResult) string {
	deps := make([]string, len(r.UnmetDependencies))
	for i, d := range r.UnmetDependencies {
		deps[i] = fmt.Sprintf("%s (>= %s)", d.Name, d.Version)
	}
	return strings.Join(deps, "\n")
}
