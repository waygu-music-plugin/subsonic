package subsonic

import (
	"encoding/json"

	"github.com/waygu-music-plugin/sdk/pluginapi"
	"github.com/waygu-music-plugin/sdk/pluginapi/storage"
	backend "github.com/waygu-music-plugin/subsonic/internal/client"
)

func credentials(id pluginapi.ID) storage.Credential {
	if id == "" {
		id = ID
	}
	return storage.For(id).LegacyCredential("credentials", "")
}

type subsonicCredentials struct {
	Password string `json:"password"`
}

func LoadCredentials(id pluginapi.ID, cfg *backend.Config) {
	if cfg == nil {
		return
	}
	b, err := credentials(id).Read()
	var sec subsonicCredentials
	if err != nil || json.Unmarshal(b, &sec) != nil {
		return
	}
	cfg.Password = sec.Password
}

func SaveCredentials(id pluginapi.ID, cfg backend.Config) error {
	if cfg == (backend.Config{}) {
		return nil
	}
	b, _ := json.Marshal(subsonicCredentials{Password: cfg.Password})
	return credentials(id).Write(b)
}

func ClearCredentials(id pluginapi.ID) { _ = credentials(id).Delete() }
