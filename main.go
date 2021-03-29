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

	tpl, err := template.New("html").Parse(html)
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

const html = `
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.jsdelivr.net/npm/chart.js@2.8.0"></script>
  </head>
  <body>
    <div style="width:75%;">
      <canvas id="myChart"></canvas>
    </div>
  </body>
  <script type="text/javascript">
    (function() {
      var labels = [];
      var temperatures = [];
      var humidities = [];

      {{range .Data}}
        labels.push('{{.Date}}');
        temperatures.push('{{.Temperature}}');
        humidities.push('{{.Humidity}}');
      {{end}}
      var lineChartData = {
        labels: labels,
        datasets: [{
          label: 'Temperatures',
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgb(255, 99, 132)',
          fill: false,
          data: temperatures,
          yAxisID: 'y-axis-1',
        }, {
          label: 'Humidities',
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgb(54, 162, 235)',
          fill: false,
          data: humidities,
          yAxisID: 'y-axis-2'
        }]
      };

      var ctx = document.getElementById('myChart').getContext('2d');
      var chart = Chart.Line(ctx, {
        data: lineChartData,
        options: {
          responsive: true,
          hoverMode: 'index',
          stacked: false,
          scales: {
            yAxes: [{
              type: 'linear',
              display: true,
              position: 'left',
              id: 'y-axis-1',
            }, {
              type: 'linear',
              display: true,
              position: 'right',
              id: 'y-axis-2',
              gridLines: {
                drawOnChartArea: false, // only want the grid lines for one axis to show up
              },
            }],
          }
        }
      });
    })();
  </script>
</html>
`
