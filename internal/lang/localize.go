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

func GetManager() *Translator {
	once.Do(func() {
		instance = &Translator{}
	})
	return instance
}

func LoadLanguage(path string, langs ...string) {
	// Default language is the first specified (english if none)
	defaultLang := language.English
	if len(langs) > 0 {
		defaultLang = language.Make(langs[0])
	}
	logger.Infof("Available languages: %v (default: %s)", langs, defaultLang)

	// Create a bundle to use for the program's lifetime
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load translations into bundle
	loadDefaults := false
	if path != "" {
		// Load custom language file
		_, err := bundle.LoadMessageFile(path)
		if err != nil {
			logger.Errorf("Error loading custom language file (%s): %s", path, err.Error())
			loadDefaults = true
		} else {
			logger.Infof("Loading custom messages: %s", path)
		}
	} else {
		loadDefaults = true
	}

	if loadDefaults {
		// Load embedded language files
		// langDir := filepath.Dir(path)
		for _, lang := range langs { //[1:] {
			langFile := fmt.Sprintf("lang.%s.toml", lang)
			_, err := bundle.LoadMessageFileFS(langFS, langFile) //filepath.Join(langDir, langFile))
			if err != nil {
				logger.Errorf("Error loading language file (%s): %s", lang, err.Error())
			}
		}
	}

	// Create a Localizer to use for a set of language preferences
	lm := GetManager()
	lm.Bundle = bundle
	lm.Localizer = i18n.NewLocalizer(bundle, langs...)
	lm.Langs = langs
}

func (lm Translator) Localize(messageID string, langs ...string) string {
	res, err := lm.Localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
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
		res = messageID
	}
	return res
}
