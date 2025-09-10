package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"strings"
	"sync"
)

// I18n 结构体
type I18n struct {
	bundle *i18n.Bundle
}

var (
	I18nUtil *I18n
	once     sync.Once
)

func InitI18n() {
	once.Do(func() {
		bundle := i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		// 加载本地化文件
		bundle.MustLoadMessageFile("core/locales/en.json")
		bundle.MustLoadMessageFile("core/locales/pt-BR.json")
		bundle.MustLoadMessageFile("core/locales/id.json")
		I18nUtil = &I18n{bundle: bundle}
		log.Printf("i18n Succeed")
	})
}

// Translate 翻译函数
func (i *I18n) Translate(c *gin.Context, key string, data map[string]interface{}) string {
	if i == nil || i.bundle == nil {
		return key
	}
	lang := strings.ToLower(c.GetHeader("Accept-Language"))
	var langMap = map[string]string{
		"en": "en",
		"pt": "pt-BR",
		"id": "id",
	}
	if _, ok := langMap[lang]; !ok {
		lang = "en"
	}

	localizer := i18n.NewLocalizer(i.bundle, lang)
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return message
}
