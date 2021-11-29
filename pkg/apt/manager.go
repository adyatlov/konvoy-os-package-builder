package apt

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"konvoy-os-package-builder/bundle"
)

const aptCachePath = "/var/cache/apt/archives/"

var _ bundle.PackageManager = &Manager{}

type Manager struct {
	tmpDir string
}

func NewManager() (*Manager, error) {
	m := &Manager{}
	var err error
	m.tmpDir, err = os.MkdirTemp("", "konvoy-os-package-builder-*")
	if err != nil {
		return nil, fmt.Errorf("cannot create a temporary directory for APT package manager. Error: %w", err)
	}
	return m, nil
}

func (m *Manager) Name() string {
	return "apt"
}

func (m *Manager) ParseNameVersion(packageFileName string) (bundle.NameVersion, error) {
	// Fix package directory name
	packageFileName = strings.Replace(packageFileName, "=", "_", -1)
	parts := strings.Split(packageFileName, "_")
	if len(parts) == 1 {
		return bundle.NameVersion{Name: parts[0]}, nil
	}
	return bundle.NameVersion{Name: parts[0], Version: parts[1]}, nil
}

func (m *Manager) IsMain(packageDirName, packageFileName string) bool {
	// Fix package directory name
	packageDirName = strings.Replace(packageDirName, "=", "_", -1)
	return strings.HasPrefix(packageFileName, packageDirName)
}

func (m *Manager) CheckInstall(p *bundle.Package) (bundle.InstallResult, error) {
	res := bundle.InstallResult{Package: p}
	packageTmpDir, err := os.MkdirTemp(m.tmpDir, fmt.Sprintf("CheckInstall-%s-%s-*", p.Name, p.Version))
	if err != nil {
		return res, fmt.Errorf("cannot create temporary directory for to install package %s. Error: %w",
			p.Path, err)
	}
	if err = extractPackage(p, packageTmpDir); err != nil {
		return res, fmt.Errorf("cannot copy package %s to %s. Error: %w", p.Path, packageTmpDir, err)
	}
	cmd := exec.Command("sh", "-c", "apt-get -s install -y "+path.Join(packageTmpDir, "*"))
	msg, err := cmd.CombinedOutput()
	if err == nil {
		res.Result = bundle.ResultOk
		return res, nil
	}
	if _, ok := err.(*exec.ExitError); !ok {
		return res, fmt.Errorf("cannot launch apt-get command. Error: %w", err)
	}
	res.Result = parseResultType(string(msg))
	if res.Result != bundle.ResultUnmetDependencies {
		return res, nil
	}
	res.UnmetDependencies = parseDependencies(string(msg))
	fmt.Println("=================================================\n", res.UnmetDependencies, string(msg))
	return res, nil
}

func (m *Manager) CheckInstallLatestVersion(name string) (bundle.InstallResultType, error) {
	cmd := exec.Command("sh", "-c", "apt-get -s install -y "+name)
	msg, err := cmd.CombinedOutput()
	if err == nil {
		return bundle.ResultOk, nil
	}
	if _, ok := err.(*exec.ExitError); !ok {
		return -1, fmt.Errorf("cannot launch apt-get command. Error: %w", err)
	}
	return parseResultType(string(msg)), nil
}

func (m *Manager) UpdateDependencies(p *bundle.Package) error {
	if err := clearAPTCache(); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(m.tmpDir, fmt.Sprintf("UpdateDependencies-%s-%s-*", p.Name, p.Version))
	if err != nil {
		return fmt.Errorf("cannot create temporary directory for extracting package %s. Error: %w",
			p.Path, err)
	}
	if err = extractPackage(p, tmpDir); err != nil {
		return fmt.Errorf("cannot extraact package %s to %s. Error: %w", p.Path, tmpDir, err)
	}
	cmd := exec.Command("sh", "-c", "apt-get install -d -y --reinstall "+path.Join(tmpDir, "*"))
	msg, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("cannot launch apt-get command. Error: %w", err)
		}
		return fmt.Errorf("cannot download package %s with apt-get install -d. Command ouput:\n%s",
			p.Path, string(msg))
	}
	// When you launch the command above, APT downloads dependencies to its cache.
	// Add those dependencies to the temporary package dir and create new dependencies.
	downloadedDependenciesDir := path.Join(tmpDir, "downloaded_dependencies")
	if err = os.Mkdir(downloadedDependenciesDir, 0700); err != nil {
		return fmt.Errorf("cannot create directory for downloaded dependencies. Error: %w", err)
	}
	if err := copyDebFilesFromAptCache(downloadedDependenciesDir); err != nil {
		return err
	}
	fileSystem := os.DirFS(downloadedDependenciesDir)
	debFiles, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return fmt.Errorf("cannot read directory %s. Error: %w", downloadedDependenciesDir, err)
	}
	for _, f := range debFiles {
		dep, err := bundle.NewPackage(fileSystem, f.Name(), m)
		if err != nil {
			return fmt.Errorf("cannot create downloaded package %s. Error: %w", f.Name(), err)
		}
		p.Dependencies = append(p.Dependencies, dep)
	}
	return nil
}

