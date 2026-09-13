// check-architecture checks the production import graph under a Go module's
// internal/ directory against the layered-module policy in
// references/ARCHITECTURE.md. references/ARCHITECTURE-CHECKS.md documents the
// configuration file, the rules, the exit codes, and what the checker does not
// cover. It uses only the standard library and `go list`.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const version = "1.0.0"

// The layer vocabulary is project policy (ARCHITECTURE.md, "The preferred
// layer names"); the checker recognizes exactly these directory names.
const (
	layerHandlers     = "handlers"
	layerServices     = "services"
	layerRepositories = "repositories"
	layerModels       = "models"
)

var layerNames = []string{layerHandlers, layerServices, layerRepositories, layerModels}

// allowedSameModule lists the layers a layer may import inside its own
// module. The module root contract is always allowed; it is not a layer.
var allowedSameModule = map[string][]string{
	layerHandlers:     {layerHandlers, layerServices, layerModels},
	layerServices:     {layerServices, layerModels},
	layerRepositories: {layerRepositories, layerModels},
	layerModels:       {layerModels},
}

var defaultTransportDrivers = []string{
	"net/http", "github.com/gofiber/fiber", "github.com/gin-gonic/gin",
	"github.com/labstack/echo", "github.com/go-chi/chi", "google.golang.org/grpc",
}

var defaultDatabaseDrivers = []string{
	"database/sql", "github.com/jackc/pgx", "gorm.io", "github.com/lib/pq",
	"github.com/jmoiron/sqlx", "go.mongodb.org/mongo-driver", "entgo.io/ent",
}

// config is architecture.json in the module root. Paths in Platform and in
// Known are relative to the module path, e.g. "internal/order/handlers";
// external import paths in Known.To are written as imported.
type config struct {
	Layout      string            `json:"layout"`       // "layers" (shape B) or "modules" (shape C/E)
	Internal    string            `json:"internal"`     // directory holding the checked packages; default "internal"
	App         string            `json:"app"`          // composition-root package under Internal; default "app"
	PlatformDir string            `json:"platform_dir"` // shared-infrastructure directory under Internal; default "platform"
	Platform    []string          `json:"platform"`     // approved shared packages business code may import
	Modules     map[string]string `json:"modules"`      // layout "modules": name -> "layered" | "flat"
	Drivers     struct {
		Transport []string `json:"transport"`
		Database  []string `json:"database"`
	} `json:"drivers"`
	Known []exception `json:"known"`
}

// exception is one reviewed violation the gate tolerates. Every field but
// Until is required; an entry that matches nothing is stale and fails.
type exception struct {
	Rule   string `json:"rule"`
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
	Owner  string `json:"owner"`
	Until  string `json:"until,omitempty"`
}

type kind int

const (
	kindOutside      kind = iota // not under <module>/<internal>/
	kindUnclassified             // under internal/ but neither app, platform, nor a configured module or layer
	kindApp
	kindPlatform
	kindLayer       // a layer package of a layered module, or of the single module in layout "layers"
	kindLayeredRoot // the root contract package of a layered module
	kindFlat        // a flat module's root package or one of its subpackages
)

type class struct {
	kind     kind
	module   string // "" in layout "layers"
	layer    string // one of layerNames for kindLayer
	approved bool   // kindPlatform: on the approved list
	root     bool   // kindFlat: the module's root package
}

type pkgInfo struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

type violation struct {
	Rule    string `json:"rule"`
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Message string `json:"message"`
}

type report struct {
	Module     string      `json:"module"`
	Layout     string      `json:"layout"`
	Checked    []string    `json:"checked"`
	Violations []violation `json:"violations"`
	Total      int         `json:"total"`
	Suppressed int         `json:"suppressed"`
	Truncated  bool        `json:"truncated"`
	Status     string      `json:"status,omitempty"`
}

func (c *config) defaults() {
	if c.Internal == "" {
		c.Internal = "internal"
	}
	if c.App == "" {
		c.App = "app"
	}
	if c.PlatformDir == "" {
		c.PlatformDir = "platform"
	}
	if len(c.Drivers.Transport) == 0 {
		c.Drivers.Transport = defaultTransportDrivers
	}
	if len(c.Drivers.Database) == 0 {
		c.Drivers.Database = defaultDatabaseDrivers
	}
}

