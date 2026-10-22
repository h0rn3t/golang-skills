// bench-verdict compares two files of Go benchmark output, the base code's
// and the changed code's, and decides whether the change is kept.
// bench-compare.sh --verdict builds both test binaries, interleaves their
// runs, and calls it; docs/SCRIPT_JSON_CONTRACTS.md specifies its JSON. It
// uses only the standard library, so a verdict needs neither benchstat nor a
// network.
package main

import (
	"bufio"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
)

const version = "1.1.0"

const (
	verdictKeep   = "keep"
	verdictRevert = "revert"
	verdictNoisy  = "noisy"
)

const (
	changeBetter   = "better"
	changeWorse    = "worse"
	changeSame     = "same"
	changeUnjudged = "unjudged"
)

// The units the verdict judges. Any other unit (MB/s, a b.ReportMetric) is
// shown and not judged: which direction is better is the benchmark author's.
const (
	unitSec    = "sec/op"
	unitBytes  = "B/op"
	unitAllocs = "allocs/op"
)

// minSamples is the smallest run count per side with a 95% confidence
// interval for the median; at 6 the interval is already the whole range.
const minSamples = 6

type key struct{ name, unit string }

// samples holds every value per benchmark and unit, in first-seen order.
type samples struct {
	order []key
	vals  map[key][]float64
}

type row struct {
	Name        string   `json:"name"`
	Unit        string   `json:"unit"`
	Base        float64  `json:"base"`
	New         float64  `json:"new"`
	BaseCIPct   *float64 `json:"base_ci_pct"`
	NewCIPct    *float64 `json:"new_ci_pct"`
	DeltaPct    *float64 `json:"delta_pct"`
	P           float64  `json:"p"`
	NBase       int      `json:"n_base"`
	NNew        int      `json:"n_new"`
	Significant bool     `json:"significant"`
	Change      string   `json:"change"`
}

type report struct {
	Verdict   string   `json:"verdict"`
	Reason    string   `json:"reason"`
	Threshold float64  `json:"threshold_pct"`
	Alpha     float64  `json:"alpha"`
	Rows      []row    `json:"rows"`
	Unmatched []string `json:"unmatched"`
}

// parse reads benchmark result lines: a name, an iteration count, then value
// and unit pairs. ns/op becomes sec/op, the unit benchstat reports.
func parse(r io.Reader) (samples, error) {
	s := samples{vals: map[key][]float64{}}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 4 || len(f)%2 != 0 || !strings.HasPrefix(f[0], "Benchmark") {
			continue
		}
		if _, err := strconv.Atoi(f[1]); err != nil {
			continue
		}
		for i := 2; i < len(f); i += 2 {
			v, err := strconv.ParseFloat(f[i], 64)
			if err != nil {
				continue
			}
			unit := f[i+1]
			if unit == "ns/op" {
				v, unit = v*1e-9, unitSec
			}
			k := key{f[0], unit}
			if _, seen := s.vals[k]; !seen {
				s.order = append(s.order, k)
			}
			s.vals[k] = append(s.vals[k], v)
		}
	}
	return s, sc.Err()
}

func compare(base, changed samples, threshold, alpha float64) (report, error) {
	rep := report{Threshold: threshold, Alpha: alpha, Rows: []row{}, Unmatched: []string{}}
	for _, k := range base.order {
		if _, ok := changed.vals[k]; !ok {
			rep.Unmatched = append(rep.Unmatched, k.name+" "+k.unit+": base only")
		}
	}
	for _, k := range changed.order {
		x, ok := base.vals[k]
		if !ok {
			rep.Unmatched = append(rep.Unmatched, k.name+" "+k.unit+": new only")
			continue
		}
		y := changed.vals[k]
		if len(x) < minSamples || len(y) < minSamples {
			return rep, fmt.Errorf("%s %s has %d base and %d new samples; a verdict needs at least %d per side", k.name, k.unit, len(x), len(y), minSamples)
		}
		x, y = slices.Sorted(slices.Values(x)), slices.Sorted(slices.Values(y))
		r := row{Name: k.name, Unit: k.unit, Base: median(x), New: median(y), NBase: len(x), NNew: len(y)}
		r.BaseCIPct, r.NewCIPct = ciPct(x, r.Base), ciPct(y, r.New)
		r.DeltaPct = deltaPct(r.Base, r.New)
		r.P = uTest(x, y)
		r.Significant = r.P < alpha
		r.Change = judge(r, threshold)
		rep.Rows = append(rep.Rows, r)
	}
	if len(rep.Rows) == 0 {
		return rep, errors.New("no benchmark appears in both files")
	}
	rep.Verdict, rep.Reason = decide(rep.Rows, threshold)
	return rep, nil
}

