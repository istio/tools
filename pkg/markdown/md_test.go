package markdown

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			name: "AnalysisMessageWeakSchema",
		},
		{
			name: "links",
		},
		{
			name: "telemetry",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := readInput(tc.name)
			assert.NoError(t, err)
			out, err := readOutput(tc.name)
			assert.NoError(t, err)

			got := Run(in)
			assert.Equal(t, string(out), string(got))
		})
	}
}

func readInput(name string) ([]byte, error) {
	return readFile(fmt.Sprintf("%s.input", name))
}

func readOutput(name string) ([]byte, error) {
	return readFile(fmt.Sprintf("%s.output", name))
}

func readFile(f string) ([]byte, error) {
	return os.ReadFile(path.Join("testdata", f))
}
