package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"konvoy-os-package-builder/bundle"
	"konvoy-os-package-builder/pkg/apt"

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
	bundle.CheckAndFixBundle(b)
	err = bundleToTarball(b, "konvoy_v1.8.3_amd64_debs.tar.gz")
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func bundleToTarball(b *bundle.Bundle, tarBallPath string) error {
	f, err := os.Create(tarBallPath)
	if err != nil {
		return fmt.Errorf("cannot create file %s. Error: %w", tarBallPath, err)
	}
	//noinspection GoUnhandledErrorResult
	defer f.Close()
	gw := gzip.NewWriter(f)
	//noinspection GoUnhandledErrorResult
	defer gw.Close()
	tw := tar.NewWriter(gw)
	//noinspection GoUnhandledErrorResult
	defer tw.Close()
	for _, p := range b.Packages {
		if err = packageToTarball(p, p.Path, tw); err != nil {
			return err
		}
		for _, d := range p.Dependencies {
			dPath := path.Join(path.Base(path.Dir(p.Path)), path.Base(d.Path))
			if err = packageToTarball(d, dPath, tw); err != nil {
				return err
			}
		}
	}
	return nil
}

func packageToTarball(p *bundle.Package, packagePath string, tw *tar.Writer) error {
	r, err := p.Open()
	if err != nil {
		return fmt.Errorf("cannot open package file. Error: %w", err)
	}
	info, err := p.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat package file. Error: %w", err)
	}
	header := &tar.Header{
		Name:    packagePath,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}
	err = tw.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("cannot write header for package %s. Error: %w", p.Path, err)
	}
	_, err = io.Copy(tw, r)
	if err != nil {
		return fmt.Errorf("cannot copy package bytes to tar. Error: %w", err)
	}
	return nil
}
