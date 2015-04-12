package main

import (
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/wilhelm-murdoch/biscuit"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// Name defines the name of this utility
	Name = "biscuit-server"
	// Version defines this utility's current version
	Version = "1.0.0"
	// Simple HTML template for testing panel
	Template = `
	<!DOCTYPE html>
	<html>
		<head>
			<title>{{ .Name }} {{ .Version }} - test suite</title>
		</head>
		<body>
			<h2>Language Detector</h2>
			<form method="post">
				<textarea rows="10" cols="80" name="text">{{ .Text }}</textarea><br />
				<small>testing against: <em>{{ join .Bodies "," }}</em></small>
				<hr />
				<input type="submit" value="Process ..." />
			</form>
			{{ if .Results }}
				{{ $scores := .Scores }}

				<h3>Results:</h3>
				<ul>
					{{ range $result := .Results }}
						<li><code>{{$result}} = {{ index $scores $result }}</code></li>
					{{ end }}
				</ul>
			{{ end }}
		</body>
	</html>`
)

var (
	port    = flag.Int("p", 8001, "server port assignment")
	version = flag.Bool("v", false, "current version of this server")
	dir     = flag.String("d", "./corpora/*.csv", "glob path pointing to stored tables")
	bodies  = []string{}
	models  = make(map[string]*biscuit.Model)
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: biscuitserver [flags]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println(Name, Version)
		os.Exit(0)
	}

	files, err := filepath.Glob(*dir)
	if err != nil {
		log.Println("Could not load bodies from path:", err)
		os.Exit(1)
	}
	var wg sync.WaitGroup

	log.Printf("LOADING %d MODEL(S) ...", len(files))
	for _, file := range files {
		wg.Add(1)

		go loadModel(file, &wg)
	}

	wg.Wait()
}

func loadModel(file string, wg *sync.WaitGroup) {
	defer wg.Done()

	label := filepath.Base(file)[:2]
	bodies = append(bodies, label)

	model, err := biscuit.NewModelFromFile(label, file, 3)
	if err != nil {
		log.Println("Could not create model from body:", err)
		os.Exit(1)
	}

	log.Println("... loaded model:", label)

	models[label] = model
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()
	defer Output(w, bodies, []string{}, map[string]float64{}, "")
}

func Process(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer r.Body.Close()

	r.ParseForm()

	unknown := biscuit.NewModelFromText("unknown", r.Form.Get("text"), 3)

	var modelInstances = make([]*biscuit.Model, 0, len(models))

	for _, v := range models {
		modelInstances = append(modelInstances, v)
	}

	results, scores, err := unknown.MatchReturnAll(modelInstances)
	if err != nil {
		panic(err)
	}

	Output(w, bodies, results, scores, r.Form.Get("text"))
}

func Output(w http.ResponseWriter, bodies, results []string, scores map[string]float64, text string) {
	ctx := struct {
		Bodies  []string
		Name    string
		Version string
		Results []string
		Scores  map[string]float64
		Text    string
	}{
		bodies,
		Name,
		Version,
		results,
		scores,
		text,
	}

	fm := template.FuncMap{
		"join": strings.Join,
	}

	t := template.Must(template.New("form").Funcs(fm).Parse(Template))
	t.Execute(w, ctx)
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", Process)

	log.Printf("Running %s v.%s on %d", Name, Version, *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}
