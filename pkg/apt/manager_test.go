package apt

import (
	"reflect"
	"testing"

	"konvoy-os-package-builder/bundle"
)

func TestManager_IsMainPackage(t *testing.T) {
	type args struct {
		packageDirName  string
		packageFileName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "detects main package",
			args: args{
				packageDirName:  "apt-transport-https",
				packageFileName: "apt-transport-https_1.2.35_amd64.deb",
			},
			want: true,
		},
		{
			name: "detects main package with =",
			args: args{
				packageDirName:  "containerd.io=1.4.7-1",
				packageFileName: "containerd.io_1.4.7-1_amd64.deb",
			},
			want: true,
		},
		{
			name: "does not detect main package",
			args: args{
				packageDirName:  "kubeadm=1.20.11-00",
				packageFileName: "cri-tools_1.13.0-01_amd64.deb",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Manager{}
			if got := m.IsMain(tt.args.packageDirName, tt.args.packageFileName); got != tt.want {
				t.Errorf("IsMain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDependencies(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want []bundle.NameVersion
	}{
		{
			name: "Find dependencies",
			args: args{unmetDependenciesOutput},
			want: []bundle.NameVersion{
				{"kubelet", "1.13.0"},
				{"kubectl", "1.13.0"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDependencies(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseResultType(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want bundle.InstallResultType
	}{
		{
			name: "Correctly parses unknown result",
			args: args{essentialPackagesToBeRemovedOutput},
			want: bundle.ResultUnknownProblem,
		},
		{
			name: "Correctly parses unmet dependencies result",
			args: args{unmetDependenciesOutput},
			want: bundle.ResultUnmetDependencies,
		},
		{
			name: "Correctly parses  already installed result",
			args: args{newerVersionInstalledOutput},
			want: bundle.ResultNewerAlreadyInstalled,
		},
		{
			name: "Correctly parses unable to locate package result",
			args: args{unableToLocatePackage},
			want: bundle.ResultCannotFindPackage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseResultType(tt.args.msg); got != tt.want {
				t.Errorf("parseResultType() = %v, want %v", got, tt.want)
			}
		})
	}
}
