package bundle

import (
	"fmt"
	"io/fs"
	"path"
)

type Bundle struct {
	Manager  PackageManager
	Packages []*Package
}

func NewBundle(fileSystem fs.FS, manager PackageManager) (*Bundle, error) {
	b := &Bundle{Manager: manager}
	entries, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, fmt.Errorf("cannot read from bundle. Error: %w", err)
	}
	b.Packages = make([]*Package, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		p, err := newPackageDir(fileSystem, entry.Name(), manager)
		if err != nil {
			return nil, fmt.Errorf("cannot create package from dir %s. Error: %w", entry.Name(), err)
		}
		b.Packages = append(b.Packages, p)
	}
	return b, nil
}

type Package struct {
	NameVersion
	Path             string
	Dependencies     []*Package
	VersionEssential bool
	fileSystem       fs.FS
	manager          PackageManager
}

type NameVersion struct {
	Name    string
	Version string
}

func NewPackage(fileSystem fs.FS, packagePath string, manager PackageManager) (*Package, error) {
	fileOrDir, err := fs.Stat(fileSystem, packagePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat \"%s\" to open package. Error: %w", packagePath, err)
	}
	if fileOrDir.IsDir() {
		return newPackageDir(fileSystem, packagePath, manager)
	}
	return newPackageFile(fileSystem, packagePath, manager)
}

func newPackageFile(fileSystem fs.FS, packageFilePath string, manager PackageManager) (*Package, error) {
	p := &Package{Path: packageFilePath, fileSystem: fileSystem, manager: manager}
	packageFileName := path.Base(packageFilePath)
	var err error
	p.NameVersion, err = manager.ParseNameVersion(packageFileName)
	if err != nil {
		return nil, fmt.Errorf("cannot parse package name and version %s. Error: %w", packageFileName, err)
	}
	return p, nil
}
func newPackageDir(fileSystem fs.FS, packageDirPath string, manager PackageManager) (*Package, error) {
	entries, err := fs.ReadDir(fileSystem, packageDirPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get list of files from package directory %s. Error: %w", packageDirPath, err)
	}
	var mainPackage *Package
	dependencies := make([]*Package, 0, len(entries)-1)
	packageDirName := path.Base(packageDirPath)
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Do not second level
		}
		packageFileName := entry.Name()
		p, err := newPackageFile(fileSystem, path.Join(packageDirPath, packageFileName), manager)
		if err != nil {
			return nil, fmt.Errorf("cannot create package. Error: %w", err)
		}
		if manager.IsMain(packageDirName, packageFileName) {
			if mainPackage != nil {
				return nil, fmt.Errorf("more than one candidate for the main package %s found", packageDirName)
			}
			mainPackage = p
			nv, err := manager.ParseNameVersion(packageDirName)
			if err != nil {
				return nil, fmt.Errorf("cannot detect if the package version is essential. Error: %w", err)
			}
			// When the package version is not essential, the package directory does not contain version.
			mainPackage.VersionEssential = nv.Version != ""
			continue
		}
		dependencies = append(dependencies, p)
	}
	if mainPackage == nil {
		return nil, fmt.Errorf("no candidates for the main package %s found", packageDirPath)
	}
	mainPackage.Dependencies = dependencies
	return mainPackage, nil
}

func (p *Package) Open() (fs.File, error) {
	return p.fileSystem.Open(p.Path)
}
