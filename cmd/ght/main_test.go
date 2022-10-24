package main_test

import (
	"bytes"
	"math"
	"strings"
	"testing"

	mainapp "github.com/cardil/ghet/cmd/ght"
	"github.com/cardil/ghet/internal/ght"
	"github.com/wavesoftware/go-commandline"
	"gotest.tools/v3/assert"
)

func TestMainFunc(t *testing.T) {
	retcode := math.MinInt64
	defer func() {
		ght.Options = nil
	}()
	var buf bytes.Buffer
	ght.Options = []commandline.Option{
		commandline.WithExit(func(code int) {
			retcode = code
		}),
		commandline.WithOutput(&buf),
		commandline.WithArgs(""),
	}

	mainapp.Main()

	out := buf.String()
	assert.Check(t, strings.Contains(out, "Gʰet artifacts from GitHub releases"))
	assert.Check(t, retcode == math.MinInt64)
}
