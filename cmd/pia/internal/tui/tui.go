package tui

import (
	"bytes"
	"github.com/crookdc/pia"
	"github.com/crookdc/pia/squeak"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
	"io"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	resolver pia.KeyResolver
	*tview.Application
	pages   *tview.Pages
	console *console
	content *content
	finder  *finder
	history *history
}

func (a *App) view(path string) {
	tx, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer tx.Close()
	src, err := io.ReadAll(pia.WrapReader(a.resolver, tx))
	if err != nil {
		panic(err)
	}
	a.display(string(src))
}

func (a *App) execute(path string) {
	cfg, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	tx, err := pia.ParseTransaction(filepath.Dir(path), pia.WrapReader(a.resolver, bytes.NewReader(cfg)))
	if err != nil {
		panic(err)
	}
	res, err := tx.Execute(squeak.NewInterpreter(tx.WD, a.console.log))
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBufferString("")
	if err := ResponseFormatter(buf, res); err != nil {
		panic(err)
	}
	text := buf.String()
	a.history.push(entry{
		method:    tx.Method,
		endpoint:  tx.URL.Target,
		timestamp: time.Now(),
		text:      text,
	})
	a.display(text)
}

func (a *App) display(text string) {
	a.content.text.SetText(text)
	a.pages.SwitchToPage("content")
}

func (a *App) input(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEsc {
		a.pages.SwitchToPage("dashboard")
		return nil
	}
	switch ev.Rune() {
	case 'h':
		a.history.enter()
		a.pages.SwitchToPage("history")
		return nil
	case 'f':
		a.pages.SwitchToPage("finder")
		return nil
	case 'c':
		a.console.enter()
		if a.pages.HasPage("console") {
			a.pages.RemovePage("console")
		} else {
			a.pages.AddPage("console", a.console.root(), true, true)
		}
		return nil
	default:
		return ev
	}
}

func Run(wd string, props map[string]string) error {
	if err := clipboard.Init(); err != nil {
		return err
	}
	app := App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
		console:     newConsole(bytes.NewBufferString("")),
		content:     newContent(),
		finder:      newFinder(wd),
		history:     newHistory(128),
		resolver: pia.DelegatingKeyResolver{
			Delegates: map[string]pia.KeyResolver{
				"env":   pia.EnvironmentResolver{},
				"props": pia.MapResolver(props),
			},
		},
	}
	app.history.viewCallback = func(e *entry) {
		app.display(e.text)
	}
	app.finder.executeCallback = app.execute
	app.finder.viewCallback = app.view
	app.pages.AddPage("dashboard", tview.NewTextView().SetText(`
	
	pia - the postman alternative for technical people. 

	Usage:
	f - open finder window
		x - execute currently selected file
			y - copy output to clipboard
		v - view file contents after preprocessing
			y - copy output to clipboard
	h - open history
	c - toggle console

	<ESC> brings you back here.

	created by E. R. Nilsson @ github.com/ernilsson
	`), true, true)
	app.pages.AddPage("finder", app.finder.root(), true, false)
	app.pages.AddPage("content", app.content.root(), true, false)
	app.pages.AddPage("history", app.history.root(), true, false)
	app.SetInputCapture(app.input)
	return app.SetRoot(app.pages, true).Run()
}
