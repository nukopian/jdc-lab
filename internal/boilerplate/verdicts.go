package boilerplate

import (
	"bufio"
	"os"
	"strings"

	"github.com/nukopian/jdc-lab/internal/headings"
)

// Verdicts reads a run's repeated.tsv, as cmd/boilerplate writes it, and
// returns the keys (see headings.Key) of the texts it judged boilerplate,
// for programs that use a run's verdicts without reading every ad again.
// Its columns are groups, ads, chars, kinds, verdict and text.
func Verdicts(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	keys := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		f := strings.SplitN(sc.Text(), "\t", 6)
		if len(f) == 6 && f[4] == "boilerplate" {
			keys[headings.Key(f[5])] = true
		}
	}
	return keys, sc.Err()
}
