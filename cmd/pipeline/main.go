// Command pipeline runs every stage on the samples of ads together, and
// compares two such runs, for improving the stages (see
// docs/PIPELINE.md). A sample is a .jsonl file in data/ingestion, one for
// each export.
//
//	pipeline run [-serial] RUN
//
// puts every sample into one file, data/runs/RUN/ads.jsonl, and runs
// cmd/blocks, cmd/heading-counts, cmd/sections, cmd/boilerplate and
// cmd/view on it, built afresh, so the code as it stands is what runs.
// The stages run at once, as none reads another's output, so a run takes
// as long as its slowest stage; -serial runs them one at a time, for
// timings that other stages sharing the machine don't sway.
// Each stage's tables go in data/runs/RUN/STAGE/ and its summary in
// data/runs/RUN/STAGE.txt, and data/runs/RUN/samples.txt lists the
// samples read. A run of the same name is replaced. A name can have a
// folder, as claude/before does, to keep runs in groups. It ends with how
// long each stage took, and how much memory it used, which it also keeps
// in data/runs/RUN/timings.tsv.
//
//	pipeline compare BEFORE AFTER
//
// shows the summary lines whose figures moved between two runs, and the
// headings newly mapped and newly unmapped, after a warning if the runs
// read different samples.
//
// From the lab's folder:
//
//	go run ./cmd/pipeline run latest
//
// or, once installed with `go install ./cmd/...`, just pipeline.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nukopian/jdc-lab/internal/runs"
)

// The folders the samples are in, and the runs go in.
const (
	ingestion = "data/ingestion"
	runsDir   = "data/runs"
)

// stages are the programs a run runs, each with its folder.
var stages = []struct{ program, dir string }{
	{"blocks", "blocks"},
	{"heading-counts", "headings"},
	{"sections", "sections"},
	{"boilerplate", "boilerplate"},
	{"view", "view"},
}

// shown is how many headings compare lists each way, of the most text.
const shown = 15

