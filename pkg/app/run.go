package app

import (
	"errors"
	"fmt"
	deb "github.com/dawidd6/deber/pkg/debian"
	doc "github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func run(cmd *cobra.Command, args []string) error {
	steps := map[string]func(*doc.Docker, *deb.Debian, *naming.Naming) error{
		"check":   runCheck,
		"build":   runBuild,
		"create":  runCreate,
		"start":   runStart,
		"tarball": runTarball,
		"scan":    runScan,
		"update":  runUpdate,
		"deps":    runDeps,
		"package": runPackage,
		"test":    runTest,
		"stop":    runStop,
		"remove":  runRemove,
		"archive": runArchive,
	}
	keys := []string{
		"check",
		"build",
		"create",
		"start",
		"tarball",
		"scan",
		"update",
		"deps",
		"package",
		"test",
		"stop",
		"remove",
		"archive",
	}

	log.Info("Parsing Debian changelog")
	debian, err := deb.ParseChangelog()
	if err != nil {
		return log.FailE(err)
	}
	log.Done()

	log.Info("Connecting with Docker")
	docker, err := doc.New()
	if err != nil {
		return log.FailE(err)
	}
	log.Done()

	name := naming.New(
		cmd.Use,
		debian.TargetDist,
		debian.SourceName,
		debian.PackageVersion,
	)

	if include != "" && exclude != "" {
		return errors.New("can't specify --include and --exclude together")
	}

	if include != "" {
		for key := range steps {
			if !strings.Contains(include, key) {
				delete(steps, key)
			}
		}
	}

	if exclude != "" {
		for key := range steps {
			if strings.Contains(exclude, key) {
				delete(steps, key)
			}
		}
	}

	for i := range keys {
		f, ok := steps[keys[i]]
		if ok {
			err := f(docker, debian, name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runCheck(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Checking archive")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		log.Skip()
		os.Exit(0)
	}

	return log.DoneE()
}

func runBuild(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := docker.IsImageBuilt(name.Image)
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt {
		isImageOld, err := docker.IsImageOld(name.Image)
		if err != nil {
			return log.FailE(err)
		}
		if !isImageOld {
			return log.SkipE()
		}
	}

	for _, repo := range []string{"debian", "ubuntu"} {
		tags, err := doc.GetTags(repo)
		if err != nil {
			return log.FailE(err)
		}

		for _, tag := range tags {
			if tag.Name == debian.TargetDist {
				from := fmt.Sprintf("%s:%s", repo, debian.TargetDist)

				log.Drop()

				err := docker.BuildImage(name.Image, from)
				if err != nil {
					return log.FailE(err)
				}

				return log.DoneE()
			}
		}
	}

	return log.FailE(errors.New("dist image not found"))
}

func runCreate(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := docker.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	err = docker.CreateContainer(name)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runStart(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStarted {
		return log.SkipE()
	}

	err = docker.StartContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runTarball(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Moving tarball")

	source := filepath.Join(name.SourceParentDir, debian.TarballFileName)
	target := filepath.Join(name.BuildDir, debian.TarballFileName)

	if debian.TarballFileName != "" {
		err := os.Rename(source, target)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}

func runScan(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Scanning archive")

	log.Drop()

	err := docker.ExecContainer(name.Container, "scan")
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runUpdate(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Updating cache")

	log.Drop()

	err := docker.ExecContainer(name.Container, "sudo", "apt-get", "update")
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runDeps(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Installing dependencies")

	log.Drop()

	err := docker.ExecContainer(name.Container, "sudo", "mk-build-deps", "-ri", "-t", "apty")
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runPackage(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Packaging software")

	file := fmt.Sprintf("%s/%s", name.ArchiveDir, "Packages")
	info, err := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	log.Drop()

	flags := os.Getenv("DEBER_DPKG_BUILDPACKAGE_FLAGS")
	if flags == "" {
		flags = "-tc"
	}
	command := append([]string{"dpkg-buildpackage"}, strings.Split(flags, " ")...)
	err = docker.ExecContainer(name.Container, command...)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runTest(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Testing package")

	log.Drop()

	err := docker.ExecContainer(name.Container, "debc")
	if err != nil {
		return log.FailE(err)
	}

	err = docker.ExecContainer(name.Container, "sudo", "debi", "--with-depends", "--tool", "apty")
	if err != nil {
		return log.FailE(err)
	}

	flags := os.Getenv("DEBER_LINTIAN_FLAGS")
	if flags == "" {
		flags = "-i"
	}
	command := append([]string{"lintian"}, strings.Split(flags, " ")...)
	err = docker.ExecContainer(name.Container, command...)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runStop(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStopped {
		return log.SkipE()
	}

	err = docker.StopContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runRemove(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	err = docker.RemoveContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runArchive(docker *doc.Docker, debian *deb.Debian, name *naming.Naming) error {
	log.Info("Archiving build")

	info, err := os.Stat(name.ArchivePackageDir)
	if info != nil {
		return log.SkipE()
	}

	err = os.Rename(name.BuildDir, name.ArchivePackageDir)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
