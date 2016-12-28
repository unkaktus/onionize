// gui.go - simple GTK3 GUI for onionize.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"log"
	"os"

	"github.com/gotk3/gotk3/gtk"
)


const applicationTitle = "onionize"

var grid *gtk.Grid

func guiMain() {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle(applicationTitle)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetDefaultSize(1, 1)
	win.SetResizable(false)
	w := beforeWidget()
	win.Add(w)

	go func(){
		u := <-urlCh
		urlEntry, err := gtk.EntryNew()
		if err != nil {
			log.Fatal("Unable to create entry:", err)
		}
		urlEntry.SetHExpand(true)
		grid.RemoveRow(0)
		grid.InsertRow(0)
		grid.Attach(urlEntry, 0, 0, 1, 1)
		urlEntry.SetText(u)
		grid.ShowAll()
	}()
	win.ShowAll()

	gtk.Main()
	os.Exit(0)
}

func beforeWidget() *gtk.Widget {
	var err error
	grid, err = gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	fchooserBtn, err := gtk.FileChooserButtonNew("Select a path", gtk.FILE_CHOOSER_ACTION_OPEN)
	if err != nil {
		log.Fatal("Unable to create file button:", err)
	}

	fchooserBtn.SetHExpand(false)
	grid.Attach(fchooserBtn, 0, 0, 1, 1)

	doBtn, err := gtk.ButtonNewWithLabel("onionize")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	doBtn.Connect("clicked", func() {
		path := fchooserBtn.GetFilename()
		if path == "" {
			return
		}
		fchooserBtn.SetSensitive(false)
		doBtn.SetSensitive(false)
		doBtn.SetLabel("onionizing...")
		grid.ShowAll()
		p := Parameters{
			Path: path,
		}
		paramsCh <- p

	})
	grid.Attach(doBtn, 1, 0, 1, 1)


	return &grid.Container.Widget
}

