package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/jszwec/csvutil"
)

type TemplateArgument struct {
	Data []Data
}

type Data struct {
	Date        string  `csv:"Date"`
	Temperature float64 `csv:"temperature"`
	Humidity    int     `csv:"humidity"`
}

const (
	app = "temphum"
)

var (
	flags    *flag.FlagSet
	filename string
	addr     string
)

func setFlags() {
	flags = flag.NewFlagSet(app, flag.ExitOnError)
	flags.StringVar(&addr, "addr", ":8888", "http service address")
}

func main() {
	setFlags()
	flags.Parse(os.Args[1:])

	if flags.NArg() != 1 {
		fmt.Println("please specify filename")
		return
	}
	filename = flags.Args()[0]

	http.HandleFunc("/", handler)
	log.Print("Listening on http://localhost:8888/")
	http.ListenAndServe(":8888", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	html, err := ioutil.ReadFile("index.tmpl")
	if err != nil {
		errorResponse(err, w)
		return
	}

	var data []Data
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		errorResponse(err, w)
		return
	}
	if err = csvutil.Unmarshal(b, &data); err != nil {
		errorResponse(err, w)
		return
	}
	t := TemplateArgument{Data: data}

	tpl, err := template.New("html").Parse(string(html))
	if err != nil {
		errorResponse(err, w)
		return
	}

	buf := new(bytes.Buffer)
	tpl.Execute(buf, t)
	fmt.Fprint(w, buf.String())
}

func errorResponse(err error, w http.ResponseWriter) {
	fmt.Fprintf(w, "Error occurred: %v", err)
}
