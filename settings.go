package mvp

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreyvit/jsonfix"
	"github.com/andreyvit/mvp/jsonext"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/andreyvit/mvp/postmark"
	"github.com/andreyvit/plainsecrets"
	"golang.org/x/exp/maps"
)

type Settings struct {
	Env           string
	Configuration *Configuration

	// Configuration options
	LocalOverridesFile string
	AutoEncryptSecrets bool
	KeyringFile        string

	// HTTP server options
	BindAddr             string
	BindPort             int
	ForwardingProxyCIDRs string

	// job options
	WorkerCount           int
	EphemeralWorkerCount  int
	EphemeralQueueMaxSize int

	// app options
	AppName                  string // user-visible app name
	AppID                    string // unchangeable internal name for various identification purposes
	BaseURL                  string
	RateLimits               map[RateLimitPreset]map[RateLimitGranularity]RateLimitSettings
	MaxRateLimitRequestDelay jsonext.Duration
	AppBehaviors

	DataDir string

	VerboseDB bool

	Postmark                     postmark.Credentials
	PostmarkDefaultMessageStream string
	EmailDefaultFrom             string
	EmailDefaultLayout           string

	JWTIssuers []string // Issuer and Audience for this app's tokens
}

type Secrets map[string]string

func (secrets Secrets) Optional(name string, val interface{ Set(string) error }) bool {
	str := secrets[name]
	if str == "" {
		return false
	}
	err := val.Set(str)
	if err != nil {
		log.Fatalf("** ERROR: invalid value of secret %s: %v", name, err)
	}
	return true
}

func (secrets Secrets) Required(name string, val interface{ Set(string) error }) {
	ok := secrets.Optional(name, val)
	if !ok {
		log.Fatalf("** ERROR: missing secret %s", name)
	}
}

func (secrets Secrets) OptionalNamedKeySet(name string, keys *mvpm.NamedKeySet, minKeyLen, maxKeyLen int) bool {
	active := secrets[name]
	if active == "" {
		return false
	}

	m := make(map[string][]byte)
	prefix := name + "_"
	for k, v := range secrets {
		if keyName, ok := strings.CutPrefix(k, prefix); ok {
			key, err := hex.DecodeString(v)
			if err != nil {
				log.Fatalf("** ERROR: %s: invalid hex value: %v", k, err)
			}
			if len(key) < minKeyLen {
				log.Fatalf("** ERROR: %s: key is only %d bytes, wanted at least %d bytes", k, len(key), minKeyLen)
			}
			if maxKeyLen > 0 && len(key) > maxKeyLen {
				log.Fatalf("** ERROR: %s: key is %d bytes, wanted at most %d bytes", k, len(key), maxKeyLen)
			}
			m[keyName] = key
		}
	}
	if len(m) == 0 {
		return false
	}
	if m[active] == nil {
		log.Fatalf("** ERROR: %s: active key %q is not in the set", name, active)
	}
	*keys = mvpm.NamedKeySet{Keys: m, ActiveKeyName: active}
	return true
}

func (secrets Secrets) RequiredNamedKeySet(name string, keys *mvpm.NamedKeySet, minKeyLen, maxKeyLen int) {
	ok := secrets.OptionalNamedKeySet(name, keys, minKeyLen, maxKeyLen)
	if !ok {
		log.Fatalf("** ERROR: missing secret key set %s", name)
	}
}

func LoadConfig(ge *Configuration, env string, installHook func(*Settings)) *Settings {
	settings := BaseSettings.New()
	settings.Env = env
	full := BaseSettings.AnyFull(settings)

	configBySection := make(map[string]json.RawMessage)

	decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(ge.EmbeddedConfig)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&configBySection)
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.json: %w", err))
	}

	for e := range ge.Envs {
		if e == env {
			parseConfigSections(ge, ge.Envs[e], configBySection, full)
		} else {
			// parse all other sections to ensure we're not surprised after deployment
			dummy := BaseSettings.AnyFull(BaseSettings.New())
			parseConfigSections(ge, ge.Envs[e], configBySection, dummy)
		}
	}

	if installHook != nil {
		installHook(settings)
	}

	if overridesFile := settings.LocalOverridesFile; overridesFile != "" {
		raw, err := os.ReadFile(overridesFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("** %v", fmt.Errorf("LocalOverridesFile %s is missing, create with {} as its contents", overridesFile))
			}
			log.Fatalf("** %v", fmt.Errorf("%s: %w", overridesFile, err))
		}

		decoder := json.NewDecoder(bytes.NewReader(jsonfix.Bytes(raw)))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("%s: %w", overridesFile, err))
		}
	}

	if settings.KeyringFile == "" {
		log.Fatalf("** %v", fmt.Errorf("config.json: empty KeyringFile"))
	}
	keyring, err := plainsecrets.ParseKeyringFile(settings.KeyringFile)
	if err != nil {
		log.Fatalf("** %v", err)
	}

	vals, err := plainsecrets.ParseString(fmt.Sprintf("@all = %s\n%s", strings.Join(maps.Keys(ge.Envs), " "), ge.EmbeddedSecrets))
	if err != nil {
		log.Fatalf("** %v", fmt.Errorf("config.secrets.txt: %w", err))
	}

	log.Printf("Settings = %s", must(json.Marshal(settings)))
	settings.Configuration = ge

	if settings.AutoEncryptSecrets && ge.LocalDevAppRoot != "" {
		secretsFile := filepath.Join(ge.LocalDevAppRoot, ge.SecretsFileName)
		_, err := os.Stat(secretsFile)
		if err != nil {
			log.Fatalf("Cannot autoencrypt secrets because file %v does not exist", secretsFile)
		}

		n, failed, err := vals.EncryptAllInFile(secretsFile, keyring)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("autoencrypt failed: %w", err))
		}
		if len(failed) > 0 {
			var msgs []string
			var msgSet = make(map[string]bool)
			for _, v := range failed {
				s := v.Err.Error()
				if !msgSet[s] {
					msgSet[s] = true
					msgs = append(msgs, s)
				}
			}
			log.Fatalf("** %v", fmt.Errorf("autoencrypt failed: %s", strings.Join(msgs, ", ")))
		}
		if n > 0 {
			log.Printf("Auto-encrypted %d secret(s).", n)
		}
	}

	secrets, err := vals.EnvValues(env, keyring)
	if err != nil {
		log.Fatalf("** %v", err)
	}
	ge.LoadSecrets(settings, secrets)

	return settings
}

func parseConfigSections(ge *Configuration, sections []string, configBySection map[string]json.RawMessage, settings any) {
	for _, section := range sections {
		if configBySection[section] == nil {
			log.Fatalf("** %v", fmt.Errorf("%s: missing section %s", ge.ConfigFileName, section))
		}
		decoder := json.NewDecoder(bytes.NewReader(configBySection[section]))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(settings)
		if err != nil {
			log.Fatalf("** %v", fmt.Errorf("%s: %s: %w", ge.ConfigFileName, section, err))
		}
	}
}
