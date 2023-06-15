package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cardil/ghet/pkg/output"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const speedInterval = time.Second / 5

type NewProgressFunc func(ctx context.Context, totalSize int, message Message) Progress

type Progress interface {
	Runnable[ProgressControl]
}

type ProgressControl interface {
	io.Writer
	Error(err error)
}

func NewBubbleProgress(ctx context.Context, totalSize int, message Message) Progress {
	return &BubbleProgress{
		InputOutput: output.PrinterFrom(ctx),
		TotalSize:   totalSize,
		Message:     message,
	}
}

var _ NewProgressFunc = NewBubbleProgress

type BubbleProgress struct {
	output.InputOutput
	Message
	TotalSize int

	prog       progress.Model
	tea        *tea.Program
	downloaded int
	speed      int
	prevSpeed  []int
	err        error
	ended      chan struct{}
}

func (b *BubbleProgress) With(fn func(ProgressControl) error) error {
	b.start()
	defer b.stop()
	return fn(b)
}

func (b *BubbleProgress) Error(err error) {
	b.err = err
	b.tea.Send(tea.Quit())
}

func (b *BubbleProgress) Write(bytes []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	noOfBytes := len(bytes)
	b.downloaded += noOfBytes
	b.speed += noOfBytes
	if b.TotalSize > 0 {
		percent := float64(b.downloaded) / float64(b.TotalSize)
		b.onProgress(percent)
	}
	return noOfBytes, nil
}

func (b *BubbleProgress) Init() tea.Cmd {
	return b.tickSpeed()
}

func (b *BubbleProgress) View() string {
	return b.display(b.prog.View()) +
		"\n" + helpStyle("Press Ctrl+C to cancel")
}

func (b *BubbleProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	handle := bubbleProgressHandler{b}
	switch event := msg.(type) {
	case tea.WindowSizeMsg:
		return handle.windowSize(event)

	case tea.KeyMsg:
		return handle.keyPressed(event)

	case speedChange:
		return handle.speedChange()

	case percentChange:
		return handle.percentChange(event)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		return handle.progressFrame(event)

	default:
		return b, nil
	}
}

type bubbleProgressHandler struct {
	*BubbleProgress
}

func (b bubbleProgressHandler) windowSize(event tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	const percentLen = 4
	b.prog.Width = event.Width - len(b.display("")) + percentLen
	return b, nil
}

func (b bubbleProgressHandler) keyPressed(event tea.KeyMsg) (tea.Model, tea.Cmd) {
	if event.Type == tea.KeyCtrlC {
		b.err = context.Canceled
		return b, tea.Quit
	}
	return b, nil
}

func (b bubbleProgressHandler) speedChange() (tea.Model, tea.Cmd) {
	b.prevSpeed = append(b.prevSpeed, b.speed)
	const speedAvgAmount = 4
	if len(b.prevSpeed) > speedAvgAmount {
		b.prevSpeed = b.prevSpeed[1:]
	}
	b.speed = 0
	if b.downloaded < b.TotalSize {
		return b, b.tickSpeed()
	}
	return b, nil
}

func (b bubbleProgressHandler) percentChange(event percentChange) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	cmds = append(cmds, b.prog.SetPercent(float64(event)))

	if event >= 1.0 {
		cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
	}

	return b, tea.Batch(cmds...)
}

func (b bubbleProgressHandler) progressFrame(event progress.FrameMsg) (tea.Model, tea.Cmd) {
	progressModel, cmd := b.prog.Update(event)
	if m, ok := progressModel.(progress.Model); ok {
		b.prog = m
	}
	return b, cmd
}

func (b *BubbleProgress) display(bar string) string {
	const padding = 2
	const pad = " ⋮ "
	paddingLen := padding + b.Message.BoundingBoxSize() - b.Message.TextSize()
	titlePad := strings.Repeat(" ", paddingLen)
	total := humanizeBytes(float64(b.TotalSize), "")
	totalFmt := fmt.Sprintf("%6.2f %-3s", total.num, total.unit)
	return b.Message.Text + titlePad + bar + pad + b.speedFormatted() +
		pad + totalFmt
}

func (b *BubbleProgress) speedFormatted() string {
	s := humanizeBytes(b.speedPerSecond(), "/s")
	return fmt.Sprintf("%6.2f %-5s", s.num, s.unit)
}

func (b *BubbleProgress) speedPerSecond() float64 {
	speed := 0.
	for _, s := range b.prevSpeed {
		speed += float64(s)
	}
	if len(b.prevSpeed) > 0 {
		speed /= float64(len(b.prevSpeed))
	}
	return speed / float64(speedInterval.Microseconds()) *
		float64(time.Second.Microseconds())
}

func (b *BubbleProgress) tickSpeed() tea.Cmd {
	return tea.Every(speedInterval, func(ti time.Time) tea.Msg {
		return speedChange{}
	})
}

func (b *BubbleProgress) start() {
	b.prog = progress.New(progress.WithDefaultGradient())
	b.tea = tea.NewProgram(b,
		tea.WithInput(b.InOrStdin()),
		tea.WithOutput(b.OutOrStdout()),
	)
	b.ended = make(chan struct{})
	go func() {
		t := b.tea
		_, _ = t.Run()
		close(b.ended)
		if err := t.ReleaseTerminal(); err != nil {
			panic(err)
		}
	}()
}

func (b *BubbleProgress) stop() {
	if b.tea == nil {
		return
	}

	<-b.ended
	b.tea = nil
	b.ended = nil
}

func (b *BubbleProgress) onProgress(percent float64) {
	b.tea.Send(percentChange(percent))
}

func finalPause() tea.Cmd {
	const pause = 500 * time.Millisecond
	return tea.Tick(pause, func(_ time.Time) tea.Msg {
		return nil
	})
}

type percentChange float64

type speedChange struct{}
