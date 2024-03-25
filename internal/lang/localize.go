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

func LoadLanguages(langFile string, langs ...string) Translator {
	// Default language is the first specified (english if none)
	defaultLang := language.English
	if len(langs) == 0 {
		langs = append(langs, defaultLang.String())
	} else if len(langs) > 0 {
		defaultLang = language.Make(langs[0])
	}

	logger.Infof("Available languages: %v (default: %s)", langs, defaultLang)

	// Create a bundle to use for the program's lifetime
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load translations into bundle
	loadEmbedded := false
	if langFile != "" {
		// Load custom language file
		_, err := bundle.LoadMessageFile(langFile)
		if err != nil {
			logger.Errorf("Error loading custom language file (%s): %s", langFile, err.Error())
			loadEmbedded = true
		} else {
			logger.Infof("Loading custom messages: %s", langFile)
		}
	} else {
		loadEmbedded = true
	}

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
		logger.Error(err, langs)
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
		logger.Error(err, langs)
		res = messageID
	}
	return res
}
