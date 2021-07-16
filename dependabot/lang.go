package dependabot

const (
	Bundler Lang = iota + 1
	Cargo
	Composer
	Docker
	Elm
	GHAction
	GitSubmodule
	GoModules
	Gradle
	Hex
	Maven
	Npm
	NuGet
	Pip
	Terraform
)

type Lang int

// TODO: Setup all langs

var Langs = []Lang{
	Docker,
	GHAction,
	GoModules,
	Npm,
}

var langsDepsFiles = map[Lang]DepFile{
	Docker:    {FilesPattern: []string{"Dockerfile"}},
	GHAction:  {FilesPattern: []string{".yml", ".yaml"}, Path: ".github/workflows"},
	GoModules: {FilesPattern: []string{"go.mod"}},
	Npm:       {FilesPattern: []string{"package.json"}},
}

var langEcosystem = map[Lang]string{
	Docker:    "docker",
	GHAction:  "github-actions",
	GoModules: "gomod",
	Npm:       "npm",
}

type DepFile struct {
	FilesPattern []string
	Path         string
}

func (l Lang) DepsFile() DepFile {
	return langsDepsFiles[l]
}

func (l Lang) Ecosystem() string {
	return langEcosystem[l]
}

type Repository struct {
	Name string
	Deps []DependencyFile
}

type DependencyFile struct {
	Lang Lang
	Path string
}
