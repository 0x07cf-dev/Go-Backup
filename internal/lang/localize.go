package lang

import (
	"embed"
	"fmt"
	"sync"

	"github.com/0x07cf-dev/go-backup/internal/logger"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Translator struct {
	Bundle    *i18n.Bundle
	Localizer *i18n.Localizer
	Langs     []string
}

var (
	instance *Translator
	once     sync.Once
)

//go:embed lang.*.toml
var langFS embed.FS

func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{}
	})
	return instance
}

func LoadLanguages(langFile string, lang string) Translator {
	// Default language is the first specified (english if none)
	var defaultLang language.Tag
	if lang == "" {
		defaultLang = language.English
	} else {
		defaultLang = language.Make(lang)
	}
	langs := []string{defaultLang.String()}

	if len(langs) == 0 {
		langs = append(langs, "en")
	}

	// Create a bundle to use for the program's lifetime
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load translations into bundle
	// Look for custom language file first
	loadEmbedded := false
	if langFile != "" {
		_, err := bundle.LoadMessageFile(langFile)
		if err != nil {
			logger.Errorf("Error loading custom language file (%s): %s", langFile, err.Error())
			loadEmbedded = true
		} else {
			logger.Debugf("Loading custom messages: %s", langFile)
		}
	} else {
		loadEmbedded = true
	}

	// Fallback to embedded language files
	if loadEmbedded {
		// Load embedded language files
		for _, lang := range langs { //[1:] {
			langFile := fmt.Sprintf("lang.%s.toml", lang)
			_, err := bundle.LoadMessageFileFS(langFS, langFile) //filepath.Join(langDir, langFile))
			if err != nil {
				logger.Errorf("Error loading language file (%s): %s", lang, err.Error())
			}
		}
	}

	logger.Debugf("Available languages: %v (default: %s)", langs, defaultLang)

	// Create a Localizer to use for a set of language preferences
	tr := GetTranslator()
	tr.Bundle = bundle
	tr.Localizer = i18n.NewLocalizer(bundle, langs...)
	tr.Langs = langs
	return *tr
}

func (lm Translator) Localize(messageID string, langs ...string) string {
	res, err := lm.Localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		logger.Errorf("%s %s", err, langs)
		res = messageID
	}
	return res
}

func (lm Translator) LocalizeTemplate(messageID string, template map[string]string, langs ...string) string {
	res, err := lm.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: template,
	})
	if err != nil {
		logger.Errorf("%s %s", err, langs)
		res = messageID
	}
	return res
}
