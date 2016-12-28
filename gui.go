// gui.go - simple GTK3 GUI for onionize.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
// +build gui

package main

import (
	"log"
	"os"

	"github.com/gotk3/gotk3/gtk"
)


const applicationTitle = "onionize"

var win *gtk.Window
var grid *gtk.Grid

func guiMain() {
	gtk.Init(nil)

	var err error
	win, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
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
		urlEntry.SetText(u)
		urlEntry.SetHExpand(true)
		grid.RemoveRow(1)
		grid.RemoveRow(0)
		grid.InsertRow(0)
		grid.Attach(urlEntry, 0, 0, 1, 1)
		grid.ShowAll()
		win.Resize(1,1)
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
	grid.SetRowSpacing(12)
	grid.SetColumnSpacing(12)

	slugChkBox, err := gtk.CheckButtonNewWithLabel("slug")
	if err != nil {
		log.Fatal(err)
	}
	slugChkBox.SetActive(true)
	grid.Attach(slugChkBox, 2, 0, 1, 1)

	combo, err := gtk.ComboBoxTextNew()
	if err != nil {
		log.Fatal(err)
	}
	combo.AppendText("file")
	combo.AppendText("directory")
	combo.AppendText("zip")
	combo.SetActive(0)
	grid.Attach(combo, 0, 0, 1, 1)
	var fchooserBtn *gtk.FileChooserButton

	updateFileChooser := func(pathtype string) {
		var err error
		switch pathtype {
		case "directory":
			fchooserBtn, err = gtk.FileChooserButtonNew("Select a path", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
			if err != nil {
				log.Fatal(err)
			}
		case "file":
			fchooserBtn, err = gtk.FileChooserButtonNew("Select a path", gtk.FILE_CHOOSER_ACTION_OPEN)
			if err != nil {
				log.Fatal(err)
			}
		case "zip":
			fchooserBtn, err = gtk.FileChooserButtonNew("Select a path", gtk.FILE_CHOOSER_ACTION_OPEN)
			if err != nil {
				log.Fatal(err)
			}
			ffilter, err := gtk.FileFilterNew()
			if err != nil {
				log.Fatal(err)
			}
			ffilter.AddPattern("*.zip")
			fchooserBtn.AddFilter(ffilter)
		}
		fchooserBtn.SetHExpand(false)
		w, err := grid.GetChildAt(1, 0)
		if err == nil {
			w.Destroy()
		}
		grid.Attach(fchooserBtn, 1, 0, 1, 1)
		grid.ShowAll()
		win.Resize(1, 1)
	}
	combo.Connect("changed", func() {
		activeText := combo.GetActiveText()
		updateFileChooser(activeText)
	})
	updateFileChooser("file")


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
		combo.SetSensitive(false)
		slugChkBox.SetSensitive(false)
		grid.ShowAll()
		p := Parameters{
			Path: path,
			Zip:  "zip" == combo.GetActiveText(),
			Slug: slugChkBox.GetActive(),
		}
		paramsCh <- p

	})
	grid.Attach(doBtn, 0, 1, 3, 1)


	return &grid.Container.Widget
}

