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
	"net/url"

	libonionize "github.com/nogoegst/onionize/lib"
	"github.com/nogoegst/ui"
)

func guiMain(paramsCh chan<- libonionize.Parameters, linkChan <-chan url.URL, errChan <-chan error) {
	err := ui.Main(func() {
		var thepath string
		openButton := ui.NewButton("Open")
		onionizeButton := ui.NewButton("onionize")
		onionizeButton.Disable()
		linkEntry := ui.NewEntry()
		linkEntry.SetReadOnly(true)
		slugCheckbox := ui.NewCheckbox("slug")
		slugCheckbox.SetChecked(true)
		zipCheckbox := ui.NewCheckbox("zip")
		pwEntry := ui.NewPasswordEntry()
		box := ui.NewVerticalBox()
		box.Append(openButton, false)

		box.Append(slugCheckbox, false)
		box.Append(zipCheckbox, false)
		box.Append(pwEntry, false)
		box.Append(onionizeButton, false)
		box.Append(linkEntry, false)
		window := ui.NewWindow("onionize", 200, 100, false)
		window.SetChild(box)
		fadeOut := func() {
			openButton.Disable()
			slugCheckbox.Disable()
			zipCheckbox.Disable()
			pwEntry.Disable()
			onionizeButton.Disable()
		}
		fadeIn := func() {
			openButton.Enable()
			slugCheckbox.Enable()
			zipCheckbox.Enable()
			pwEntry.Enable()
			onionizeButton.Enable()
		}

		openButton.OnClicked(func(*ui.Button) {
			filename := ui.OpenFile(window)
			if filename != "" {
				thepath = filename
				openButton.SetText(filename)
				onionizeButton.Enable()
			} else {
				openButton.SetText("Open")
				onionizeButton.Disable()
			}
		})
		onionizeButton.OnClicked(func(*ui.Button) {
			fadeOut()
			p := libonionize.Parameters{
				Debug:           debug,
				ControlPath:     "default://",
				ControlPassword: "",
				Path:            thepath,
				Zip:             zipCheckbox.Checked(),
				Slug:            slugCheckbox.Checked(),
				Passphrase:      pwEntry.Text(),
			}
			paramsCh <- p
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
		go func() {
			for {
				select {
				case link := <-linkChan:
					linkString := link.String()
					ui.QueueMain(func() {
						linkEntry.SetText(linkString)
					})
				case err := <-errChan:
					ui.QueueMain(func() {
						ui.MsgBox(window, "Error", err.Error())
						fadeIn()
					})
				}
			}
		}()
	})
	if err != nil {
		log.Fatal(err)
	}

}
