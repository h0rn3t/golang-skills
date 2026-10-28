package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const mod = "example.com/shop"

func modulesConfig() *config {
	cfg := &config{
		Layout:   "modules",
		Platform: []string{"internal/platform/clock"},
		Modules:  map[string]string{"order": "layered", "billing": "layered", "search": "flat"},
	}
	cfg.defaults()
	return cfg
}

func layersConfig() *config {
	cfg := &config{Layout: "layers"}
	cfg.defaults()
	return cfg
}

func pkg(path string, imports ...string) pkgInfo {
	return pkgInfo{ImportPath: mod + "/" + path, Imports: imports}
}

func in(path string) string { return mod + "/" + path }

func rules(vs []violation) []string {
	var out []string
	for _, v := range vs {
		out = append(out, v.Rule+" "+v.From+" -> "+v.To)
	}
	slices.Sort(out)
	return out
}

func TestCoherentLayeredServicePasses(t *testing.T) {
	pkgs := []pkgInfo{
		pkg("internal/handlers", in("internal/services"), in("internal/models"), "net/http"),
		pkg("internal/services", in("internal/models")),
		pkg("internal/repositories", in("internal/models"), "database/sql"),
		pkg("internal/models"),
		pkg("internal/app", in("internal/handlers"), in("internal/services"), in("internal/repositories"), "database/sql"),
		pkg("cmd/server", in("internal/app")),
	}
	if got := check(layersConfig(), mod, pkgs, false); len(got) != 0 {
		t.Fatalf("coherent B reported %v", rules(got))
	}
}

func TestHandlerToRepositoryShortcutIsALayerViolation(t *testing.T) {
	pkgs := []pkgInfo{pkg("internal/handlers", in("internal/repositories")), pkg("internal/repositories"), pkg("internal/handlers/v2", in("internal/repositories"))}
	got := rules(check(layersConfig(), mod, pkgs, false))
	want := []string{
		"layer internal/handlers -> internal/repositories",
		"layer internal/handlers/v2 -> internal/repositories",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v (the nested handlers/v2 package is still a handler)", got, want)
	}
}

func TestForeignImplementationVersusForeignRootContract(t *testing.T) {
	cfg := modulesConfig()
	pkgs := []pkgInfo{
		pkg("internal/billing/services", in("internal/order/repositories"), in("internal/order")),
		pkg("internal/billing/handlers", in("internal/order")),
		pkg("internal/search", in("internal/order")),
		pkg("internal/order"),
		pkg("internal/order/repositories"),
	}
	got := rules(check(cfg, mod, pkgs, false))
	want := []string{
		"ownership internal/billing/handlers -> internal/order",
		"ownership internal/billing/services -> internal/order/repositories",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v: services and a flat module may consume a foreign root contract; nobody imports a foreign implementation", got, want)
	}
}

