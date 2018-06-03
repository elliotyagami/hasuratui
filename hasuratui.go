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
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/jroimartin/gocui"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func getHightlightedBlue(sentence string) string {
	return "\033[34;1m" + sentence + "\033[0m"
}
func getHightlightedYellow(sentence string) string {
	return "\033[33;1m" + sentence + "\033[0m"
}
func getHightlightedGreen(sentence string) string {
	return "\033[32;7m" + sentence + "\033[0m"
}
func getHightlightedOutput(sentence string) string {
	return "\033[34;4m" + sentence + "\033[0m"
}
func getHightlightedRed(sentence string) string {
	return "\033[31;1m" + sentence + "\033[0m"
}

type M map[string]interface{}

var direction []M
var (
	viewArr      = []string{"project_list", "link", "documentation", "terminal", "output", "cmd_panel"} //, "status_bar"}
	nextIndex    = 0
	directionMap = map[int]map[string]int{0: {"left": 2, "right": 1, "up": 5, "down": 3}, 1: {"left": 0, "right": 2, "up": 5, "down": 5}, 2: {"left": 1, "right": 0, "up": 4, "down": 4}, 3: {"left": 4, "right": 1, "up": 0, "down": 5}, 4: {"left": 1, "right": 3, "up": 2, "down": 2}, 5: {"left": 3, "right": 4, "up": 1, "down": 0}}
	urlMap       = map[string]string{}
	linkMap      = map[string]string{"Hasura Install": "install", "Hasura login": "login", "discord": "https://discordapp.com/channels/407792526867693568/411785187631038474", "hasura site": "https://hasura.io/", "hasura dashboard": "https://dashboard.hasura.io/clusters"}
	canOpen      = false
	directory    = ""
	cmdList      = []string{"hasura ms list", "hasura api-console", "hasura cluster status", "hasura cluster create --infra free", "hasura cluster list", "hasura cluster set-default"}
)

func getDir() string {
	return fmt.Sprintf("cd %s ; ", directory)
}
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

	// out, err := g.View("terminal")
	// if err != nil {
	// 	return err
	// }
	// fmt.Fprintln(out, fmt.Sprintf("Going from view %s %d", v.Name(), nextIndex))

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}
	return nil
}
func preView(g *gocui.Gui, v *gocui.View) error {
	nextIndex = (nextIndex - 1) % len(viewArr)
	name := viewArr[nextIndex]

	// out, err := g.View("terminal")
	// if err != nil {
	// 	return err
	// }
	// fmt.Fprintln(out, fmt.Sprintf("Going from view %s %d", v.Name(), nextIndex))

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}
	return nil
}
func status(g *gocui.Gui, text string) error {

	out, err := g.View("status_bar")
	if err != nil {
		return err
	}
	out.Clear()
	text = getHightlightedGreen(text)
	fmt.Fprintln(out, fmt.Sprintf("%s", text))
	return nil
}
func statusOutput(g *gocui.Gui, input, text string) error {
	out, err := g.View("output")
	if err != nil {
		return err
	}
	user, _ := user.Current()
	usr := user.Username
	host, _ := os.Hostname()
	host = getHightlightedRed(usr + "@" + host + "\033[30;1m=>!\033[0m")
	input = getHightlightedBlue(input)
	text = getHightlightedOutput(text)
	fmt.Fprintln(out, fmt.Sprintf("%s %s \n%s", host, input, text))
	return nil
}
func getpage(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	v.Editable = true

	if canOpen {
		// statusBlue(g, fmt.Sprintf("Loading page: %s", l))
		l = "https://docs.hasura.io/0.15/manual/" + urlMap[l]
		output, err := exec.Command("sh", "-c", "hscrape "+l).Output()
		_ = status(g, fmt.Sprintf("Loaded page: %s", l))
		if err != nil {
			return err
		}
		v.Clear()
		fmt.Fprintf(v, "%s", output)
		canOpen = false
		return nil
	}
	return nil
}

