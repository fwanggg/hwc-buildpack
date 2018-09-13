package supply

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	//this is where app is located
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
	AddBinDependencyLink(string, string) error
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
	RootDir() string
}

type Installer interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/installer.go
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/command.go
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(dir string, program string, args ...string) (string, error)
}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
}

func (s *Supplier) InstallRiverbed() error{
	//todo:
	//return nil if no appinternals in VCAP_SERVICES, this would also fix unittest
	//check VCAP_SERVICES to make sure we are bound already
	//check which fields
	s.Log.BeginStep("Installing/SUpplying Riverbed")
	s.Log.BeginStep("depdir" + s.Stager.DepDir())
	//func ExtractTarGz(tarfile, destDir string) error
	if err:=libbuildpack.ExtractTarGz(filepath.Join(s.Manifest.RootDir(),"panorama.tgz"),s.Stager.DepDir()); err!=nil{
		return fmt.Errorf("extarct tgz: %s", err)
	}
	if err:=os.MkdirAll(filepath.Join(s.Stager.DepDir(),"profile.d"),0777); err!=nil{
		return fmt.Errorf("os.MkdirAll: %s", err)
	}

	//s.Stager.DepDir() is depdir in staging process
	//%DEPS_DIR% should be used at runtime
	if err:=ioutil.WriteFile(filepath.Join(s.Stager.DepDir(),"profile.d","riverbed.bat"),[]byte(`%DEPS_DIR%\`+ s.Stager.DepsIdx() +`\Panorama\hedzup\mn\bin\DotNetRegister64.exe
	set COR_ENABLE_PROFILING=1
	set COR_PROFILER={CEBBDDAB-C0A5-4006-9C04-2C3B75DF2F80}
	`), 0777) ; err!=nil{
		return fmt.Errorf("ioutil.WRiteFIle: %s", err)
	}

	//fmt.Println(exec.Command("find", s.Stager.DepDir()).Output())
	return nil

}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying hwc")

	dep := libbuildpack.Dependency{Name: "hwc", Version: "12.0.0"}
	depDir := filepath.Join(s.Stager.DepDir(), "hwc")
	if err := s.Installer.InstallDependency(dep, depDir); err != nil {
		return err
	}
	s.InstallRiverbed()
	return nil
}