func (c *config) validate() error {
	switch c.Layout {
	case "layers", "modules":
	default:
		return fmt.Errorf(`layout must be "layers" or "modules", got %q`, c.Layout)
	}
	for name, mode := range c.Modules {
		if mode != "layered" && mode != "flat" {
			return fmt.Errorf(`module %q: mode must be "layered" or "flat", got %q`, name, mode)
		}
		if strings.Contains(name, "/") {
			return fmt.Errorf("module %q: a module is one directory under %s/", name, c.Internal)
		}
	}
	if c.Layout == "modules" && len(c.Modules) == 0 {
		return errors.New(`layout "modules" needs at least one entry in "modules"`)
	}
	return nil
}

// classify places one import path in the policy. The module prefix is exact:
// a module path that itself contains "/internal/" is never split there.
func classify(cfg *config, modulePath, importPath string) class {
	prefix := modulePath + "/" + cfg.Internal + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return class{kind: kindOutside}
	}
	rel := strings.TrimPrefix(importPath, prefix)
	first, rest, _ := strings.Cut(rel, "/")
	switch first {
	case cfg.App:
		return class{kind: kindApp}
	case cfg.PlatformDir:
		return class{kind: kindPlatform, approved: slices.Contains(cfg.Platform, cfg.Internal+"/"+rel)}
	}
	if cfg.Layout == "layers" {
		if slices.Contains(layerNames, first) {
			return class{kind: kindLayer, layer: first}
		}
		return class{kind: kindUnclassified}
	}
	switch cfg.Modules[first] {
	case "flat":
		return class{kind: kindFlat, module: first, root: rest == ""}
	case "layered":
		if rest == "" {
			return class{kind: kindLayeredRoot, module: first}
		}
		layer, _, _ := strings.Cut(rest, "/")
		if slices.Contains(layerNames, layer) {
			return class{kind: kindLayer, module: first, layer: layer}
		}
	}
	return class{kind: kindUnclassified}
}

func hasDriverPrefix(importPath string, drivers []string) bool {
	for _, d := range drivers {
		if importPath == d || strings.HasPrefix(importPath, d+"/") {
			return true
		}
	}
	return false
}

// display shortens a module package to its module-relative path so that
// output and architecture.json stay portable across module renames.
func display(modulePath, importPath string) string {
	return strings.TrimPrefix(importPath, modulePath+"/")
}

// check walks every production import of every package under internal/ and
// returns the policy violations, before the known list is applied.
func check(cfg *config, modulePath string, pkgs []pkgInfo, includeTests bool) []violation {
	var out []violation
	add := func(rule, from, to, msg string) {
		out = append(out, violation{Rule: rule, From: display(modulePath, from), To: display(modulePath, to), Message: msg})
	}
	for _, p := range pkgs {
		from := classify(cfg, modulePath, p.ImportPath)
		switch from.kind {
		case kindOutside:
			continue
		case kindUnclassified:
			add("unclassified", p.ImportPath, "", fmt.Sprintf("package under %s/ is neither %q, %q, a configured module, nor a layer; add it to architecture.json or move it", cfg.Internal, cfg.App, cfg.PlatformDir))
			continue
		}
		imports := slices.Clone(p.Imports)
		if includeTests {
			imports = append(imports, p.TestImports...)
			imports = append(imports, p.XTestImports...)
		}
		slices.Sort(imports)
		for _, imp := range slices.Compact(imports) {
			if !strings.HasPrefix(imp, modulePath+"/") && imp != modulePath {
				checkDriver(cfg, from, p.ImportPath, imp, add)
				continue
			}
			to := classify(cfg, modulePath, imp)
			checkEdge(cfg, from, to, p.ImportPath, imp, add)
		}
	}
	return out
}

func checkDriver(cfg *config, from class, fromPath, imp string, add func(rule, from, to, msg string)) {
	transport := hasDriverPrefix(imp, cfg.Drivers.Transport)
	database := hasDriverPrefix(imp, cfg.Drivers.Database)
	switch {
	case from.kind == kindLayeredRoot && (transport || database):
		add("driver", fromPath, imp, "a layered module's root contract imports no transport or database driver")
	case from.kind == kindLayer && (from.layer == layerServices || from.layer == layerModels) && (transport || database):
		add("driver", fromPath, imp, from.layer+" import no transport or database driver; handlers own transport, repositories own persistence")
	case from.kind == kindLayer && from.layer == layerHandlers && database:
		add("driver", fromPath, imp, "handlers import no database driver; call a service, which consumes a repository")
	}
}