func TestCompositionAndPlatformDirections(t *testing.T) {
	cfg := modulesConfig()
	pkgs := []pkgInfo{
		pkg("internal/order/services", in("internal/app"), in("internal/platform/clock"), in("internal/platform/cache")),
		pkg("internal/platform/clock", in("internal/order")),
		pkg("internal/platform/cache", in("internal/app")),
		pkg("internal/app", in("internal/order/services"), in("internal/platform/cache")),
	}
	got := rules(check(cfg, mod, pkgs, false))
	want := []string{
		"composition internal/order/services -> internal/app",
		"composition internal/platform/cache -> internal/app",
		"platform internal/order/services -> internal/platform/cache",
		"platform internal/platform/clock -> internal/order",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestModulePathContainingInternalIsClassifiedByExactPrefix(t *testing.T) {
	const tricky = "example.com/internal/shop"
	cfg := modulesConfig()
	p := pkgInfo{ImportPath: tricky + "/internal/order/services", Imports: []string{tricky + "/internal/order/models"}}
	if got := check(cfg, tricky, []pkgInfo{p}, false); len(got) != 0 {
		t.Fatalf("services -> own models reported %v", rules(got))
	}
	if c := classify(cfg, tricky, tricky+"/internal/order/services"); c.kind != kindLayer || c.module != "order" || c.layer != layerServices {
		t.Fatalf("classify = %+v, want order/services", c)
	}
}

func TestRootContractAndDrivers(t *testing.T) {
	cfg := modulesConfig()
	pkgs := []pkgInfo{
		pkg("internal/order", in("internal/order/models"), "net/http"),
		pkg("internal/order/services", "net/http"),
		pkg("internal/order/models", "gorm.io/gorm"),
		pkg("internal/order/handlers", "net/http", "database/sql"),
		pkg("internal/order/repositories", "database/sql", "github.com/jackc/pgx/v5"),
		pkg("internal/search", "database/sql", "net/http"),
	}
	got := rules(check(cfg, mod, pkgs, false))
	want := []string{
		"contract internal/order -> internal/order/models",
		"driver internal/order -> net/http",
		"driver internal/order/handlers -> database/sql",
		"driver internal/order/models -> gorm.io/gorm",
		"driver internal/order/services -> net/http",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v: repositories own drivers and a flat module's root may contain I/O", got, want)
	}
}

func TestUnclassifiedPackagesAndLocalTargets(t *testing.T) {
	cfg := modulesConfig()
	pkgs := []pkgInfo{
		pkg("internal/util"),
		pkg("internal/order/adapters"),
		pkg("internal/order/services", in("pkg/strutil")),
		pkg("pkg/strutil"),
	}
	got := rules(check(cfg, mod, pkgs, false))
	want := []string{
		"unclassified internal/order/adapters -> ",
		"unclassified internal/order/services -> pkg/strutil",
		"unclassified internal/util -> ",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestKnownListSuppressesExactMatchesAndFailsOnStaleDuplicateOrUnexplained(t *testing.T) {
	found := []violation{{Rule: "layer", From: "internal/order/handlers", To: "internal/order/repositories"}}
	known := []exception{
		{Rule: "layer", From: "internal/order/handlers", To: "internal/order/repositories", Reason: "legacy export path", Owner: "orders team", Until: "2026-12-31"},
		{Rule: "layer", From: "internal/order/handlers", To: "internal/order/repositories", Reason: "listed twice", Owner: "orders team"},
		{Rule: "ownership", From: "internal/billing/services", To: "internal/order/repositories", Reason: "fixed last sprint", Owner: "billing team"},
		{Rule: "driver", From: "internal/order/services", To: "net/http"},
	}
	kept, suppressed := applyKnown(known, found)
	if suppressed != 1 {
		t.Fatalf("suppressed = %d, want 1", suppressed)
	}
	got := rules(kept)
	want := []string{
		"duplicate internal/order/handlers -> internal/order/repositories",
		"stale internal/billing/services -> internal/order/repositories",
		"stale internal/order/services -> net/http",
		"unexplained internal/order/services -> net/http",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if kept, suppressed = applyKnown(nil, found); suppressed != 0 || len(kept) != 1 {
		t.Fatalf("empty known list must keep every violation: kept %d, suppressed %d", len(kept), suppressed)
	}
}

func TestParseArgsRejectsBadLimit(t *testing.T) {
	for _, args := range [][]string{{"--limit", "nope"}, {"--limit=-1"}, {"--limit"}, {"--bogus"}} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("parseArgs(%v) accepted", args)
		}
	}
	o, err := parseArgs([]string{"--json", "--limit", "3", "--include-tests", "some/dir"})
	if err != nil || !o.jsonOut || o.limit != 3 || !o.includeTests || o.root != "some/dir" || !strings.HasSuffix(o.configPath, "architecture.json") {
		t.Fatalf("parseArgs = %+v, %v", o, err)
	}
}

// The errors a first run meets name their fix: go.mod is checked before the
// config, and a missing config prints a starter file.
func TestRunExplainsMissingGoModAndConfig(t *testing.T) {
	noModule := t.TempDir()
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "go.mod"), []byte("module example.com/parent\n\ngo 1.27\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(parent, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, root, want string
	}{
		{"no go.mod", noModule, "no go.mod in " + noModule},
		{"go.mod in a parent", sub, "its module's go.mod is in " + parent},
		{"no architecture.json", parent, `{"layout": "layers", "known": []}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run([]string{tt.root}, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("run(%q) = %d, stderr %q; want 2 and %q", tt.root, code, stderr.String(), tt.want)
			}
		})
	}
}

// Each starter the missing-config error prints is a file the checker accepts.
func TestStarterConfigsLoad(t *testing.T) {
	loaded := 0
	for line := range strings.Lines(starterConfig) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		path := filepath.Join(t.TempDir(), "architecture.json")
		if err := os.WriteFile(path, []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := loadConfig(path); err != nil {
			t.Errorf("loadConfig(%s) = %v, want the printed starter accepted", line, err)
		}
		loaded++
	}
	if loaded != 2 {
		t.Errorf("starterConfig holds %d JSON lines, want one per layout (2)", loaded)
	}
}

func TestConfigValidation(t *testing.T) {
	bad := &config{Layout: "hexagonal"}
	bad.defaults()
	if err := bad.validate(); err == nil {
		t.Fatal("unknown layout accepted")
	}
	empty := &config{Layout: "modules"}
	empty.defaults()
	if err := empty.validate(); err == nil {
		t.Fatal(`layout "modules" without modules accepted`)
	}
}
