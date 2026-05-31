package main

import "rich_chat/lang_pack_load"

// LanguagePackWrapper wraps the language pack for easier access
type LanguagePackWrapper struct {
	lp *lang_pack_load.LanguagePack
}

// NewLanguagePackWrapper creates a new language pack wrapper
func NewLanguagePackWrapper(filePath, language string) *LanguagePackWrapper {
	lp := lang_pack_load.NewLanguagePack(filePath, language)
	lp.Load()
	return &LanguagePackWrapper{
		lp: lp,
	}
}

// Get retrieves a localized string by key
func (w *LanguagePackWrapper) Get(key string) string {
	return w.lp.G(key)
}