func judge(r row, threshold float64) string {
	switch r.Unit {
	case unitSec, unitBytes, unitAllocs:
	default:
		return changeUnjudged
	}
	if !r.Significant || r.New == r.Base {
		return changeSame
	}
	// Allocations clear the threshold like time and bytes: a fix that drops
	// time by 89% and adds 7 map allocations to 6000 is kept. A change from 0
	// has no percentage and always counts.
	if r.DeltaPct != nil && math.Abs(*r.DeltaPct) < threshold {
		return changeSame
	}
	if r.New > r.Base {
		return changeWorse
	}
	return changeBetter
}

// decide reverts on any regression in time or allocations, keeps on any
// improvement, and calls a run noisy when nothing changed by the threshold
// and a time interval is wider than it. More bytes per op alone is not a
// regression: a preallocated buffer trades bytes for allocations.
func decide(rows []row, threshold float64) (verdict, reason string) {
	var worse, better, wide []string
	for _, r := range rows {
		switch {
		case r.Change == changeWorse && r.Unit != unitBytes:
			worse = append(worse, r.Name+" "+r.Unit+" "+formatDelta(r))
		case r.Change == changeBetter:
			better = append(better, r.Name+" "+r.Unit+" "+formatDelta(r))
		case r.Unit == unitSec && (wider(r.BaseCIPct, threshold) || wider(r.NewCIPct, threshold)):
			wide = append(wide, r.Name)
		}
	}
	switch {
	case len(worse) > 0:
		return verdictRevert, "regression: " + strings.Join(worse, "; ")
	case len(better) > 0:
		return verdictKeep, "improved: " + strings.Join(better, "; ")
	case len(wide) > 0:
		return verdictNoisy, fmt.Sprintf("no change of %g%% or more, and the run-to-run spread is wider than that: %s", threshold, strings.Join(wide, ", "))
	}
	return verdictRevert, fmt.Sprintf("no significant change of %g%% or more", threshold)
}