func main() {
	log.SetFlags(0)
	if _, err := os.Stat("go.mod"); err != nil {
		log.Fatal("pipeline: run it from the lab's folder")
	}
	const usage = "usage: pipeline run [-serial] RUN | pipeline compare BEFORE AFTER"
	var err error
	switch {
	case len(os.Args) >= 3 && os.Args[1] == "run":
		// A subcommand's flags come after its name, so it has a FlagSet
		// of its own, parsing what follows.
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		serial := fs.Bool("serial", false, "run the stages one at a time, for timings the others don't sway")
		fs.Usage = func() { log.Print(usage); fs.PrintDefaults() }
		fs.Parse(os.Args[2:])
		if fs.NArg() != 1 {
			fs.Usage()
			os.Exit(2)
		}
		err = run(fs.Arg(0), *serial)
	case len(os.Args) == 4 && os.Args[1] == "compare":
		err = compare(os.Stdout, runsDir, os.Args[2], os.Args[3])
	default:
		log.Fatal(usage)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// checkName reports whether name is a run's name: plain names, such as
// latest, or with a folder, as claude/before, but nothing that leaves
// data/runs.
func checkName(name string) error {
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || strings.ContainsRune(part, filepath.Separator) {
			return fmt.Errorf("pipeline: %q: a run's name is a plain name, such as latest, or one in a folder, such as claude/before", name)
		}
	}
	return nil
}

// run runs every stage on every sample together, into data/runs/name:
// the stages one at a time if serial, or else all at once.
func run(name string, serial bool) error {
	if err := checkName(name); err != nil {
		return err
	}
	samples, err := runs.Samples(ingestion)
	if err != nil {
		return err
	}
	if len(samples) == 0 {
		return fmt.Errorf("pipeline: no samples (.jsonl files) in %s", ingestion)
	}
	dir := filepath.Join(runsDir, filepath.FromSlash(name))
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	all, err := combine(samples, filepath.Join(dir, "ads.jsonl"))
	if err != nil {
		return err
	}
	if err := list(samples, filepath.Join(dir, "samples.txt")); err != nil {
		return err
	}

	// The stages are built once, so that their timings leave out the
	// compiling that go run would do.
	bin, err := build()
	if err != nil {
		return err
	}
	defer os.RemoveAll(bin)

	start := time.Now()
	timings, err := stage(bin, all, dir, serial)
	if err != nil {
		return err
	}
	fmt.Printf("%d samples together, in %s\n\n", len(samples), dir)
	if err := runs.WriteTimings(os.Stdout, timings); err != nil {
		return err
	}
	how := "the stages at once"
	if serial {
		how = "the stages one at a time"
	}
	fmt.Printf("\nin %.1fs, %s, on %d processors\n", time.Since(start).Seconds(), how, runtime.NumCPU())
	tsv, err := os.Create(filepath.Join(dir, "timings.tsv"))
	if err != nil {
		return err
	}
	if err := runs.WriteTimingsTSV(tsv, timings); err != nil {
		tsv.Close()
		return err
	}
	return tsv.Close()
}

// list writes the samples a run read to path, with their ads, one to a
// line.
func list(samples []runs.Sample, path string) error {
	var b strings.Builder
	for _, s := range samples {
		n, err := s.Ads()
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s\t%d\n", s.Name, n)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// build builds the stages' programs into a new temporary folder, and
// returns it.
func build() (string, error) {
	bin, err := os.MkdirTemp("", "pipeline-")
	if err != nil {
		return "", err
	}
	args := []string{"build", "-o", bin + string(filepath.Separator)}
	for _, st := range stages {
		args = append(args, "./cmd/"+st.program)
	}
	cmd := exec.Command("go", args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(bin)
		return "", fmt.Errorf("pipeline: building the stages: %v\n%s", err, stderr.String())
	}
	return bin, nil
}

// stage runs each stage, built in bin, on s, the samples together, into
// dir, one at a time if serial, or else all at once, and returns how each
// ran, in the order of stages.
func stage(bin string, s runs.Sample, dir string, serial bool) ([]runs.Timing, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	ads, err := s.Ads()
	if err != nil {
		return nil, err
	}
	if serial {
		var timings []runs.Timing
		for _, st := range stages {
			t, err := runStage(bin, st.program, st.dir, s, ads, dir)
			if err != nil {
				return nil, err
			}
			timings = append(timings, t)
		}
		return timings, nil
	}

	// Each goroutine writes only its own stage's slot, so they need no
	// lock, and Wait makes their writes visible here.
	timings := make([]runs.Timing, len(stages))
	errs := make([]error, len(stages))
	var wg sync.WaitGroup
	for i, st := range stages {
		wg.Go(func() {
			timings[i], errs[i] = runStage(bin, st.program, st.dir, s, ads, dir)
		})
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return timings, nil
}

// runStage runs program, built in bin, on s, of ads ads, into dir/out,
// with its summary in dir/out.txt, and returns how it ran.
func runStage(bin, program, out string, s runs.Sample, ads int, dir string) (runs.Timing, error) {
	summary, err := os.Create(filepath.Join(dir, out+".txt"))
	if err != nil {
		return runs.Timing{}, err
	}
	cmd := exec.Command(filepath.Join(bin, program), s.Path, filepath.Join(dir, out))
	cmd.Stdout = summary
	var stderr strings.Builder
	cmd.Stderr = &stderr
	start := time.Now()
	err = cmd.Run()
	wall := time.Since(start)
	if cerr := summary.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return runs.Timing{}, fmt.Errorf("%s on %s: %v\n%s", program, s.Name, err, stderr.String())
	}
	ps := cmd.ProcessState
	return runs.Timing{
		Sample: s.Name, Stage: program, Ads: ads, Wall: wall,
		CPU: ps.UserTime() + ps.SystemTime(), PeakRSS: peakRSS(ps),
	}, nil
}

// combine writes every sample into one file, at path, and returns it as
// the sample named all.
func combine(samples []runs.Sample, path string) (runs.Sample, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return runs.Sample{}, err
	}
	out, err := os.Create(path)
	if err != nil {
		return runs.Sample{}, err
	}
	for _, s := range samples {
		in, err := os.Open(s.Path)
		if err != nil {
			out.Close()
			return runs.Sample{}, err
		}
		err = copyLines(out, in)
		in.Close()
		if err != nil {
			out.Close()
			return runs.Sample{}, err
		}
	}
	return runs.Sample{Name: "all", Path: path}, out.Close()
}

// copyLines copies in to out, ending it with a line break if it lacks
// one, so that the next file's first line doesn't join its last.
func copyLines(out io.Writer, in io.Reader) error {
	last := make([]byte, 1)
	n, err := io.Copy(io.MultiWriter(out, lastByte(last)), in)
	if err == nil && n > 0 && last[0] != '\n' {
		_, err = io.WriteString(out, "\n")
	}
	return err
}

// lastByte is a writer that keeps the last byte written to it.
type lastByte []byte

func (l lastByte) Write(p []byte) (int, error) {
	if len(p) > 0 {
		l[0] = p[len(p)-1]
	}
	return len(p), nil
}

// compare writes to w how run after moved from run before, both in dir,
// first warning if they read different samples, as their figures then
// differ for that reason too.
func compare(w io.Writer, dir, before, after string) error {
	was := filepath.Join(dir, filepath.FromSlash(before))
	now := filepath.Join(dir, filepath.FromSlash(after))
	a, errA := os.ReadFile(filepath.Join(was, "samples.txt"))
	b, errB := os.ReadFile(filepath.Join(now, "samples.txt"))
	if errA != nil || errB != nil || string(a) != string(b) {
		fmt.Fprintf(w, "The runs read different samples, so figures differ for that too:\n%s: %s\n%s: %s\n\n",
			before, strings.Join(strings.Fields(string(a)), " "), after, strings.Join(strings.Fields(string(b)), " "))
	}
	return compareSample(w, was, now)
}

// compareSample writes how a run moved between two of its folders.
func compareSample(w io.Writer, was, now string) error {
	for _, st := range stages {
		// A summary names the folder it wrote to, which differs
		// between runs by the run's name alone.
		a, errA := os.ReadFile(filepath.Join(was, st.dir+".txt"))
		b, errB := os.ReadFile(filepath.Join(now, st.dir+".txt"))
		if errA != nil || errB != nil {
			fmt.Fprintf(w, "%s: not in both runs\n", st.program)
			continue
		}
		changes := runs.Changed(strings.ReplaceAll(string(a), was, "…"), strings.ReplaceAll(string(b), now, "…"))
		if len(changes) == 0 {
			fmt.Fprintf(w, "%s: unchanged\n", st.program)
			continue
		}
		fmt.Fprintf(w, "%s:\n", st.program)
		for _, c := range changes {
			if c.Before != "" {
				fmt.Fprintf(w, "  - %s\n", c.Before)
			}
			if c.After != "" {
				fmt.Fprintf(w, "  + %s\n", c.After)
			}
		}
	}
	return moved(w, was, now)
}

// moved writes the headings newly mapped and unmapped between two runs.
func moved(w io.Writer, was, now string) error {
	a, err := runs.Unmapped(filepath.Join(was, "sections", "unmapped.tsv"))
	if err != nil {
		return err
	}
	b, err := runs.Unmapped(filepath.Join(now, "sections", "unmapped.tsv"))
	if err != nil {
		return err
	}
	mapped, unmapped := runs.Moved(a, b)
	for _, list := range []struct {
		what string
		hs   []runs.Heading
	}{{"newly mapped", mapped}, {"newly unmapped", unmapped}} {
		fmt.Fprintf(w, "%s: %d headings\n", list.what, len(list.hs))
		for i, h := range list.hs {
			if i == shown {
				fmt.Fprintf(w, "  … and %d more\n", len(list.hs)-shown)
				break
			}
			fmt.Fprintf(w, "  %6d chars %4d ads  %s\n", h.Chars, h.Ads, h.Text)
		}
	}
	return nil
}
