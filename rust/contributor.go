package rust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/kardolus/rust-cnb/utils"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "rustup"
	Cache      = "cache"
)

var CargoToml string

type Contributor struct {
	CacheMetadata      Metadata
	manager            PackageManager
	app                application.Application
	rustupLayer        layers.DependencyLayer
	pkgLayer           layers.Layer
	launchLayer        layers.Layers
	buildContribution  bool
	launchContribution bool
}

type PackageManager interface {
	Install(location string, layer layers.Layer) error
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

func NewContributor(context build.Build, manager PackageManager) (Contributor, bool, error) {
	plan, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	if plan.Version == "" {
		context.Logger.SubsequentLine("Dependency version not specified, but is required")
		return Contributor{}, false, nil
	}

	CargoToml = filepath.Join(context.Application.Root, utils.CARGO_TOML)
	if exists, err := helper.FileExists(CargoToml); err != nil {
		return Contributor{}, false, err
	} else if !exists {
		return Contributor{}, false, fmt.Errorf("unable to find " + utils.CARGO_TOML)
	}

	cargoLock := filepath.Join(context.Application.Root, utils.CARGO_LOCK)
	hex, err := hash(cargoLock)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		manager:       manager,
		app:           context.Application,
		rustupLayer:   context.Layers.DependencyLayer(dep),
		pkgLayer:      context.Layers.Layer(Dependency),
		launchLayer:   context.Layers,
		CacheMetadata: Metadata{Dependency, hex},
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = false
	}
	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	if err := c.rustupLayer.Contribute(c.contributeRustup, c.flags()...); err != nil {
		return err
	}

	// nil means, run every time
	if err := c.pkgLayer.Contribute(nil, c.contributePackages, c.flags()...); err != nil {
		return err
	}

	return c.contributeStartCommand()
}

func (c Contributor) contributeRustup(artifact string, layer layers.DependencyLayer) error {
	layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
	return helper.ExtractTarGz(artifact, layer.Root, 1)
}

func (c Contributor) contributePackages(layer layers.Layer) error {
	layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
	return c.manager.Install(c.app.Root, layer)
}

func (c Contributor) contributeStartCommand() error {
	meta, err := utils.ParseAppMetadata(c.app.Root)
	if err != nil {
		return err
	}

	return c.launchLayer.WriteMetadata(layers.Metadata{Processes: []layers.Process{{"web", filepath.Join(c.app.Root, "target", "release", meta.Package.Name)}}})
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}
	return flags
}

func hash(cargoLock string) (string, error) {
	var hash [32]byte

	if _, err := os.Stat(cargoLock); err == nil {
		buf, err := ioutil.ReadFile(cargoLock)
		if err != nil {
			return "", err
		}
		hash = sha256.Sum256(buf)
	} else { // set "random" hash
		timestamp := time.Now().Unix()
		hash = sha256.Sum256([]byte(string(timestamp)))
	}

	return hex.EncodeToString(hash[:]), nil
}
