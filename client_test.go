package ecos

import (
	"testing"
)

func TestNewClientRequiresKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("want error for empty apiKey")
	}
}

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient("KEY")
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Setenv("ECOS_API_KEY", "ENVKEY")
	if _, err := NewClientFromEnv(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ECOS_API_KEY", "")
	if _, err := NewClientFromEnv(); err == nil {
		t.Fatal("want error when ECOS_API_KEY unset")
	}
}

func TestPageOrDefault(t *testing.T) {
	s, e := (Page{}).orDefault()
	if s != 1 || e != 100 {
		t.Errorf("zero Page = %d..%d, want 1..100", s, e)
	}
	s, e = (Page{Start: 5, End: 20}).orDefault()
	if s != 5 || e != 20 {
		t.Errorf("Page{5,20} = %d..%d", s, e)
	}
}

func TestParseFloat(t *testing.T) {
	got, err := parseFloat("2.5")
	if err != nil || got != 2.5 {
		t.Errorf("parseFloat(2.5) = %v, %v", got, err)
	}
	got, err = parseFloat("1,397,923")
	if err != nil || got != 1397923 {
		t.Errorf("parseFloat(1,397,923) = %v, %v", got, err)
	}
	if _, err = parseFloat(""); err == nil {
		t.Error("want error for empty string")
	}
}
