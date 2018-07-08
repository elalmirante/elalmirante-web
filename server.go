package main

import (
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/elalmirante/elalmirante/config"
	"github.com/elalmirante/elalmirante/models"
	"github.com/elalmirante/elalmirante/query"
)

var mimeTypes = map[string]string{
	"css": "text/css",
	"js":  "application/javascript",
}

var layoutView, _ = Asset("views/layout.html.tmpl")
var errorView, _ = Asset("views/error.html.tmpl")
var deployView, _ = Asset("views/deploy.html.tmpl")
var queryView, _ = Asset("views/query.html.tmpl")

// string views, from go-bindata
var deployViewStr = string(layoutView) + string(deployView)
var queryViewStr = string(layoutView) + string(queryView)
var errorViewStr = string(layoutView) + string(errorView)

func createMux() *http.ServeMux {
	mux := http.NewServeMux()

	// server assets from go-bindata
	mux.Handle("/assets/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request := strings.TrimPrefix(r.URL.Path, "/")
			file, err := Asset(request)

			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			ext := filepath.Ext(request)
			w.Header().Set("Content-Type", mimeTypes[ext])
			w.Write(file)
		}))

	// serve config file
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, configFile)
	})

	// query handler
	mux.Handle("/query", validateConfig(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		servers, _ := config.GetServersFromConfigFile(configFile)

		// get query value or '*' if doesnt exists.
		q := r.URL.Query().Get("q")
		if q == "" {
			q = "*"
		}

		renderQuery(w, servers, q)
	})))

	// deploy handler
	mux.Handle("/deploy", validateConfig(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		servers, _ := config.GetServersFromConfigFile(configFile)
		q := r.PostFormValue("q")
		ref := r.PostFormValue("ref")

		// If Get, dont show servers
		if r.Method == http.MethodGet {
			servers = nil
		} else {
			servers = query.ExecSorted(servers, q)
		}

		serverResults := deployServers(servers, ref)
		renderDeploy(w, serverResults, q, ref)
	})))

	// index
	mux.Handle("/", validateConfig(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/query", http.StatusTemporaryRedirect)
	})))

	return mux
}

// helper render functions
func renderDeploy(w http.ResponseWriter, serverResults []serverResult, q, ref string) {
	tmpl, err := template.New("query").Parse(deployViewStr)

	if err != nil {
		renderError(w, err)
		return
	}

	data := struct {
		Q       string
		Ref     string
		Results []serverResult
	}{
		Q:       q,
		Ref:     ref,
		Results: serverResults,
	}

	tmpl.ExecuteTemplate(w, "layout", &data)
}

func renderQuery(w http.ResponseWriter, servers []models.Server, q string) {
	tmpl, err := template.New("query").Parse(queryViewStr)

	if err != nil {
		renderError(w, err)
		return
	}

	data := struct {
		Q       string
		Servers []models.Server
	}{
		Q:       q,
		Servers: query.ExecSorted(servers, q),
	}

	tmpl.ExecuteTemplate(w, "layout", &data)
}

func renderError(w http.ResponseWriter, msg error) {
	tmpl, err := template.New("error").Parse(errorViewStr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Error string
	}{
		Error: msg.Error(),
	}

	tmpl.ExecuteTemplate(w, "layout", &data)
}

// middleware to validate configuration.
func validateConfig(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := config.GetServersFromConfigFile(configFile)

		if err != nil {
			renderError(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}
