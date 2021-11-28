package main

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"

	"konvoy-os-package-builder/bundle"
	"konvoy-os-package-builder/pkg/apt"

	"github.com/disiqueira/gotree"
	"github.com/nlepage/go-tarfs"
)

func main() {
	f, err := os.Open("backup_konvoy_v1.8.3_amd64_debs.tar.gz")
	must(err)
	gzr, err := gzip.NewReader(f)
	must(err)
	fileSystem, err := tarfs.New(gzr)
	must(err)
	m, err := apt.NewManager()
	must(err)
	b, err := bundle.NewBundle(fileSystem, m)
	must(err)
	tree := PrintBundleTree(b)
	fmt.Println(tree)

	p := b.Packages[5]
	r, err := m.CheckInstall(p)
	must(err)
	fmt.Println(r.Result, r.Package.Name, r.Package.Version, r.Package.Dependencies)
	fmt.Println("Dependencies before update:")
	for _, d := range p.Dependencies {
		fmt.Println(d.Name, d.Version)
	}
	err = m.UpdateDependencies(p)
	must(err)
	fmt.Println("Dependencies after update:")
	for _, d := range p.Dependencies {
		fmt.Println(d.Name, d.Version)
	}
	r, err = m.CheckInstall(p)
	must(err)
	fmt.Println(r.Result, r.Package.Name, r.Package.Version, r.Package.Dependencies)

	p, err = m.DownloadLatestVersion("nfs-common")
	must(err)
	fmt.Println(PrintPackageTree(p))

	rt, err := m.CheckInstallLatestVersion("kubeadm")
	must(err)
	fmt.Println(rt)
}

func PrintBundleTree(b *bundle.Bundle) string {
	bundleNode := gotree.New("/")
	for _, p := range b.Packages {
		packageNode := bundleNode.Add(fmt.Sprintf("%s - v%s", p.Name, p.Version))
		if len(p.Dependencies) == 0 {
			packageNode.Add("No dependencies")
			continue
		}
		for _, d := range p.Dependencies {
			packageNode.Add(fmt.Sprintf("%s - v%s", d.Name, d.Version))
		}
	}
	return bundleNode.Print()
}

func PrintPackageTree(p *bundle.Package) string {
	packageNode := gotree.New(fmt.Sprintf("%s - v%s", p.Name, p.Version))
	for _, d := range p.Dependencies {
		packageNode.Add(fmt.Sprintf("%s - v%s", d.Name, d.Version))
	}
	return packageNode.Print()
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
