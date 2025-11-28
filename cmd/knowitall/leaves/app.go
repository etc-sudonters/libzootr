package leaves

import (
	"context"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/slipup"
)

func NewApp(ctx context.Context, std *dontio.Std, mount tea.Model) App {
	if mount == nil {
		panic("cannot mount nil model")
	}

	ctx, cancel := context.WithCancelCause(ctx)

	return App{mounted: mount, std: std, ctx: ctx, cancelCause: cancel}
}

type App struct {
	mounted     tea.Model
	std         *dontio.Std
	ctx         context.Context
	cancelCause context.CancelCauseFunc
	Err         error
	statusLine  appStatusLine
}

type appStatusLine struct {
	msg   string
	last  time.Time
	style lipgloss.Style
}

func (this *appStatusLine) tick(now time.Time) {
	if this.msg != "" && time.Now().Sub(this.last).Seconds() > 5 {
		this.msg = ""
		this.last = time.Time{}
	}
}

func (this *appStatusLine) write(msg string, when time.Time) {
	this.msg = msg
	this.last = when
}

func (this *appStatusLine) view() string {
	if this.msg != "" {
		return fmt.Sprintf("%s: %s", this.last.Format(time.TimeOnly), this.msg)
	}
	return ""
}

type StatusMsg string
type StdOutMsg string
type StdErrMsg string
type Wrote struct {
	Err error
	N   int
}

func WriteToStdOutF(msg string, v ...any) tea.Cmd {
	return func() tea.Msg {
		return StdOutMsg(fmt.Sprintf(msg, v...))
	}
}

func WriteToStdErrF(msg string, v ...any) tea.Cmd {
	return func() tea.Msg {
		return StdErrMsg(fmt.Sprintf(msg, v...))
	}
}

func WriteStatusMsg(msg string, v ...any) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(fmt.Sprintf(msg, v...))
	}
}

func writeTo(w io.Writer, msg string) tea.Cmd {
	return func() tea.Msg {
		var wrote Wrote
		wrote.N, wrote.Err = fmt.Fprintln(w, msg)
		return wrote
	}
}

func (this App) Init() tea.Cmd {
	return this.mounted.Init()
}

func (this App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if ctxErr := this.ctx.Err(); ctxErr != nil {
		this.Err = ctxErr
		return this, tea.Batch(tea.Quit, WriteToStdErrF("context canceled: %s", ctxErr))
	}

	this.statusLine.tick(time.Now())

	switch msg := msg.(type) {
	case StatusMsg:
		this.statusLine.write(string(msg), time.Now())
		return this, nil
	case StdErrMsg:
		return this, writeTo(this.std.Err, string(msg))
	case StdOutMsg:
		return this, writeTo(this.std.Out, string(msg))
	case Wrote:
		if msg.Err != nil {
			err := slipup.Describe(msg.Err, "failed to write")
			this.Err = err
			this.cancelCause(err)
			return this, tea.Batch(tea.Quit, cmd)
		}
		return this, nil
	case tea.WindowSizeMsg:
		cmd = WriteToStdErrF("resized: %dx%d", msg.Width, msg.Height)
		break
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return this, tea.Batch(tea.Quit, WriteToStdErrF("ctrl+c interrupt, exiting immediately"))
		}
	}

	var mountedCmd tea.Cmd
	this.mounted, mountedCmd = this.mounted.Update(msg)
	return this, tea.Batch(cmd, mountedCmd)
}

func (this App) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, this.mounted.View(), this.statusLine.view())
}
