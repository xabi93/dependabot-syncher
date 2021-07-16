package dependabot

import (
	_ "embed"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const ConfPathFile = ".github/dependabot.yml"

const (
	dependabotVersion = 2
	//TODO: Allow custom setup
	updateInterval = "weekly"
)

type conf struct {
	Version int      `yaml:"version"`
	Updates []update `yaml:"updates"`
}

type update struct {
	Ecosystem string   `yaml:"package-ecosystem"`
	Directory string   `yaml:"directory"`
	Schedule  schedule `yaml:"schedule"`
}

type schedule struct {
	Interval string
}

func Generate(repo Repository) ([]byte, error) {
	updates := []update{}

	has := map[string]struct{}{}

	for _, r := range repo.Deps {
		path := filepath.Dir(r.Path)
		path = "/" + strings.TrimPrefix(path, ".")
		if _, ok := has[r.Lang.Ecosystem()+"_"+path]; ok {
			continue
		} else {
			has[r.Lang.Ecosystem()+"_"+path] = struct{}{}
		}

		updates = append(updates, update{
			Ecosystem: r.Lang.Ecosystem(),
			Directory: path,
			Schedule: schedule{
				Interval: updateInterval,
			},
		})
	}

	return yaml.Marshal(conf{
		Version: dependabotVersion,
		Updates: updates,
	})
}
