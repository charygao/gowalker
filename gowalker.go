// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Go Walker is a server that generates Go projects API documentation on the fly.
package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"

	"github.com/go-macaron/i18n"
	"github.com/go-macaron/pongo2"
	"github.com/go-macaron/session"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/Unknwon/gowalker/pkg/context"
	_ "github.com/Unknwon/gowalker/pkg/prometheus"
	"github.com/Unknwon/gowalker/pkg/setting"
	"github.com/Unknwon/gowalker/routes"
	"github.com/Unknwon/gowalker/routes/apiv1"
)

const Version = "2.5.1.1202"

func init() {
	setting.AppVer = Version
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {
	m := macaron.New()
	if !setting.DisableRouterLog {
		m.Use(macaron.Logger())
	}
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			SkipLogging: setting.ProdMode,
		},
	))
	m.Use(macaron.Static("raw",
		macaron.StaticOptions{
			Prefix:      "raw",
			SkipLogging: setting.ProdMode,
		}))
	m.Use(pongo2.Pongoer(pongo2.Options{
		IndentJSON: !setting.ProdMode,
	}))
	m.Use(i18n.I18n())
	m.Use(session.Sessioner())
	m.Use(context.Contexter())
	return m
}

func main() {
	log.Info("Go Walker %s", Version)
	log.Info("Run Mode: %s", strings.Title(macaron.Env))

	m := newMacaron()
	m.Get("/", routes.Home)
	m.Get("/search", routes.Search)
	m.Get("/search/json", routes.SearchJSON)

	m.Group("/api", func() {
		m.Group("/v1", func() {
			m.Get("/badge", apiv1.Badge)
		})
	})

	m.Get("/-/metrics", promhttp.Handler())

	m.Get("/robots.txt", func() string {
		return `User-agent: *
Disallow: /search`
	})
	m.Get("/*", routes.Docs)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", setting.HTTPPort)
	log.Info("Listen: http://%s", listenAddr)
	if err := http.ListenAndServe(listenAddr, m); err != nil {
		log.Fatal(2, "Failed to start server: %v", err)
	}
}
