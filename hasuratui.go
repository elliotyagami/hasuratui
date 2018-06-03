// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jroimartin/gocui"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func getYellowColored(sentence string) string {
	return "\033[33;1m" + sentence + "\033[0m"
}
func getHightlightedYellow(sentence string) string {
	return "\033[33;7m" + sentence + "\033[0m"
}

type M map[string]interface{}

var direction []M
var (
	viewArr      = []string{"project_list", "documentation", "terminal", "link", "history", "cmd_panel"}
	nextIndex    = 0
	directionMap = map[int]map[string]int{0: {"left": 2, "right": 1, "up": 5, "down": 3}, 1: {"left": 0, "right": 2, "up": 5, "down": 5}, 2: {"left": 1, "right": 0, "up": 4, "down": 4}, 3: {"left": 4, "right": 1, "up": 0, "down": 5}, 4: {"left": 1, "right": 3, "up": 2, "down": 2}, 5: {"left": 3, "right": 4, "up": 1, "down": 0}}
	urlMap       = map[string]string{}
)

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if nextIndex == 2 || nextIndex == 6 {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex = (nextIndex + 1) % len(viewArr)
	name := viewArr[nextIndex]

	out, err := g.View("terminal")
	if err != nil {
		return err
	}
	fmt.Fprintln(out, fmt.Sprintf("Going from view %s %d", v.Name(), nextIndex))

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}
	return nil
}

func getpage(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	out, err := g.View("terminal")
	if err != nil {
		return err
	}
	l = urlMap[l]
	fmt.Fprintln(out, fmt.Sprintf("Going from view %s %s", out.Name(), l))
	// pagecontent := scrapePage(urlMap[l])
	fmt.Fprintf(v, "%s", "pagecontent")
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	column1 := int(float64(maxX-2) * .2)
	column2 := int(float64(maxX-2) * .45)
	screenMiddle := int(float64(maxY-2) * .45)
	cmdBar := int(float64(maxY-2) * .08)
	screenQuarter := int(float64(maxY-2) * .75)
	if v, err := g.SetView("project_list", 1, 1, 1+column1, 1+screenMiddle); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Project list"
		v.Wrap = true
		v.Autoscroll = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorYellow
		v.SelFgColor = gocui.ColorBlack
		if _, err = setCurrentViewOnTop(g, "project_list"); err != nil {
			return err
		}
		filename, _ := getFileName()
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(v, "%s", b)
	}
	if v, err := g.SetView("link", 1, 1+screenMiddle, 1+column1, maxY-cmdBar-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Links"
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("documentation", 1+column1, 1, 1+column1+column2, maxY-cmdBar-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Documentation"
		v.Wrap = true
		v.Autoscroll = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorYellow
		v.SelFgColor = gocui.ColorBlack

		out, err := g.View("documentation")
		if err != nil {
			return err
		}
		indexEntries := scrapeIndexHasura()
		urlMap = indexEntries
		for k := range indexEntries {
			fmt.Fprintln(out, k)
		}
		// indexEntries = getYellowColored(indexEntries)
	}
	if v, err := g.SetView("cmd_panel", 1, maxY-cmdBar-1, 1+column1+column2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Cmd Panel"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("terminal", 1+column1+column2, 1, maxX-1, 1+screenQuarter); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Terminal"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true

		// c := exec.Command("bash")

		// // Start the command with a pty.
		// ptmx, err := pty.Start(c)
		// if err != nil {
		// 	return err
		// }
		// go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
		// go func() { _, _ = io.Copy(os.Stdout, ptmx) }()
	}
	if v, err := g.SetView("history", 1+column1+column2, 1+screenQuarter, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Title = "History"
	}
	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writetoFile(filename, dirname string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		f, err = os.Create(filename)
	}
	if _, err = f.WriteString(dirname + "\n"); err != nil {
		panic(err)
	}
	defer f.Close()
}
func getFileName() (string, string) {
	dirname, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir + "/.hasura/project_list", dirname
}

func main() {
	project := flag.Bool("select", false, "for declaring a project")
	flag.Parse()
	if *project {

		filename, dirname := getFileName()
		fmt.Println("Selecting: " + dirname)
		writetoFile(filename, dirname)
		unqiue(filename)
	} else {

		g, err := gocui.NewGui(gocui.OutputNormal)
		if err != nil {
			log.Panicln(err)
		}
		defer g.Close()

		g.Highlight = true
		g.Cursor = true
		g.SelFgColor = gocui.ColorYellow
		g.SetManagerFunc(layout)

		if err := keybindings(g); err != nil {
			log.Panicln(err)
		}

		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panicln(err)
		}
	}
}
func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}
func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("documentation", gocui.KeyEnter, gocui.ModNone, getpage); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("project_list", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("project_list", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("documentation", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("documentation", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	// if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
	// 	return err
	// }
	// if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
	// 	return err
	// }

	// if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
	// 	return err
	// }
	// if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, getpage); err != nil {
	// 	return err
	// }
	// if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
	// 	return err
	// }

	// if err := g.SetKeybinding("main", gocui.KeyCtrlS, gocui.ModNone, saveMain); err != nil {
	// 	return err
	// }
	// if err := g.SetKeybinding("main", gocui.KeyCtrlW, gocui.ModNone, saveVisualMain); err != nil {
	// 	return err
	// }
	return nil
}