func checkEdge(cfg *config, from, to class, fromPath, toPath string, add func(rule, from, to, msg string)) {
	if from.kind == kindApp {
		return
	}
	switch to.kind {
	case kindOutside:
		add("unclassified", fromPath, toPath, "module-local target outside "+cfg.Internal+"/ has no ownership class; move it under "+cfg.Internal+"/ or record an exception")
	case kindUnclassified:
		// Reported once, on the target package itself.
	case kindApp:
		add("composition", fromPath, toPath, "business and platform code never import the composition root; it imports them")
	case kindPlatform:
		if from.kind != kindPlatform && !to.approved {
			add("platform", fromPath, toPath, "shared package is not on the approved platform list in architecture.json")
		}
	default: // business target
		if from.kind == kindPlatform {
			add("platform", fromPath, toPath, "platform imports no business module; business depends on platform, never the reverse")
			return
		}
		if from.module != to.module {
			checkCrossModule(from, to, fromPath, toPath, add)
			return
		}
		checkSameModule(from, to, fromPath, toPath, add)
	}
}

func checkCrossModule(from, to class, fromPath, toPath string, add func(rule, from, to, msg string)) {
	consumer := from.kind == kindFlat || (from.kind == kindLayer && from.layer == layerServices)
	switch {
	case to.kind == kindLayeredRoot || (to.kind == kindFlat && to.root):
		if !consumer {
			add("ownership", fromPath, toPath, "a foreign module's root contract is consumed by services (or a flat module), not by "+describe(from))
		}
	default:
		add("ownership", fromPath, toPath, "foreign implementation package; import the module's root contract from services, or declare a consumer-side interface wired in app")
	}
}

func checkSameModule(from, to class, fromPath, toPath string, add func(rule, from, to, msg string)) {
	switch from.kind {
	case kindLayeredRoot:
		add("contract", fromPath, toPath, "a layered module's root contract imports none of its implementation packages")
	case kindLayer:
		if to.kind == kindLayeredRoot {
			return
		}
		if to.kind == kindLayer && slices.Contains(allowedSameModule[from.layer], to.layer) {
			return
		}
		add("layer", fromPath, toPath, from.layer+" may import "+strings.Join(append(allowedSameModule[from.layer], "the module root"), ", ")+"; not "+describe(to))
	}
}

func describe(c class) string {
	switch c.kind {
	case kindLayer:
		return c.layer
	case kindLayeredRoot:
		return "the module root contract"
	case kindFlat:
		return "a flat module package"
	case kindApp:
		return "the composition root"
	case kindPlatform:
		return "a platform package"
	}
	return "an unclassified package"
}

// applyKnown removes reviewed violations and reports every problem with the
// list itself: an entry nothing matches (stale), a duplicate, or one with no
// reason or owner. The list can only shrink through a review, never grow to
// make a run pass.
func applyKnown(known []exception, found []violation) (kept []violation, suppressed int) {
	seen := map[string]int{}
	used := map[string]bool{}
	key := func(rule, from, to string) string { return rule + " " + from + " -> " + to }
	for _, e := range known {
		k := key(e.Rule, e.From, e.To)
		seen[k]++
		if seen[k] == 2 {
			kept = append(kept, violation{Rule: "duplicate", From: e.From, To: e.To, Message: "known entry listed twice: " + k})
		}
		if strings.TrimSpace(e.Reason) == "" || strings.TrimSpace(e.Owner) == "" {
			kept = append(kept, violation{Rule: "unexplained", From: e.From, To: e.To, Message: "known entry needs a reason and an owner: " + k})
		}
	}
	for _, v := range found {
		k := key(v.Rule, v.From, v.To)
		if seen[k] > 0 {
			used[k] = true
			suppressed++
			continue
		}
		kept = append(kept, v)
	}
	for _, e := range known {
		k := key(e.Rule, e.From, e.To)
		if !used[k] && seen[k] > 0 {
			seen[k] = 0 // report once
			kept = append(kept, violation{Rule: "stale", From: e.From, To: e.To, Message: "known entry no longer matches a violation; remove it: " + k})
		}
	}
	return kept, suppressed
}

type options struct {
	root, configPath      string
	jsonOut, includeTests bool
	limit                 int
}

