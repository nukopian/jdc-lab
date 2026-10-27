// Command pipeline runs every stage on every sample of ads, and compares
// two such runs, for improving the stages (see docs/PIPELINE.md). A
// sample is a .jsonl file in data/ingestion; with more than one, all of
// them together are a sample too, named all.
//
//	pipeline run RUN
//
// runs cmd/blocks, cmd/heading-counts, cmd/sections, cmd/boilerplate and
// cmd/view on each sample, built afresh, so the code as it stands is what runs. Each
// stage's tables go in data/runs/RUN/SAMPLE/STAGE/ and its summary in
// data/runs/RUN/SAMPLE/STAGE.txt. A run of the same name is replaced.
// It ends with how long each stage took on each sample, and how much
// memory it used, which it also keeps in data/runs/RUN/timings.tsv.
//
//	pipeline compare BEFORE AFTER
//
// shows, for each sample, the summary lines whose figures moved between
// two runs, and the headings newly mapped and newly unmapped.
//
// From the lab's folder:
//
//	go run ./cmd/pipeline run before
//
// or, once installed with `go install ./cmd/...`, just pipeline.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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

// stages are the programs run on each sample, each with its folder.
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
	var err error
	switch {
	case len(os.Args) == 3 && os.Args[1] == "run":
		err = run(os.Args[2])
	case len(os.Args) == 4 && os.Args[1] == "compare":
		err = compare(os.Stdout, runsDir, os.Args[2], os.Args[3])
	default:
		log.Fatal("usage: pipeline run RUN | pipeline compare BEFORE AFTER")
	}
	if err != nil {
		log.Fatal(err)
	}
}

// run runs every stage on every sample, into data/runs/name.
func run(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsRune(name, filepath.Separator) {
		return fmt.Errorf("pipeline: %q: a run's name is a plain name, such as before", name)
	}
	samples, err := runs.Samples(ingestion)
	if err != nil {
		return err
	}
	if len(samples) == 0 {
		return fmt.Errorf("pipeline: no samples (.jsonl files) in %s", ingestion)
	}
	dir := filepath.Join(runsDir, name)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if len(samples) > 1 {
		all, err := combine(samples, filepath.Join(dir, "all", "ads.jsonl"))
		if err != nil {
			return err
		}
		samples = append(samples, all)
	}

	// The stages are built once, so that their timings leave out the
	// compiling that go run would do.
	bin, err := build()
	if err != nil {
		return err
	}
	defer os.RemoveAll(bin)

	// The samples run at once, each in a goroutine, its stages in turn;
	// errors.Join gathers what went wrong in any of them. Each sample is
	// one process at a time, but a stage can use several processors, so
	// samples running together slow each other a little.
	start := time.Now()
	var wg sync.WaitGroup
	errs := make([]error, len(samples))
	timings := make([][]runs.Timing, len(samples))
	for i, s := range samples {
		wg.Go(func() { timings[i], errs[i] = stage(bin, s, filepath.Join(dir, s.Name)) })
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return err
	}
	for _, s := range samples {
		fmt.Printf("%-16s %s\n", s.Name, filepath.Join(dir, s.Name))
	}

	all := slices.Concat(timings...)
	fmt.Println()
	if err := runs.WriteTimings(os.Stdout, all); err != nil {
		return err
	}
	fmt.Printf("\n%d samples in %.1fs, on %d processors\n", len(samples), time.Since(start).Seconds(), runtime.NumCPU())
	tsv, err := os.Create(filepath.Join(dir, "timings.tsv"))
	if err != nil {
		return err
	}
	if err := runs.WriteTimingsTSV(tsv, all); err != nil {
		tsv.Close()
		return err
	}
	return tsv.Close()
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

// stage runs each stage, built in bin, on one sample, into dir, and
// returns how each ran.
func stage(bin string, s runs.Sample, dir string) ([]runs.Timing, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	ads, err := s.Ads()
	if err != nil {
		return nil, err
	}
	var timings []runs.Timing
	for _, st := range stages {
		summary, err := os.Create(filepath.Join(dir, st.dir+".txt"))
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(filepath.Join(bin, st.program), s.Path, filepath.Join(dir, st.dir))
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
			return nil, fmt.Errorf("%s on %s: %v\n%s", st.program, s.Name, err, stderr.String())
		}
		ps := cmd.ProcessState
		timings = append(timings, runs.Timing{
			Sample: s.Name, Stage: st.program, Ads: ads, Wall: wall,
			CPU: ps.UserTime() + ps.SystemTime(), PeakRSS: peakRSS(ps),
		})
	}
	return timings, nil
}

// combine writes every sample into one file, as the sample named all.
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
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			out.Close()
			return runs.Sample{}, err
		}
	}
	return runs.Sample{Name: "all", Path: path}, out.Close()
}

// compare writes to w how each sample moved from run before to run after,
// both in dir. A sample in only one of them, such as a new export's, is
// named, with nothing to compare.
func compare(w io.Writer, dir, before, after string) error {
	was, err := sampleDirs(filepath.Join(dir, before))
	if err != nil {
		return err
	}
	now, err := sampleDirs(filepath.Join(dir, after))
	if err != nil {
		return err
	}
	for _, name := range now {
		if !slices.Contains(was, name) {
			fmt.Fprintf(w, "== %s: new in %s, not in %s\n\n", name, after, before)
			continue
		}
		fmt.Fprintf(w, "== %s\n", name)
		if err := compareSample(w, filepath.Join(dir, before, name), filepath.Join(dir, after, name)); err != nil {
			return err
		}
		fmt.Fprintln(w)
	}
	for _, name := range was {
		if !slices.Contains(now, name) {
			fmt.Fprintf(w, "== %s: in %s, gone from %s\n\n", name, before, after)
		}
	}
	return nil
}

// sampleDirs returns the names of a run's samples: its folders.
func sampleDirs(run string) ([]string, error) {
	entries, err := os.ReadDir(run)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// compareSample writes how one sample moved between its folders in two
// runs.
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

// moved writes the headings newly mapped and unmapped between two runs of
// one sample.
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