func wider(ci *float64, threshold float64) bool {
	return ci == nil || *ci > threshold
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// ciPct is the half-width of the 95% confidence interval for the median, as a
// percentage of the median. The interval runs between the order statistics
// the binomial distribution picks, as benchstat's does; nil when the sample
// is too small for one or the median is zero and the interval is not.
func ciPct(sorted []float64, med float64) *float64 {
	n := len(sorted)
	k := 0 // the largest k with P(Binomial(n, 1/2) < k) <= 0.025
	var cum float64
	for j := range n {
		cum += binomial(n, j) / math.Exp2(float64(n))
		if cum > 0.025 {
			break
		}
		k = j + 1
	}
	if k == 0 {
		return nil
	}
	half := max(med-sorted[k-1], sorted[n-k]-med)
	switch {
	case half == 0:
		return new(0.0)
	case med == 0:
		return nil
	}
	return new(half / med * 100)
}

func binomial(n, k int) float64 {
	c := 1.0
	for i := 1; i <= k; i++ {
		c = c * float64(n-k+i) / float64(i)
	}
	return c
}

func deltaPct(base, changed float64) *float64 {
	switch {
	case base != 0:
		return new((changed - base) / base * 100)
	case changed == 0:
		return new(0.0)
	}
	return nil // from zero to something: no percentage
}

// uTest returns the two-sided p-value of the Mann-Whitney U test: exact when
// no value repeats and the samples total at most 50, else the normal
// approximation with the tie correction. Allocation counts tie by nature.
func uTest(x, y []float64) float64 {
	type obs struct {
		v    float64
		base bool
	}
	all := make([]obs, 0, len(x)+len(y))
	for _, v := range x {
		all = append(all, obs{v, true})
	}
	for _, v := range y {
		all = append(all, obs{v, false})
	}
	slices.SortFunc(all, func(a, b obs) int { return cmp.Compare(a.v, b.v) })
	var rankSum, ties float64
	for i := 0; i < len(all); {
		j := i
		for j < len(all) && all[j].v == all[i].v {
			j++
		}
		rank := float64(i+1+j) / 2 // the mean of ranks i+1 through j
		for _, o := range all[i:j] {
			if o.base {
				rankSum += rank
			}
		}
		t := float64(j - i)
		ties += t*t*t - t
		i = j
	}
	n1, n2 := len(x), len(y)
	u := rankSum - float64(n1*(n1+1))/2
	if ties == 0 && n1+n2 <= 50 {
		return exactU(n1, n2, int(math.Round(u)))
	}
	n := float64(n1 + n2)
	variance := float64(n1*n2) / 12 * (n + 1 - ties/(n*(n-1)))
	if variance <= 0 {
		return 1 // every value is equal
	}
	z := (math.Abs(u-float64(n1*n2)/2) - 0.5) / math.Sqrt(variance)
	if z <= 0 {
		return 1
	}
	return math.Erfc(z / math.Sqrt2)
}

// exactU is the two-sided p-value of U = u for samples of n1 and n2 distinct
// values. The number of orderings giving each U is a coefficient of the
// Gaussian binomial [n1+n2 choose n1], built one factor at a time; up to 50
// samples every count is an integer a float64 holds exactly.
func exactU(n1, n2, u int) float64 {
	counts := make([]float64, n1*n2+n1+1)
	counts[0] = 1
	for i := 1; i <= n1; i++ {
		// [n2+i choose i] = [n2+i-1 choose i-1] * (1-q^(n2+i)) / (1-q^i)
		top, k := i*n2+i, n2+i
		for d := top; d >= k; d-- {
			counts[d] -= counts[d-k]
		}
		for d := i; d <= top; d++ {
			counts[d] += counts[d-i]
		}
	}
	var total, low, high float64
	for d, c := range counts[:n1*n2+1] {
		total += c
		if d <= u {
			low += c
		}
		if d >= u {
			high += c
		}
	}
	return min(1, 2*min(low, high)/total)
}

func formatValue(v float64, unit string) string {
	if unit == unitSec {
		for _, s := range []struct {
			scale  float64
			suffix string
		}{{1, "s"}, {1e-3, "ms"}, {1e-6, "µs"}} {
			if v >= s.scale {
				return fmt.Sprintf("%.4g%s", v/s.scale, s.suffix)
			}
		}
		return fmt.Sprintf("%.4gns", v/1e-9)
	}
	if math.Abs(v) < 1e15 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf("%.4g", v)
}

func formatCI(ci *float64) string {
	if ci == nil {
		return "±?"
	}
	return fmt.Sprintf("±%.0f%%", *ci)
}

func formatDelta(r row) string {
	switch {
	case !r.Significant:
		return "~"
	case r.DeltaPct == nil:
		return "+inf"
	}
	return fmt.Sprintf("%+.2f%%", *r.DeltaPct)
}

func writeText(w io.Writer, rep report) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "benchmark\tunit\tbase\tnew\tdelta\tp\tn\tchange")
	for _, r := range rep.Rows {
		fmt.Fprintf(tw, "%s\t%s\t%s %s\t%s %s\t%s\t%.3f\t%d+%d\t%s\n",
			r.Name, r.Unit, formatValue(r.Base, r.Unit), formatCI(r.BaseCIPct),
			formatValue(r.New, r.Unit), formatCI(r.NewCIPct), formatDelta(r), r.P, r.NBase, r.NNew, r.Change)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	for _, u := range rep.Unmatched {
		fmt.Fprintf(w, "unmatched: %s\n", u)
	}
	_, err := fmt.Fprintf(w, "verdict: %s (%s); threshold %g%%, alpha %g\n", rep.Verdict, rep.Reason, rep.Threshold, rep.Alpha)
	return err
}

func readSamples(path string) (samples, error) {
	f, err := os.Open(path)
	if err != nil {
		return samples{}, err
	}
	defer f.Close() //nolint:errcheck // read-only file: Close reports nothing parse needs
	return parse(f)
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("bench-verdict", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "print the report as JSON")
	threshold := fs.Float64("threshold", 5, "smallest sec/op, B/op or allocs/op change in percent that counts")
	alpha := fs.Float64("alpha", 0.05, "significance level of the U test")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 2 || *threshold <= 0 || *alpha <= 0 || *alpha >= 1 {
		fmt.Fprintln(stderr, "usage: bench-verdict [--json] [--threshold PCT] [--alpha A] BASE NEW")
		return 2
	}
	base, err := readSamples(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "error: read base: %v\n", err)
		return 2
	}
	changed, err := readSamples(fs.Arg(1))
	if err != nil {
		fmt.Fprintf(stderr, "error: read new: %v\n", err)
		return 2
	}
	rep, err := compare(base, changed, *threshold, *alpha)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	if *jsonOut {
		err = json.NewEncoder(stdout).Encode(rep)
	} else {
		err = writeText(stdout, rep)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: write report: %v\n", err)
		return 2
	}
	if rep.Verdict == verdictKeep {
		return 0
	}
	return 1
}

func main() {
	for _, a := range os.Args[1:] {
		if a == "-v" || a == "--version" {
			fmt.Printf("bench-verdict v%s\n", version)
			return
		}
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
