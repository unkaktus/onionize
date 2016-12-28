// gui.go - simple GTK3 GUI for onionize.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
)


const applicationTitle = "onionize"

var urlEntry *gtk.Entry

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

	win.Add(windowWidget())

	go func(){
		u := <-urlCh
		_, err := glib.IdleAdd(urlEntry.SetText, u)
		if err != nil {
			log.Fatal(err)
		}
	}()
	win.ShowAll()

	gtk.Main()
}

func windowWidget() *gtk.Widget {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	fchooserBtn, err := gtk.FileChooserButtonNew("Select a path", gtk.FILE_CHOOSER_ACTION_OPEN)
	if err != nil {
		log.Fatal("Unable to create file button:", err)
	}

	fchooserBtn.SetHExpand(true)
	grid.Attach(fchooserBtn, 0, 0, 1, 1)

	insertBtn, err := gtk.ButtonNewWithLabel("onionize")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	insertBtn.Connect("clicked", func() {
		path := fchooserBtn.GetFilename()
		if path == "" {
			return
		}
		insertBtn.SetSensitive(false)
		grid.ShowAll()
		p := Parameters{
			Path: path,
		}
		paramsCh <- p

	})
	grid.Attach(insertBtn, 1, 0, 1, 1)

	urlEntry, err = gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create entry:", err)
	}
	urlEntry.SetHExpand(true)

	grid.Attach(urlEntry, 0, 1, 2, 1)

	return &grid.Container.Widget
}

