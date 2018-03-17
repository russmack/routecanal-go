package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/russmack/routecanal-go"
)

var devMode = true

const (
	// Defaults.
	defaultConfigFile = "config.yaml"
)

// config holds the application configuration.
type config struct {
	LogPath string `yaml:"logPath"`
}

var templates = template.Must(
	template.ParseFiles(
		"templates/about.html",
		"templates/items.html",
		"templates/index.html"))

func main() {
	configFilename := defaultConfigFile
	if len(os.Args) > 1 {
		configFilename = os.Args[1]
	}

	err := run(configFilename)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

var cfg config

func run(configFilename string) error {
	// Load the config.
	var err error
	cfg, err = getConfig(configFilename)
	if err != nil {
		return err
	}

	router := routecanal.New()

	cssRoute := routecanal.NewRoute().
		SetPattern("/css/").
		SetHandler(css)

	indexRoute := routecanal.NewRoute().
		SetPattern("/").
		SetHandler(index)

	regexRoute := routecanal.NewRoute().
		SetPattern(`/items/([a-z-0-9]*)/`).
		SetHandler(items)

	aboutRoute := routecanal.NewRoute().
		SetPattern(`/about`).
		SetHandler(about)

	router.AddRoute(cssRoute)
	router.AddRoute(indexRoute)
	router.AddRoute(regexRoute)
	router.AddRoute(aboutRoute)

	port := ":9876"
	log.Println("Listening on ", port)
	http.ListenAndServe(port, router)

	return nil
}

// getConfig reads the config file, unmarshals the YAML, and sets defaults.
func getConfig(configFilename string) (config, error) {
	var cfg config

	// Read the config file.
	raw, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return cfg, err
	}

	// Unmarshal the YAML.
	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("%s: failed to read YAML: %v", configFilename, err)
	}

	//if cfg.LogPath == "" {
	//	cfg.LogPath = defaultLogPath
	//}

	return cfg, nil
}

type responseLogs struct {
	//
}

type PageData struct {
	Title string
	Items []string
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *PageData) {
	if devMode {
		log.Println("Rendering template from file.")
		tmpl := template.Must(template.ParseFiles("templates/" + tmpl + ".html"))
		err := tmpl.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Println("Rendering template from cache.")
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func css(w http.ResponseWriter, req *http.Request, params map[string]string) error {
	fmt.Printf("PARAMS in css handler: %+v\n", params)

	ex, err := os.Executable()
	if err != nil {
		return err
	}
	exPath := filepath.Dir(ex)

	fsp := path.Join(exPath, "/templates/css")

	fs := http.FileServer(http.Dir(fsp))
	sph := http.StripPrefix("/css", fs)
	sph.ServeHTTP(w, req)

	return nil
}

func index(w http.ResponseWriter, req *http.Request, params map[string]string) error {
	p := &PageData{Title: "RouteCanal"}

	renderTemplate(w, "index", p)

	return nil
}

func items(w http.ResponseWriter, req *http.Request, params map[string]string) error {
	fmt.Printf("regex handler: %+v\n", req)
	fmt.Printf("regex handler: %+v\n", params)

	items := []string{"guitar", "bike", "rover", "jet"}

	p := &PageData{
		Title: "RouteCanal",
		Items: items,
	}

	renderTemplate(w, "items", p)

	return nil
}

func about(w http.ResponseWriter, req *http.Request, params map[string]string) error {
	p := &PageData{Title: "RouteCanal"}

	renderTemplate(w, "about", p)

	return nil
}
