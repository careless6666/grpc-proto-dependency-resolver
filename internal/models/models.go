package models

const (
	DependencyTypeGit  = iota
	DependencyTypeURL  = iota
	DependencyTypePath = iota
)

const RootPath string = ".proto_deps"

type Dependency struct {
	//locale | remote
	Type int
	Path string
	// for url | path
	DestinationPath string
	// path inside git repo
	GitPath string
	Version *VersionInfo
}

type VersionInfo struct {
	Tag            string
	CommitRevision string
}
