package main

import (
	"compress/gzip"
	"log"
	"os"

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
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
