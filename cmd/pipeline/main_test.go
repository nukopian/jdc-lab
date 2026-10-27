package main

import (
	"os"
	"path/filepath"
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
