package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/runs"
)

// write puts a run's summaries, unmapped table and list of samples into
// dir/run.
func write(t *testing.T, dir, run, summary, unmapped, samples string) {
	t.Helper()
	d := filepath.Join(dir, run)
	if err := os.MkdirAll(filepath.Join(d, "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, st := range stages {
		text := ""
		if st.dir == "sections" {
			text = summary
		}
		if err := os.WriteFile(filepath.Join(d, st.dir+".txt"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tsv := "ads\tchars\tkey\theading\n" + unmapped
	if err := os.WriteFile(filepath.Join(d, "sections", "unmapped.tsv"), []byte(tsv), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "samples.txt"), []byte(samples), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCompare(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "claude/before", "headings of a known kind:  70 of 100 uses (70%)\n", "39\t81961\tand\tand\n", "ads-1\t2000\n")
	write(t, dir, "claude/after", "headings of a known kind:  75 of 100 uses (75%)\n", "", "ads-1\t2000\n")

	var out strings.Builder
	if err := compare(&out, dir, "claude/before", "claude/after"); err != nil {
		t.Fatalf("compare: %v\n%s", err, out.String())
	}
	got := out.String()
	for _, want := range []string{
		"  - headings of a known kind:  70 of 100 uses (70%)\n  + headings of a known kind:  75 of 100 uses (75%)\n",
		"newly mapped: 1 headings\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("compare's output lacks %q; it was:\n%s", want, got)
		}
	}
	if strings.Contains(got, "different samples") {
		t.Errorf("the runs read the same samples, but compare says not:\n%s", got)
	}

	// Runs over different samples are compared, with a warning first.
	write(t, dir, "latest", "", "", "ads-1\t2000\nads-2\t5000\n")
	out.Reset()
	if err := compare(&out, dir, "claude/before", "latest"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "The runs read different samples") {
		t.Errorf("compare doesn't warn of different samples:\n%s", out.String())
	}
}

func TestRunName(t *testing.T) {
	for name, ok := range map[string]bool{
		"latest": true, "claude/before": true,
		"": false, ".": false, "..": false, "claude/..": false, "/tmp": false, "claude/": false, "a//b": false,
	} {
		if err := checkName(name); (err == nil) != ok {
			t.Errorf("checkName(%q) = %v; want ok %v", name, err, ok)
		}
	}
}

// Samples are combined a line each, even when one lacks a last newline.
func TestCombine(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.jsonl"), filepath.Join(dir, "b.jsonl")
	os.WriteFile(a, []byte(`{"id":"1"}`), 0o644) // no newline at its end
	os.WriteFile(b, []byte("{\"id\":\"2\"}\n"), 0o644)
	if _, err := combine([]runs.Sample{{Name: "a", Path: a}, {Name: "b", Path: b}}, filepath.Join(dir, "all.jsonl")); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "all.jsonl"))
	if string(got) != "{\"id\":\"1\"}\n{\"id\":\"2\"}\n" {
		t.Errorf("got %q", got)
	}
}

// fakeStages writes a shell script into a new folder for each stage,
// which writes its program's name and its arguments' last parts as its
// summary, or fails if it's fail, and returns the folder.
func fakeStages(t *testing.T, fail string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in stages are shell scripts")
	}
	bin := t.TempDir()
	for _, st := range stages {
		script := "#!/bin/sh\necho " + st.program + " $(basename \"$1\") $(basename \"$2\")\n"
		if st.program == fail {
			script = "#!/bin/sh\necho broken >&2\nexit 1\n"
		}
		if err := os.WriteFile(filepath.Join(bin, st.program), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return bin
}

func TestStage(t *testing.T) {
	for _, serial := range []bool{true, false} {
		dir := t.TempDir()
		ads := filepath.Join(dir, "ads.jsonl")
		os.WriteFile(ads, []byte("{}\n{}\n"), 0o644)
		timings, err := stage(fakeStages(t, ""), runs.Sample{Name: "all", Path: ads}, dir, serial)
		if err != nil {
			t.Fatalf("serial %v: %v", serial, err)
		}
		if len(timings) != len(stages) {
			t.Fatalf("serial %v: %d timings, want %d", serial, len(timings), len(stages))
		}
		for i, st := range stages {
			if timings[i].Stage != st.program || timings[i].Ads != 2 {
				t.Errorf("serial %v: timing %d is %s of %d ads, want %s of 2", serial, i, timings[i].Stage, timings[i].Ads, st.program)
			}
			got, _ := os.ReadFile(filepath.Join(dir, st.dir+".txt"))
			if want := st.program + " ads.jsonl " + st.dir + "\n"; string(got) != want {
				t.Errorf("serial %v: %s.txt is %q, want %q", serial, st.dir, got, want)
			}
		}
	}
}

func TestStageFails(t *testing.T) {
	for _, serial := range []bool{true, false} {
		dir := t.TempDir()
		ads := filepath.Join(dir, "ads.jsonl")
		os.WriteFile(ads, []byte("{}\n"), 0o644)
		_, err := stage(fakeStages(t, "sections"), runs.Sample{Name: "all", Path: ads}, dir, serial)
		if err == nil || !strings.Contains(err.Error(), "sections on all") || !strings.Contains(err.Error(), "broken") {
			t.Errorf("serial %v: got %v, want sections' failure and what it said", serial, err)
		}
	}
}
