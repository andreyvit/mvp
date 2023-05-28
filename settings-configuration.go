package mvp

import (
	"embed"
	"sort"

	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"golang.org/x/exp/maps"
)

type Configuration struct {
	Envs map[string][]string

	Preinstall func(settings *Settings)
	Install    func(settings *Settings)

	BuildCommit string
	BuildVer    string

	EmbeddedConfig   []byte
	EmbeddedSecrets  string
	EmbeddedStaticFS embed.FS
	EmbeddedViewsFS  embed.FS

	ConfigFileName  string
	SecretsFileName string
	StaticSubdir    string
	ViewsSubdir     string
	LocalDevAppRoot string

	Modules []*Module

	LoadSecrets func(*Settings, Secrets)

	Types map[mvpm.Type][]string

	AuthTokenCookieName string
	AuthTokenKeys       mvpm.NamedKeySet
}

func (ge *Configuration) ValidEnvs() []string {
	result := maps.Keys(ge.Envs)
	sort.Strings(result)
	return result
}
