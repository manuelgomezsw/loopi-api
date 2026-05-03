package config

import (
	"strings"
	"testing"
)

func TestResolveDBName_Development(t *testing.T) {
	name, err := resolveDBName("development")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "loopi_dev" {
		t.Errorf("got %q, want %q", name, "loopi_dev")
	}
}

func TestResolveDBName_Test(t *testing.T) {
	name, err := resolveDBName("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "loopi_dev" {
		t.Errorf("got %q, want %q", name, "loopi_dev")
	}
}

func TestResolveDBName_Production(t *testing.T) {
	name, err := resolveDBName("production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "loopi_prod" {
		t.Errorf("got %q, want %q", name, "loopi_prod")
	}
}

func TestResolveDBName_Unknown(t *testing.T) {
	_, err := resolveDBName("staging")
	if err == nil {
		t.Fatal("expected error for unknown APP_ENV, got nil")
	}
	if !strings.Contains(err.Error(), "staging") {
		t.Errorf("error message should contain the invalid value, got: %v", err)
	}
	for _, valid := range []string{"development", "test", "production"} {
		if !strings.Contains(err.Error(), valid) {
			t.Errorf("error message should list valid option %q, got: %v", valid, err)
		}
	}
}

func TestLoad_MissingAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when APP_ENV is empty, got nil")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Errorf("error should name the missing variable, got: %v", err)
	}
}

func TestLoad_InvalidAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for APP_ENV=staging, got nil")
	}
	if !strings.Contains(err.Error(), "staging") {
		t.Errorf("error should contain the invalid value, got: %v", err)
	}
	for _, valid := range []string{"development", "test", "production"} {
		if !strings.Contains(err.Error(), valid) {
			t.Errorf("error should list valid option %q, got: %v", valid, err)
		}
	}
}

func TestLoad_DevelopmentEnv(t *testing.T) {
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Name != "loopi_dev" {
		t.Errorf("got Database.Name=%q, want %q", cfg.Database.Name, "loopi_dev")
	}
}

func TestLoad_ProductionEnv(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Name != "loopi_prod" {
		t.Errorf("got Database.Name=%q, want %q", cfg.Database.Name, "loopi_prod")
	}
}
