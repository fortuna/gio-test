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
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

func main() {
	go func() {
		err := func() error {
			w := app.NewWindow(app.Title("Domain Lookup"), app.Size(500, 400))

			var ops op.Ops
			th := material.NewTheme()

			modal := component.NewModal()
			appBar := component.NewAppBar(modal)
			appBar.Title = "Domain Lookup"

			var domainInput component.TextField
			domainInput.Helper = "Helper"
			domainInput.Editor.SingleLine = true
			domainInput.Editor.Submit = true
			var lookupButton widget.Clickable
			var aResult, aaaaResult, cnameResult string

			for {
				switch e := w.NextEvent().(type) {
				case system.DestroyEvent:
					return e.Err

				case system.FrameEvent:
					// This graphics context is used for managing the rendering state.
					gtx := layout.NewContext(&ops, e)

					submitted := false
					for _, e := range domainInput.Events() {
						if _, ok := e.(widget.SubmitEvent); ok {
							submitted = true
						}
					}
					for lookupButton.Clicked(gtx) {
						submitted = true
					}

					if submitted {
						domain := strings.TrimSpace(domainInput.Text())
						ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", domain)
						if err != nil {
							aResult = "❌ " + err.Error()
						} else {
							aResult = fmt.Sprint(ips)
						}
						ips, err = net.DefaultResolver.LookupIP(context.Background(), "ip6", domain)
						if err != nil {
							aaaaResult = "❌ " + err.Error()
						} else {
							aaaaResult = fmt.Sprint(ips)
						}
						cname, err := queryCNAME(context.Background(), domain)
						if err != nil {
							cnameResult = "❌ " + err.Error()
						} else {
							cnameResult = cname
						}
					}

					layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return appBar.Layout(gtx, th, "", "")
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
											layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
												return domainInput.Layout(gtx, th, "")
											}),
											layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
											layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												return material.Button(th, &lookupButton, "Lookup").Layout(gtx)
											}),
										)
									}),
									layout.Rigid(component.Divider(th).Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										label := material.Body1(th, "A")
										label.Font.Weight = font.Bold
										return label.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										var resultLabel = material.Body1(th, aResult)
										resultLabel.Font.Typeface = "monospace"
										return resultLabel.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										label := material.Body1(th, "AAAA")
										label.Font.Weight = font.Bold
										return label.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										var resultLabel = material.Body1(th, aaaaResult)
										resultLabel.Font.Typeface = "monospace"
										return resultLabel.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										label := material.Body1(th, "CNAME")
										label.Font.Weight = font.Bold
										return label.Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										var resultLabel = material.Body1(th, cnameResult)
										resultLabel.Font.Typeface = "monospace"
										return resultLabel.Layout(gtx)
									}),
								)
							})
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
