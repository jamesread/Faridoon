package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.RequiredMigration != "6.audit-logs.sql" {
		t.Fatalf("migration=%s", cfg.RequiredMigration)
	}
	if cfg.Listen != ":8080" {
		t.Fatalf("listen=%s", cfg.Listen)
	}
}

func TestSetConfigDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("siteTitle: TestSite\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	SetConfigDir(dir)
	t.Cleanup(func() { SetConfigDir("") })
	cfg := LoadConfig()
	if cfg.SiteTitle != "TestSite" {
		t.Fatalf("title=%s", cfg.SiteTitle)
	}
}

func TestListenAddr(t *testing.T) {
	t.Setenv("PORT", "")
	cfg := &Config{Listen: ":9090"}
	if got := ListenAddr(cfg); got != ":9090" {
		t.Fatalf("fallback got %s", got)
	}
	t.Setenv("PORT", "3000")
	if got := ListenAddr(cfg); got != ":3000" {
		t.Fatalf("PORT number got %s", got)
	}
	t.Setenv("PORT", "0.0.0.0:4000")
	if got := ListenAddr(cfg); got != "0.0.0.0:4000" {
		t.Fatalf("PORT addr got %s", got)
	}
}
