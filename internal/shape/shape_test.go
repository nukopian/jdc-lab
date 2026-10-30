package shape

import (
	"io"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
)

func TestWriteRefusesAnUnknownStage(t *testing.T) {
	err := Write(io.Discard, ads.Ad{ID: "1", HTML: "<p>Drive.</p>"}, func(string) bool { return false }, "blockz")
	if err == nil || !strings.Contains(err.Error(), `no stage "blockz"`) {
		t.Errorf("Write with stage blockz: error = %v, want one naming it", err)
	}
}

func TestWriteShowsEveryStageByDefault(t *testing.T) {
	var b strings.Builder
	if err := Write(&b, ads.Ad{ID: "1", Title: "Driver", HTML: "<p>Drive.</p>"}, func(string) bool { return false }); err != nil {
		t.Fatal(err)
	}
	for _, s := range Stages {
		if !strings.Contains(b.String(), "\n## "+s+":") {
			t.Errorf("Write with no stages named: no %q stage in\n%s", s, b.String())
		}
	}
}