func (m *Manager) DownloadLatestVersion(name string) (*bundle.Package, error) {
	if err := clearAPTCache(); err != nil {
		return nil, err
	}
	cmd := exec.Command("sh", "-c", "apt-get install -d -y --reinstall "+name)
	msg, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("cannot launch apt-get command. Error: %w", err)
		}
		return nil, fmt.Errorf("cannot download package %s with apt-get install -d. Command ouput:\n%s",
			name, string(msg))
	}
	tmpDir, err := os.MkdirTemp(m.tmpDir, fmt.Sprintf("DownloadLatestVersion-%s-*", name))
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary directory for to download package %s. Error: %w",
			name, err)
	}
	packageDir := path.Join(tmpDir, name)
	if err = os.Mkdir(packageDir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create package dir %s. Error: %w", packageDir, err)
	}
	if err := copyDebFilesFromAptCache(packageDir); err != nil {
		return nil, err
	}
	fileSystem := os.DirFS(tmpDir)
	newP, err := bundle.NewPackage(fileSystem, path.Base(packageDir), m)
	if err != nil {
		return nil, fmt.Errorf("cannot create downloaded package. Error: %w", err)
	}
	return newP, nil
}

func (m *Manager) Clean() error {
	if err := os.RemoveAll(m.tmpDir); err != nil {
		return fmt.Errorf("cannot remove temporary directory %s. Error: %w", m.tmpDir, err)
	}
	return nil
}

func parseResultType(msg string) bundle.InstallResultType {
	msg = strings.ToLower(msg)
	if strings.Contains(msg, "will be downgraded") {
		return bundle.ResultNewerAlreadyInstalled
	}
	if strings.Contains(msg, "unmet dependencies") {
		return bundle.ResultUnmetDependencies
	}
	if strings.Contains(msg, "unable to locate package") {
		return bundle.ResultCannotFindPackage
	}
	return bundle.ResultUnknownProblem
}

var depReg = regexp.MustCompile(`Depends:\s*(\S+)\s+(?:\(>=\s*(\d+\.\d+\.\d+)\))?`)

func parseDependencies(msg string) []bundle.NameVersion {
	dd := depReg.FindAllStringSubmatch(msg, -1)
	deps := make([]bundle.NameVersion, len(dd))
	for i, d := range dd {
		deps[i].Name = d[1]
		if len(d) > 2 {
			deps[i].Version = d[2]
		}
	}
	return deps
}

func clearAPTCache() error {
	if err := os.RemoveAll(aptCachePath); err != nil {
		return fmt.Errorf("cannot remove %s. Error: %w", aptCachePath, err)
	}
	return nil
}

func extractPackage(p *bundle.Package, dir string) error {
	from, err := p.Open()
	if err != nil {
		return fmt.Errorf("cannot open package %s. Error: %w", p.Path, err)
	}
	//noinspection GoUnhandledErrorResult
	defer from.Close()
	toFileName := path.Join(dir, path.Base(p.Path))
	to, err := os.Create(toFileName)
	if err != nil {
		return fmt.Errorf("cannot create a file %s. Error: %w", toFileName, err)
	}
	//noinspection GoUnhandledErrorResult
	defer to.Close()
	if _, err := io.Copy(to, from); err != nil {
		return fmt.Errorf("cannot copy package %s to %s. Error: %w", p.Path, dir, err)
	}
	for _, dep := range p.Dependencies {
		if err := extractPackage(dep, dir); err != nil {
			return fmt.Errorf("cannot copy dependency %s of package %s to %s. Error: %w", dep.Path, p.Path, dir, err)
		}
	}
	return nil
}

func copyDebFilesFromAptCache(destDirPath string) error {
	entries, err := os.ReadDir(aptCachePath)
	if err != nil {
		return fmt.Errorf("cannot read dir %s. Error: %w", aptCachePath, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".deb" {
			continue
		}
		filePath := path.Join(aptCachePath, entry.Name())
		if err = copyFile(filePath, destDirPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(filePath, destDirPath string) error {
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file %s. Error: %w", filePath, err)
	}
	//noinspection GoUnhandledErrorResult
	defer r.Close()
	fileName := path.Base(filePath)
	destFilePath := path.Join(destDirPath, fileName)
	w, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("cannot create file %s. Error: %w", destFilePath, err)
	}
	//noinspection GoUnhandledErrorResult
	defer w.Close()
	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("cannot copy file %s to %s. Error: %w", filePath, destFilePath, err)
	}
	return nil
}