func documentation(g *gocui.Gui, v *gocui.View) error {
	if !canOpen {
		v.Clear()
		v.Editable = false
		indexEntries := scrapeIndexHasura()
		urlMap = indexEntries
		for k := range indexEntries {
			fmt.Fprintln(v, getHightlightedYellow(k))
		}
		canOpen = true
	}
	return nil
}
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	column1 := int(float64(maxX-2) * .15)
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
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		if _, err = setCurrentViewOnTop(g, "project_list"); err != nil {
			return err
		}
		filename, _ := getFileName()
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			// panic(err)
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
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		for i := range linkMap {
			fmt.Fprintf(v, "%s\n", i)
		}
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

		documentation(g, v)
		// indexEntries = getYellowColored(indexEntries)
	}
	if v, err := g.SetView("status_bar", 1, maxY-cmdBar-1, 1+column1+column2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Status bar"
		v.Editable = true
		v.Wrap = true
		// v.Autoscroll = true
	}
	if v, err := g.SetView("terminal", 1+column1+column2, 1, maxX-1, 1+maxY/10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Terminal"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("output", 1+column1+column2, 1+maxY/10, maxX-1, 1+screenQuarter); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Output"
		v.Editable = true
		v.Wrap = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorYellow
		v.SelFgColor = gocui.ColorBlack
		v.Autoscroll = true
	}
	if v, err := g.SetView("cmd_panel", 1+column1+column2, 1+screenQuarter, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Title = "Cmd panel"
		for i := range cmdList {
			fmt.Fprintf(v, "%s\n", cmdList[i])
		}
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

func getLine(g *gocui.Gui, text string) error {

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, text)
		if _, err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}
	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("project_list"); err != nil {
		return err
	}
	return nil
}
func terminalHandler(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	_ = status(g, "Executing commands")
	out, err := exec.Command("sh", "-c", getDir()+l).Output()
	if err != nil {
	}
	_ = statusOutput(g, l, fmt.Sprintf("%s", out))

	return nil
}

func execPreCmd(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	_ = status(g, "Executing popular hasura commands")
	out, err := exec.Command("sh", "-c", getDir()+l).Output()
	if err != nil {
	}
	_ = statusOutput(g, l, fmt.Sprintf("%s", out))

	return nil
}

func selectProject(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	_ = status(g, fmt.Sprintf("Selected %s as hasura project.", l))
	directory = l
	return nil

}
func openlink(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	li := linkMap[l]
	if li != "install" && li != "login" {
		_ = status(g, fmt.Sprintf("Opening (%s) %s in browser.", l, li))
		_, err = exec.Command("sh", "-c", "google-chrome "+li).Output()
		if err != nil {
			return err
		}
	} else if li == "install" {
		if _, err := os.Stat("/usr/local/bin/hasura"); !os.IsNotExist(err) {
			getLine(g, "Already installed")
			return nil
		}
		getLine(g, "Installing")
		go func() error {
			exec.Command("sh", "-c", "curl -L https://cli.hasura.io/install.sh | bash").Output()
			delMsg(g, v)
			return nil
		}()
	} else {
		go func() {
			getLine(g, "login")
			if _, err := os.Stat("/usr/local/bin/hasura"); !os.IsNotExist(err) {
				exec.Command("sh", "-c", "hasura login").Output()
			}
		}()

	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
		return err
	}
	if err := g.SetKeybinding("terminal", gocui.KeyEnter, gocui.ModNone, terminalHandler); err != nil {
		return err
	}
	if err := g.SetKeybinding("project_list", gocui.KeyEnter, gocui.ModNone, selectProject); err != nil {
		return err
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
	if err := g.SetKeybinding("documentation", gocui.KeyEnter, gocui.ModNone, getpage); err != nil {
		return err
	}
	if err := g.SetKeybinding("documentation", gocui.KeyArrowLeft, gocui.ModNone, documentation); err != nil {
		return err
	}
	if err := g.SetKeybinding("link", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("link", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("link", gocui.KeyEnter, gocui.ModNone, openlink); err != nil {
		return err
	}
	if err := g.SetKeybinding("cmd_panel", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("cmd_panel", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("cmd_panel", gocui.KeyEnter, gocui.ModNone, execPreCmd); err != nil {
		return err
	}

	return nil
}
