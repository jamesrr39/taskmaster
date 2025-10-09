package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/dal"
	"github.com/jamesrr39/taskmaster/db"
	"github.com/jamesrr39/taskmaster/webservices"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/joho/godotenv"
)

var app *kingpin.Application

func main() {
	godotenv.Load()
	app = kingpin.New("taskmaster", "")

	setupListTasks()
	setupRunTask()
	setupGenerateOpenapiSpec()
	setupServe()
	setupUpgradeVersion()

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
	filePath := addFilePathFlag(cmd)
	addr := cmd.Flag("addr", "").Default("localhost:8080").String()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)

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

func setupListTasks() {
	cmd := app.Command("list-tasks", "")
	filePath := addFilePathFlag(cmd)

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)
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

func setupRunTask() {
	cmd := app.Command("run-task", "")
	filePath := addFilePathFlag(cmd)
	taskName := cmd.Arg("taskName", "").Required().String()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)
		task, err := taskDAL.GetByName(*taskName)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskRun, err := taskDAL.RunTask(task)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		json.NewEncoder(os.Stdout).Encode(taskRun)

		return nil
	})
}

func addFilePathFlag(cmd *kingpin.CmdClause) *string {
	return cmd.Flag("path", "Path to Taskmaster directory").Short('C').Default(".").String()
}

func provideNow() time.Time {
	return time.Now()
}

func setupUpgradeVersion() {
	cmd := app.Command("upgrade", "")
	filePath := addFilePathFlag(cmd)
	cmd.Action(func(pc *kingpin.ParseContext) error {

		err := createDirStructure(*filePath)
		if err != nil {
			return errorsx.ErrWithStack(err)
		}

		err = db.RunMigrations(filepath.Join(*filePath, "data"))
		if err != nil {
			return errorsx.ErrWithStack(err)
		}
		return nil
	})
}

type createDirStructureTask func() error

func createDirStructure(baseDir string) errorsx.Error {
	tasks := []createDirStructureTask{
		func() error { return os.MkdirAll(filepath.Join(baseDir, "tasks"), 0755) },
		func() error { return os.MkdirAll(filepath.Join(baseDir, "data"), 0755) },
		func() error { return os.MkdirAll(filepath.Join(baseDir, "data", "logs"), 0755) },
	}

	for _, task := range tasks {
		err := task()
		if err != nil {
			return errorsx.Wrap(err)
		}
	}

	return nil
}
