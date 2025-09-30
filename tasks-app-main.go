package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/dal"
	"github.com/jamesrr39/taskmaster/webservices"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/joho/godotenv"
)

var app *kingpin.Application

func main() {
	godotenv.Load()
	app = kingpin.New("taskmaster", "")

	setupStatus()
	setupGenerateOpenapiSpec()
	setupServe()

	kingpin.MustParse(app.Parse(os.Args[1:]))

}

const (
	SpecFormatYAML       = "yaml"
	SpecFormatJSON       = "json"
	SpecFormatJSONPretty = "jsonpretty"
)

func setupGenerateOpenapiSpec() {

	cmd := app.Command("generate-openapi-spec", "")
	format := cmd.Flag("format", "output format").Short('F').Default(SpecFormatYAML).Enum(SpecFormatYAML, SpecFormatJSON, SpecFormatJSONPretty)
	outputFilePath := cmd.Flag("output", "").Short('O').Default(os.Stdout.Name()).String()
	cmd.Action(func(pc *kingpin.ParseContext) error {
		apiSchema, _ := webservices.CreateApiRouter(nil, "")

		spec := apiSchema.Reflector().Spec

		specMarshalFuncMap := map[string]func() ([]byte, error){
			SpecFormatYAML: spec.MarshalYAML,
			SpecFormatJSON: spec.MarshalJSON,
			SpecFormatJSONPretty: func() ([]byte, error) {
				return json.MarshalIndent(spec, "", "\t")
			},
		}

		specMarshalFunc, ok := specMarshalFuncMap[*format]
		if !ok {
			return errorsx.Errorf("unknown format type: %q", *format)
		}

		specBytes, err := specMarshalFunc()
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		err = os.WriteFile(*outputFilePath, specBytes, 0644)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		return nil
	})
}

func setupServe() {
	cmd := app.Command("serve", "")
	filePath := cmd.Flag("filepath", "").Default(".").String()
	addr := cmd.Flag("addr", "").Default("localhost:8080").String()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		taskDAL := dal.NewTaskDAL(*filePath)

		router, _ := webservices.CreateRouter(taskDAL, *filePath)

		server := &http.Server{
			Addr:           *addr,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		slog.Info("serving", "address", *addr)
		err = server.ListenAndServe()
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		return nil
	})
}

func setupStatus() {
	cmd := app.Command("status", "")
	filePath := cmd.Flag("filepath", "").Default(".").String()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		taskDAL := dal.NewTaskDAL(*filePath)
		tasks, err := taskDAL.GetAll()
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		b, err := json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		os.Stdout.Write(b)
		return nil
	})
}