func parseArgs(args []string) (options, error) {
	o := options{root: ".", limit: -1}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			o.jsonOut = true
		case a == "--include-tests":
			o.includeTests = true
		case a == "--limit" || a == "--config":
			if i+1 >= len(args) {
				return o, fmt.Errorf("%s needs a value", a)
			}
			i++
			if a == "--config" {
				o.configPath = args[i]
				continue
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 0 {
				return o, fmt.Errorf("--limit needs a non-negative integer, got %q", args[i])
			}
			o.limit = n
		case strings.HasPrefix(a, "--limit="):
			n, err := strconv.Atoi(strings.TrimPrefix(a, "--limit="))
			if err != nil || n < 0 {
				return o, fmt.Errorf("--limit needs a non-negative integer, got %q", a)
			}
			o.limit = n
		case strings.HasPrefix(a, "--config="):
			o.configPath = strings.TrimPrefix(a, "--config=")
		case strings.HasPrefix(a, "-"):
			return o, fmt.Errorf("unknown flag %s", a)
		default:
			o.root = a
		}
	}
	if o.configPath == "" {
		o.configPath = filepath.Join(o.root, "architecture.json")
	}
	return o, nil
}

func loadConfig(path string) (*config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w (ARCHITECTURE-CHECKS.md documents the file)", path, err)
	}
	var cfg config
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.defaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &cfg, nil
}

func goList(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("go", append([]string{"list", "-mod=readonly"}, args...)...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func loadPackages(root string) (string, []pkgInfo, error) {
	mod, err := goList(root, "-m", "-f", "{{.Path}}")
	if err != nil {
		return "", nil, err
	}
	modulePath := strings.TrimSpace(string(mod))
	raw, err := goList(root, "-json=ImportPath,Imports,TestImports,XTestImports", "./...")
	if err != nil {
		return "", nil, err
	}
	var pkgs []pkgInfo
	dec := json.NewDecoder(bytes.NewReader(raw))
	for {
		var p pkgInfo
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			return "", nil, fmt.Errorf("parse go list output: %w", err)
		}
		pkgs = append(pkgs, p)
	}
	return modulePath, pkgs, nil
}

func run(args []string, stdout, stderr io.Writer) int {
	o, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	cfg, err := loadConfig(o.configPath)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	modulePath, pkgs, err := loadPackages(o.root)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	rep := report{Module: modulePath, Layout: cfg.Layout, Checked: []string{"Imports"}}
	if o.includeTests {
		rep.Checked = append(rep.Checked, "TestImports", "XTestImports")
	}
	internal := 0
	for _, p := range pkgs {
		if classify(cfg, modulePath, p.ImportPath).kind != kindOutside {
			internal++
		}
	}
	if internal == 0 {
		rep.Status = "no_internal_packages"
	}
	found := check(cfg, modulePath, pkgs, o.includeTests)
	kept, suppressed := applyKnown(cfg.Known, found)
	rep.Total, rep.Suppressed = len(kept), suppressed
	if o.limit >= 0 && len(kept) > o.limit {
		kept, rep.Truncated = kept[:o.limit], true
	}
	rep.Violations = kept
	if rep.Violations == nil {
		rep.Violations = []violation{}
	}
	if o.jsonOut {
		enc := json.NewEncoder(stdout)
		if err := enc.Encode(rep); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
	} else {
		for _, v := range rep.Violations {
			if v.To == "" {
				fmt.Fprintf(stdout, "%-12s %s: %s\n", v.Rule, v.From, v.Message)
				continue
			}
			fmt.Fprintf(stdout, "%-12s %s -> %s: %s\n", v.Rule, v.From, v.To, v.Message)
		}
		if rep.Truncated {
			fmt.Fprintf(stdout, "... %d more not shown (--limit)\n", rep.Total-len(rep.Violations))
		}
		fmt.Fprintf(stdout, "%d violation(s), %d suppressed by known entries; module %s, layout %s, checked %s\n",
			rep.Total, rep.Suppressed, rep.Module, rep.Layout, strings.Join(rep.Checked, "+"))
		if rep.Status != "" {
			fmt.Fprintln(stdout, "status:", rep.Status)
		}
	}
	if rep.Total > 0 {
		return 1
	}
	return 0
}

func main() {
	for _, a := range os.Args[1:] {
		switch a {
		case "-h", "--help":
			fmt.Printf("check-architecture v%s - check internal/ imports against the layered-module policy\n\nUSAGE\n    check-architecture [--json] [--limit N] [--include-tests] [--config FILE] [module-root]\n\nExit 0 clean, 1 violations or a stale/duplicate/unexplained known entry, 2 error.\n", version)
			return
		case "-v", "--version":
			fmt.Printf("check-architecture v%s\n", version)
			return
		}
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
