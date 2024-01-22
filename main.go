// Copyright 2024 Vinicius Fortuna
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		err := func() error {
			w := app.NewWindow(app.Title("Domain Lookup"), app.Size(400, 300))

			var ops op.Ops
			th := material.NewTheme()
			var domainInput widget.Editor
			var lookupButton widget.Clickable
			var resultLabel = material.Body1(th, "")
			resultLabel.Font.Typeface = "monospace"

			// Header.
			title := material.H3(th, "Domain Lookup")
			for {
				switch e := w.NextEvent().(type) {
				case system.DestroyEvent:
					return e.Err
				case system.FrameEvent:
					// This graphics context is used for managing the rendering state.
					gtx := layout.NewContext(&ops, e)

					for lookupButton.Clicked(gtx) {
						ips, err := net.LookupIP(domainInput.Text())
						if err != nil {
							resultLabel.Text = "❌ " + err.Error()
						} else {
							resultLabel.Text = fmt.Sprint(ips)
						}
					}

					layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return title.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body1(th, "Domain")
							label.Font.Weight = font.Bold
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Editor(th, &domainInput, "Enter domain").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Button(th, &lookupButton, "Lookup").Layout(gtx)
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body1(th, "Result")
							label.Font.Weight = font.Bold
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return resultLabel.Layout(gtx)
						}),
					)
					// Pass the drawing operations to the GPU.
					e.Frame(gtx.Ops)
				}
			}
		}()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
