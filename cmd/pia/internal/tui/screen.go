package tui

import (
	"bytes"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func newContent() *content {
	c := &content{
		text: tview.NewTextView().SetDynamicColors(true),
	}
	c.text.SetInputCapture(c.input)
	return c
}

type content struct {
	text *tview.TextView
}

func (c *content) root() tview.Primitive {
	return c.text
}

func (c *content) input(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Rune() != 'y' {
		return ev
	}
	clipboard.Write(clipboard.FmtText, []byte(c.text.GetText(true)))
	return nil
}

func newConsole(log *bytes.Buffer) *console {
	return &console{
		text: tview.NewTextView(),
		log:  log,
	}
}

type console struct {
	log  *bytes.Buffer
	text *tview.TextView
}

func (c *console) root() tview.Primitive {
	return c.text
}

func (c *console) enter() {
	c.text.SetText(c.log.String())
}

func newFinder(wd string) *finder {
	root := tview.NewTreeNode(wd).SetColor(tcell.ColorWhiteSmoke)
	f := &finder{
		tree: tview.NewTreeView().SetRoot(root).SetCurrentNode(root),
	}
	f.toggle(root, wd)
	f.tree.SetInputCapture(f.input)
	return f
}

type finder struct {
	tree            *tview.TreeView
	executeCallback func(string)
	viewCallback    func(string)
}

func (f *finder) root() tview.Primitive {
	return f.tree
}

func (f *finder) input(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'v':
		if f.viewCallback == nil {
			return event
		}
		if f.isSelectedNodeDir() {
			return nil
		}
		path := f.tree.GetCurrentNode().GetReference().(string)
		f.viewCallback(path)
		return nil
	case 'x':
		if f.executeCallback == nil {
			return event
		}
		if f.isSelectedNodeDir() {
			return nil
		}
		path := f.tree.GetCurrentNode().GetReference().(string)
		f.executeCallback(path)
		return nil
	case rune(tcell.KeyEnter):
		if !f.isSelectedNodeDir() {
			return nil
		}
		node := f.tree.GetCurrentNode()
		f.toggle(node, node.GetReference().(string))
		return nil
	default:
		return event
	}
}

func (f *finder) isSelectedNodeDir() bool {
	path := f.tree.GetCurrentNode().GetReference().(string)
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	return info.IsDir()
}

func (f *finder) toggle(node *tview.TreeNode, path string) {
	if len(node.GetChildren()) != 0 {
		node.SetExpanded(!node.IsExpanded())
		return
	}
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		yml := strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml")
		if !yml && !file.IsDir() {
			continue
		}
		n := tview.NewTreeNode(file.Name()).
			SetReference(filepath.Join(path, file.Name()))
		if file.IsDir() {
			n.SetColor(tcell.ColorWhite)
		}
		node.AddChild(n)
	}
}

func newHistory(length int) *history {
	return &history{
		transactions: make([]*entry, length),
		list:         tview.NewList(),
	}
}

type entry struct {
	method    string
	endpoint  string
	timestamp time.Time
	text      string
}

type history struct {
	transactions []*entry
	list         *tview.List
	viewCallback func(*entry)
}

func (h *history) root() tview.Primitive {
	return h.list
}

func (h *history) enter() {
	h.list.Clear()
	for _, entry := range h.transactions {
		if entry == nil {
			break
		}
		h.list.AddItem(
			fmt.Sprintf("%s %s", entry.method, entry.endpoint),
			"",
			0,
			func() {
				if h.viewCallback == nil {
					return
				}
				h.viewCallback(entry)
			},
		)
	}
}

func (h *history) push(e entry) {
	for i := 1; i < len(h.transactions); i++ {
		h.transactions[len(h.transactions)-i] = h.transactions[len(h.transactions)-i-1]
	}
	h.transactions[0] = &e
}
