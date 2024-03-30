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

type DomainLookupApp struct {
	domainInput  component.TextField
	lookupButton widget.Clickable
	aResult      string
	aaaaResult   string
	cnameResult  string
}

func NewDomainLookupApp() *DomainLookupApp {
	app := &DomainLookupApp{}
	app.domainInput.Helper = "Helper"
	app.domainInput.Editor.SingleLine = true
	app.domainInput.Editor.Submit = true
	return app
}

func (a *DomainLookupApp) lookupDomain() {
	domain := strings.TrimSpace(a.domainInput.Text())
	ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", domain)
	if err != nil {
		a.aResult = "❌ " + err.Error()
	} else {
		a.aResult = fmt.Sprint(ips)
	}
	ips, err = net.DefaultResolver.LookupIP(context.Background(), "ip6", domain)
	if err != nil {
		a.aaaaResult = "❌ " + err.Error()
	} else {
		a.aaaaResult = fmt.Sprint(ips)
	}
	cname, err := net.DefaultResolver.LookupCNAME(context.Background(), domain)
	if err != nil {
		a.cnameResult = "❌ " + err.Error()
	} else {
		a.cnameResult = cname
	}
}


func (a *DomainLookupApp) inputLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Baseline}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.domainInput.Layout(gtx, th, "")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Button(th, &a.lookupButton, "Lookup").Layout(gtx)
		}),
	)
}

func (a *DomainLookupApp) resultLayout(gtx layout.Context, th *material.Theme, label, result string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			l := material.Body1(th, label)
			l.Font.Weight = font.Bold
			return l.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			resultLabel := material.Body1(th, result)
			resultLabel.Font.Typeface = "monospace"
			return resultLabel.Layout(gtx)
		}),
	)
}

func (a *DomainLookupApp) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			appBar := component.NewAppBar(component.NewModal())
			appBar.Title = "Domain Lookup"
			return appBar.Layout(gtx, th, "", "")
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.inputLayout(gtx, th)
					}),
					layout.Rigid(component.Divider(th).Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.resultLayout(gtx, th, "A", a.aResult)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.resultLayout(gtx, th, "AAAA", a.aaaaResult)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.resultLayout(gtx, th, "CNAME", a.cnameResult)
					}),
				)
			})
		}),
	)
}

func main() {
	go func() {
		err := func() error {
			w := app.NewWindow(app.Title("Domain Lookup"), app.Size(500, 400))

			var ops op.Ops
			th := material.NewTheme()

			app := NewDomainLookupApp()

			for {
				switch e := w.NextEvent().(type) {
				case system.DestroyEvent:
					return e.Err

				case system.FrameEvent:
					gtx := layout.NewContext(&ops, e)

					submitted := false
					for _, e := range app.domainInput.Events() {
						if _, ok := e.(widget.SubmitEvent); ok {
							submitted = true
						}
					}
					for app.lookupButton.Clicked(gtx) {
						submitted = true
					}

					if submitted {
						app.lookupDomain()
					}

					app.Layout(gtx, th)
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