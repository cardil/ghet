package tui

import (
	"context"
	"fmt"

	"github.com/cardil/ghet/pkg/output"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const spinnerColor = lipgloss.Color("205")

type NewSpinnerFunc func(ctx context.Context, message string) Spinner

type Spinner interface {
	Runnable[Spinner]
}

func NewBubbleSpinner(ctx context.Context, message string) Spinner {
	return &BubbleSpinner{
		InputOutput: output.PrinterFrom(ctx),
		Message:     message,
	}
}

var _ NewSpinnerFunc = NewBubbleSpinner

type BubbleSpinner struct {
	output.InputOutput
	Message string

	spin spinner.Model
	tea  *tea.Program
}

func (b *BubbleSpinner) With(fn func(Spinner) error) error {
	b.start()
	defer b.stop()
	return fn(b)
}

func (b *BubbleSpinner) Init() tea.Cmd {
	return b.spin.Tick
}

func (b *BubbleSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, c := b.spin.Update(msg)
	b.spin = m
	return b, c
}

func (b *BubbleSpinner) View() string {
	return fmt.Sprintf("%s %s", b.Message, b.spin.View())
}

func (b *BubbleSpinner) start() {
	b.spin = spinner.New(
		spinner.WithSpinner(spinner.Meter),
		spinner.WithStyle(spinnerStyle()),
	)
	b.tea = tea.NewProgram(b,
		tea.WithInput(b.InOrStdin()),
		tea.WithOutput(b.OutOrStdout()),
	)
	go func() {
		t := b.tea
		_, _ = t.Run()
		_ = t.ReleaseTerminal()
	}()
}

func (b *BubbleSpinner) stop() {
	if b.tea == nil {
		return
	}
	b.tea.Quit()
	b.tea = nil
	endMsg := fmt.Sprintf("%s %s\n",
		b.Message, spinnerStyle().Render("Done"))
	_, _ = b.OutOrStdout().Write([]byte(endMsg))
}

func spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(spinnerColor)
}
