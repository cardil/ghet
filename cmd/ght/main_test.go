package main_test

import (
	"bytes"
	"math"
	"testing"

	mainapp "github.com/cardil/ghet/cmd/ght"
	"github.com/cardil/ghet/internal/ght"
	"github.com/stretchr/testify/assert"
	"github.com/wavesoftware/go-commandline"
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
	assert.Contains(t, out, "Gʰet artifacts from GitHub releases")
	assert.Equal(t, retcode, math.MinInt64)
}
