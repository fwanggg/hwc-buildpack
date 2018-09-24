package supply

import (
	"fmt"
	"github.com/cloudfoundry/libbuildpack"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	//this is where app is located
	BuildDir() string
	//DepsDir and DepsIdx == DepDir
	DepDir() string
	//depends on how many build packs were installed
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
	s.Log.BeginStep("Installing/Supplying Riverbed")
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
	//`reg add` setup LoaderOptimization in HKCU so we don't have the grant set problem
	//Loading this assembly would produce a different grant set from other instances. (Exception from HRESULT: 0x80131401)
	if err:=ioutil.WriteFile(filepath.Join(s.Stager.DepDir(),"profile.d","riverbed.bat"),[]byte(`set COR_PROFILER_PATH_64=%DEPS_DIR%\`+ s.Stager.DepsIdx() +`\Panorama\hedzup\mn\bin\x64\AwDotNetProf.dll
	set COR_ENABLE_PROFILING=1
	set COR_PROFILER={CEBBDDAB-C0A5-4006-9C04-2C3B75DF2F80}
    set PATH=%PATH%;%DEPS_DIR%\`+ s.Stager.DepsIdx() +`\Panorama\hedzup\mn\bin\
    set RVBD_IN_PCF=1
	reg add HKCU\SOFTWARE\Microsoft\.NETFramework /v LoaderOptimization /t REG_DWORD /d 1 /f
	`), 0777) ; err!=nil{
		return fmt.Errorf("ioutil.WriteFIle: %s", err)
	}


	//mv AwDotNetCore2.dll and AwDotNetCore4.dll to <asp.net>\bin\ folder, this is where assembly is to be searched
	//It doesn't looks like asp.net would look at executable path for assembly and since we don't have admin right
	//registering our helper assemblies in GAC also won't work
	dllfiles := filepath.Join(s.Stager.DepDir(), "Panorama", "hedzup", "mn", "bin" , "AwDotNetCore*.dll")
	corefiles, err := filepath.Glob(dllfiles)

	if err != nil{
		return fmt.Errorf("Glob " + dllfiles + " %s", err)
	}

	for _, file := range corefiles{
		target := filepath.Join(s.Stager.BuildDir(), "bin", filepath.Base(file))
		s.Log.BeginStep("moving : " + file + " to " + target)

		if err:=libbuildpack.CopyFile(file, target); err!=nil{
			return fmt.Errorf("mv " + target + ": %s", err)
		}
	}



	return nil

}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying hwc")

	dep := libbuildpack.Dependency{Name: "hwc", Version: "12.0.0"}
	depDir := filepath.Join(s.Stager.DepDir(), "hwc")
	if err := s.Installer.InstallDependency(dep, depDir); err != nil {
		return err
	}
	if err := s.InstallRiverbed(); err != nil{
		return err
	}
	return nil
}
