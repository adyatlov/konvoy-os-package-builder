package bundle

import (
	"fmt"
	"strings"
)

func CheckAndFixBundle(p *Package, b *Bundle) *FixResult {
	res := &FixResult{Log: make([]string, 0)}
	res.AttemptsLeft = 3
	res.AddLog(fmt.Sprintf("I'm going to check if it's possible to install package \"%s\". "+
		"I'll try to update the ackage and its dependencies if it's not.", p.Name))
	res, err := SimulateInstallation(p, b, res)
	if err != nil {
		res.AddLog(fmt.Sprintf("The following error occurred during fixing the package bundle: %v\n", err))
		res.AddLog("Unfortunately, I couldn't make this package installable on this machine.")
		return res
	}
	if res.Repeat && res.AttemptsLeft > 0 {
		res.AttemptsLeft--
		res.AddLog(fmt.Sprintf("Going again. Attempts left: %d.", res.AttemptsLeft))
		return CheckAndFixBundle(p, b)
	}
	if res.Success {
		res.AddLog("SUCCESS")
	} else {
		res.AddLog("FAILED")
	}
	return res
}

func SimulateInstallation(p *Package, b *Bundle, res *FixResult) (*FixResult, error) {
	m := b.Manager
	res.AddLog("I'm going to simulate installation of the package.")
	r, err := m.CheckInstall(p)
	if err != nil {
		res.AddLog("Couldn't simulate installation due to the following error: " + err.Error())
		return res, err
	}
	switch r.Result {
	case ResultOk:
		res.AddLog("Simulated installation was successful. I'm going to download dependencies.")
		err = m.UpdateDependencies(p)
		if err != nil {
			res.AddLog("Couldn't update package dependencies due to the following error: " + err.Error())
			return res, err
		}
		res.AddLog("Package dependencies successfully updated.")
		res.Package = p
		res.Success = true
		return res, nil
	case ResultUnmetDependencies:
		res.AddLog("Cannot install the package in its current state. Reason: " +
			"the following dependencies were not met:\n" + printDependencyList(r))
		return WhenDependencyNotMet(p, b, r, res)
	case ResultNewerAlreadyInstalled:
		res.AddLog("Cannot install the package in its current state. Reason: " +
			"a newer version of the package is already installed.")
		return WhenOtherProblemsOccurred(p, b, res)
	default:
		res.AddLog("Cannot install the package in its current state.")
		return WhenOtherProblemsOccurred(p, b, res)
	}
}

func WhenOtherProblemsOccurred(p *Package, b *Bundle, res *FixResult) (*FixResult, error) {
	if !p.VersionEssential {
		res.AddLog("The version of the package is not essential, " +
			"so I'm going to replace it with the latest version.")
		return ReplaceWithLatestVersion(p, b, res)
	}
	res.AddLog("It is important to install this exact version of the package, but it's not possible. " +
		"Please try to install download the package and its dependencies manually.")
	return res, nil
}

func WhenDependencyNotMet(p *Package, b *Bundle, r InstallResult, res *FixResult) (*FixResult, error) {
	if !p.VersionEssential {
		res.AddLog("The version of the package is not essential, " +
			"so I'm going to replace it with the latest version.")
		return ReplaceWithLatestVersion(p, b, res)
	}
	res.AddLog("It is important to install this exact version of the package. I'm going to search for the " +
		"dependencies in the package bundle.")
	allPackages := make(map[string]*Package)
	for _, p := range b.Packages {
		allPackages[p.Name] = p
		for _, d := range p.Dependencies {
			allPackages[d.Name] = d
		}
	}
	var somePackagesNotFound bool
	for _, ud := range r.UnmetDependencies {
		found, ok := allPackages[ud.Name]
		if !ok {
			res.AddLog(fmt.Sprintf("Couldn't find required dependency package %s in the bundle.", ud.Name))
			somePackagesNotFound = true
		}
		p.Dependencies = append(p.Dependencies, found)
		res.AddLog(fmt.Sprintf("Package %s found and added to the dependencies.", ud.Name))
	}
	if somePackagesNotFound {
		res.AddLog("I couldn't find some dependencies in the bundle. Please try to find them manually.")
		return res, nil
	}
	res.AddLog("All the required packages were found and added to the package dependencies. I'm going to check " +
		"once again if it's possible to install the package.")
	res.Repeat = true
	res.Package = p
	return res, nil
}

func ReplaceWithLatestVersion(p *Package, b *Bundle, res *FixResult) (*FixResult, error) {
	res.AddLog("Check if it's possible to install the latest version of the package.")
	r, err := b.Manager.CheckInstallLatestVersion(p.Name)
	if err != nil {
		res.AddLog("Couldn't check if it's possible to install the latest version due to the following error: " +
			err.Error())
		return res, err
	}
	if r != ResultOk {
		res.AddLog("It's not possible to install the latest version of the package. Please, contact support and " +
			"provide them this output.")
	}
	res.AddLog("It is possible to install the latest version of the package. I'm going to download the package " +
		"and its dependencies.")
	newPackage, err := b.Manager.DownloadLatestVersion(p.Name)
	if err != nil {
		res.AddLog("Couldn't download the latest version of the package or its dependencies " +
			"due to the following error: " + err.Error())
		return res, err
	}
	res.Success = true
	res.Package = newPackage
	return res, nil
}

func printDependencyList(r InstallResult) string {
	deps := make([]string, len(r.UnmetDependencies))
	for i, d := range r.UnmetDependencies {
		deps[i] = fmt.Sprintf("%s (>= %s)", d.Name, d.Version)
	}
	return strings.Join(deps, "\n")
}

type FixResult struct {
	Log          []string
	Success      bool
	Package      *Package
	Repeat       bool
	AttemptsLeft int
}

func (r *FixResult) AddLog(l string) {
	r.Log = append(r.Log, l)
}
