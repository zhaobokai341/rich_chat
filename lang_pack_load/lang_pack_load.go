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
	content, _ := os.ReadFile(lp.file)
	json.Unmarshal(content, &lp.data)
}

func (lp *LanguagePack) G(key string) string {
	translations, _ := lp.data[key]
	text, _ := translations[lp.language]
	return text
}
