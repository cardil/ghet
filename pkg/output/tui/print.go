package tui

import (
	"context"
	"strings"

	"github.com/cardil/ghet/pkg/output"
	"github.com/charmbracelet/lipgloss"
)

type PrintfFunc func(ctx context.Context, format string, a ...any)

func FmtPrintfFunc(ctx context.Context, format string, a ...any) {
	printer := output.PrinterFrom(ctx)
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	printer.Printf(format, a...)
}

var _ PrintfFunc = FmtPrintfFunc

type Message struct {
	Text string
	Size int
}

func (m Message) BoundingBoxSize() int {
	mSize := m.TextSize()
	if mSize < m.Size {
		mSize = m.Size
	}
	return mSize
}

func (m Message) TextSize() int {
	return len(m.Text)
}

func helpStyle(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(str)
}

type humanByteSize struct {
	num  float64
	unit string
}

func humanizeBytes(bytes float64, unitSuffix string) humanByteSize {
	num := bytes
	units := []string{
		"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB",
	}
	i := 0
	const kilo = 1024
	for num > kilo && i < len(units)-1 {
		num /= kilo
		i++
	}
	return humanByteSize{
		num:  num,
		unit: units[i] + unitSuffix,
	}
}
