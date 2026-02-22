package i18n

import (
	"reflect"
	"testing"
)

func TestDefaultIsEnglish(t *testing.T) {
	// Reset to default
	Set(LangEN)
	m := Get()
	if m.MenuTitle != messagesEN.MenuTitle {
		t.Errorf("default should be EN, got MenuTitle=%q", m.MenuTitle)
	}
}

func TestSetEachLanguage(t *testing.T) {
	tests := []struct {
		lang     Lang
		expected *Messages
	}{
		{LangEN, &messagesEN},
		{LangFR, &messagesFR},
		{LangES, &messagesES},
		{LangDE, &messagesDE},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			Set(tt.lang)
			got := Get()
			if got != tt.expected {
				t.Errorf("Set(%q) did not activate the correct messages", tt.lang)
			}
		})
	}
	// Restore default
	Set(LangEN)
}

func TestUnknownLangFallsBackToEN(t *testing.T) {
	Set(Lang("xx"))
	m := Get()
	if m != &messagesEN {
		t.Error("unknown language should fall back to EN")
	}
}

func TestAllFieldsNonEmpty(t *testing.T) {
	allMsgs := []*Messages{&messagesEN, &messagesFR, &messagesES, &messagesDE}
	langs := []string{"en", "fr", "es", "de"}

	for i, msgs := range allMsgs {
		v := reflect.ValueOf(*msgs)
		typ := v.Type()
		for j := 0; j < v.NumField(); j++ {
			field := v.Field(j)
			if field.Kind() == reflect.String && field.String() == "" {
				t.Errorf("lang %s: field %s is empty", langs[i], typ.Field(j).Name)
			}
		}
	}
}

func TestSupportedLangs(t *testing.T) {
	langs := SupportedLangs()
	if len(langs) != 4 {
		t.Fatalf("expected 4 supported langs, got %d", len(langs))
	}
	expected := map[Lang]bool{LangEN: true, LangFR: true, LangES: true, LangDE: true}
	for _, l := range langs {
		if !expected[l] {
			t.Errorf("unexpected lang: %s", l)
		}
	}
}
