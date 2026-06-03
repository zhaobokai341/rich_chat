package lang_pack_load

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type LanguagePack struct {
	file     string
	language string
	data     map[string]map[string]string
}

func NewLanguagePack(file string, language string) *LanguagePack {
	fullPath := filepath.Join("../lang_pack/", file)

	return &LanguagePack{
		file:     fullPath,
		language: language,
		data:     make(map[string]map[string]string),
	}
}

func (lp *LanguagePack) Load() {
	content, err := os.ReadFile(lp.file)
	if err != nil {
		// Handle error appropriately, maybe log it
		return
	}
	err = json.Unmarshal(content, &lp.data)
	if err != nil {
		// Handle error appropriately, maybe log it
		return
	}
}

func (lp *LanguagePack) G(key string) string {
	translations, ok := lp.data[key]
	if !ok {
		return key // Return the key itself if translation not found
	}
	text, ok := translations[lp.language]
	if !ok {
		// Try fallback to English if language not found
		text, ok = translations["en"]
		if !ok {
			return key // Return the key itself if translation not found
		}
	}
	return text
}
