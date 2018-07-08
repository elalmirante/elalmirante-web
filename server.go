package main

import (
	"net/http"
	"text/template"

	"github.com/elalmirante/elalmirante/config"
	"github.com/elalmirante/elalmirante/models"
	"github.com/elalmirante/elalmirante/query"
)

func createMux() *http.ServeMux {
	mux := http.NewServeMux()

	// assets handler
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

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

func renderDeploy(w http.ResponseWriter, serverResults []serverResult, q, ref string) {
	tmpl, err := template.New("query").ParseFiles("views/layout.html.tmpl", "views/deploy.html.tmpl")

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
	tmpl, err := template.New("query").ParseFiles("views/layout.html.tmpl", "views/query.html.tmpl")

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
	tmpl, err := template.New("error").ParseFiles("views/layout.html.tmpl", "views/error.html.tmpl")

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
