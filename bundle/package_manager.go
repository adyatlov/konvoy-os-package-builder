package bundle

type InstallResultType int

const (
	ResultOk InstallResultType = iota
	ResultAlreadyInstalled
	ResultUnmetDependencies
	ResultNewerAlreadyInstalled
	ResultCannotFindPackage
	ResultUnknownProblem
)

type PackageManager interface {
	Name() string
	ParseNameVersion(packageFileName string) (NameVersion, error)
	IsMain(packageDirName, packageFileName string) bool
	CheckInstall(p *Package) (InstallResult, error)
	CheckInstallLatestVersion(name string) (InstallResultType, error)
	UpdateDependencies(p *Package) error
	DownloadLatestVersion(packageName string) (*Package, error)
	Clean() error
}

type InstallResult struct {
	Result            InstallResultType
	Package           *Package
	UnmetDependencies []NameVersion
}
