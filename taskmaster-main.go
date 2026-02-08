package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/dal"
	"github.com/jamesrr39/taskmaster/db"
	"github.com/jamesrr39/taskmaster/taskrunner"
	"github.com/jamesrr39/taskmaster/webservices"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/joho/godotenv"
)

var app *kingpin.Application

func main() {
	godotenv.Load()
	app = kingpin.New("taskmaster", "")

	setupInit()
	setupListTasks()
	setupRunTask()
	setupGenerateOpenapiSpec()
	setupServe()
	setupUpgradeVersion()
	setupGetTaskRunResult()
	setupGetTaskRunLogs()
	setupVersion()

	kingpin.MustParse(app.Parse(os.Args[1:]))

}

const (
	SpecFormatYAML       = "yaml"
	SpecFormatJSON       = "json"
	SpecFormatJSONPretty = "jsonpretty"
)

var version = "dev"

func setupVersion() {
	cmd := app.Command("version", "")
	cmd.Action(func(pc *kingpin.ParseContext) error {
		fmt.Println(version)
		return nil
	})
}

func setupInit() {
	cmd := app.Command("init", "")
	filePath := addFilePathFlag(cmd)
	cmd.Action(func(pc *kingpin.ParseContext) error {
		err := setupFoldersAndDBAction(*filePath)
		return errorsx.ErrWithStack(err)
	})
}

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
			return errorsx.ErrWithStack(errorsx.Errorf("unknown format type: %q", *format))
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

		slog.Info("serving", "address", makeHttpLink(*addr), "Openapi/Swagger address", fmt.Sprintf("%s/docs", makeHttpLink(*addr)))
		err = server.ListenAndServe()
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		return nil
	})
}

type ListTaskEntry struct {
	Task       *taskrunner.Task
	LatestRuns []*taskrunner.TaskRun
}

// LatestEntry returns the latest entry. Nil if no entries found for this task.
func (e ListTaskEntry) LatestEntry() *taskrunner.TaskRun {
	if len(e.LatestRuns) == 0 {
		return nil
	}

	return e.LatestRuns[0]
}

type SorterFields struct {
	TaskName string
	TaskRun  *taskrunner.TaskRun
}

func sorter(a, b SorterFields) bool {
	if a.TaskRun == nil && b.TaskRun == nil {
		return strings.ToLower(a.TaskName) < strings.ToLower(b.TaskName)
	}

	if a.TaskRun == nil {
		return false
	}

	if b.TaskRun == nil {
		return true
	}

	if a.TaskRun.StartTimestamp != b.TaskRun.StartTimestamp {
		return time.Time(a.TaskRun.StartTimestamp).UnixNano() > time.Time(b.TaskRun.StartTimestamp).UnixNano()
	}

	return strings.ToLower(a.TaskName) < strings.ToLower(b.TaskName)
}

func sortEntries(entries []ListTaskEntry) []ListTaskEntry {
	sort.Slice(entries, func(i, j int) bool {
		iLatestEntry := entries[i].LatestEntry()
		jLatestEntry := entries[j].LatestEntry()
		return sorter(
			SorterFields{TaskName: entries[i].Task.Name, TaskRun: iLatestEntry},
			SorterFields{TaskName: entries[j].Task.Name, TaskRun: jLatestEntry},
		)
	})
	return entries
}

func setupListTasks() {
	cmd := app.Command("list-tasks", "").Alias("ls")
	latestRunsLimit := cmd.Flag("limit", "max number of latest runs to show in the summary").Default("5").Uint()
	filePath := addFilePathFlag(cmd)

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		dbConn, err := getDBConn(*filePath)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)
		tasks, err := taskDAL.GetAll()
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		entries := []ListTaskEntry{}
		for _, task := range tasks {
			latestRuns, err := taskDAL.GetTaskLatestRunsForTask(dbConn, task.Name, *latestRunsLimit)
			if err != nil {
				return errorsx.ErrWithStack(errorsx.Wrap(err, "taskName", task.Name))
			}
			entries = append(
				entries,
				ListTaskEntry{
					Task:       task,
					LatestRuns: latestRuns,
				},
			)
		}

		entries = sortEntries(entries)

		taskEntryRows := []table.Row{}
		for _, entry := range entries {
			latestRunEndText := "Never"
			latestRunID := "Never run"
			var latestRunsTexts []string
			latestEntry := entry.LatestEntry()
			if latestEntry != nil {
				latestRunEndText = time.Time(latestEntry.StartTimestamp).Format(time.RFC1123)

				latestRunID = fmt.Sprintf("#%d", latestEntry.RunNumber)

				for _, run := range entry.LatestRuns {
					latestRunsTexts = append(latestRunsTexts, string(run.State().AsEmoji()))
				}
			}

			taskEntryRows = append(taskEntryRows, table.Row{
				entry.Task.Name,
				latestRunID,
				latestRunEndText,
				strings.Join(latestRunsTexts, " "),
			})
		}

		tw := table.NewWriter()
		tw.AppendHeader(table.Row{"Name", "Last run ID", "Last run finish", "Latest runs"})
		tw.AppendRows(taskEntryRows)
		tw.AppendFooter(table.Row{fmt.Sprintf("Limited to %d rows", *latestRunsLimit)})
		tw.SetIndexColumn(1)
		tw.SetTitle("Tasks")

		fmt.Println(tw.Render())
		return nil
	})
}

func setupRunTask() {
	cmd := app.Command("run-task", "").Alias("run")
	filePath := addFilePathFlag(cmd)
	taskName := cmd.Arg("taskName", "").Required().String()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		dbConn, err := getDBConn(*filePath)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)
		task, err := taskDAL.GetByName(*taskName)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskRun, err := taskDAL.RunTask(dbConn, task)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		switch taskRun.State() {
		case taskrunner.JOB_RUN_STATE_SUCCESS:
			fmt.Printf("%s Task %q finished successfully in %s\n", string(taskRun.State().AsEmoji()), *taskName, getTaskDuration(taskRun))
		case taskrunner.JOB_RUN_STATE_FAILED:
			return fmt.Errorf("%s Task %q failed after %s", string(taskRun.State().AsEmoji()), *taskName, getTaskDuration(taskRun))
		default:
			return errorsx.ErrWithStack(errorsx.Errorf("Task %q finished, but received unknown state: %q", *taskName, taskRun.State()))
		}

		return nil
	})
}

func setupGetTaskRunResult() {
	cmd := app.Command("results", "")
	filePath := addFilePathFlag(cmd)
	taskName := cmd.Flag("task-name", "").String()
	runNumber := cmd.Flag("run-number", "").Uint64()
	limit := cmd.Flag("limit", "max number of latest runs to show in the summary").Default("30").Uint()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		dbConn, err := getDBConn(*filePath)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)

		taskRuns, err := taskDAL.GetTaskRuns(dbConn, *taskName, *runNumber, *limit)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		sort.Slice(taskRuns, func(i, j int) bool {
			a := taskRuns[i]
			b := taskRuns[j]
			return sorter(
				SorterFields{TaskName: a.TaskName, TaskRun: a},
				SorterFields{TaskName: b.TaskName, TaskRun: b},
			)
		})

		var taskEntryRows []table.Row
		for _, taskRun := range taskRuns {
			state := taskRun.State()
			finishedText := "still running..."
			duration := time.Since(time.Time(taskRun.StartTimestamp))
			if state.IsFinished() {
				finishedText = time.Time(*taskRun.EndTimestamp).Format(time.RFC1123)
				duration = getTaskDuration(taskRun)
			}

			taskEntryRows = append(
				taskEntryRows,
				table.Row{
					taskRun.TaskName, fmt.Sprintf("#%d", taskRun.RunNumber), finishedText, duration, string(state.AsEmoji()),
				},
			)
		}

		tw := table.NewWriter()
		tw.AppendHeader(table.Row{"Name", "Last run ID", "Finished", "Duration", "state"})
		tw.AppendRows(taskEntryRows)
		tw.AppendFooter(table.Row{fmt.Sprintf("Limited to %d rows", *limit)})
		tw.SetIndexColumn(1)
		tw.SetTitle("Tasks")

		fmt.Println(tw.Render())

		return nil
	})
}

func setupGetTaskRunLogs() {
	cmd := app.Command("logs", "")
	filePath := addFilePathFlag(cmd)
	taskName := cmd.Arg("taskName", "").Required().String()
	runNumber := cmd.Flag("run-number", "").Uint64()

	cmd.Action(func(pc *kingpin.ParseContext) error {
		var err error

		dbConn, err := getDBConn(*filePath)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

		taskDAL := dal.NewTaskDAL(*filePath, provideNow)

		if *runNumber == 0 {
			taskRuns, err := taskDAL.GetTaskLatestRunsForTask(dbConn, *taskName, 1)
			if err != nil {
				return errorsx.ErrWithStack(errorsx.Wrap(err))
			}
			switch len(taskRuns) {
			case 0:
				return fmt.Errorf("no task runs for '%s'", *taskName)
			case 1:
				runNumber = &taskRuns[0].RunNumber
			default:
				return errorsx.ErrWithStack(errorsx.Errorf("expected 1 result for getting task runs, got %d", len(taskRuns)))
			}
		}

		logFile, err := taskDAL.GetLogsTask(*taskName, *runNumber)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}
		defer logFile.Close()

		_, err = io.Copy(os.Stdout, logFile)
		if err != nil {
			return errorsx.ErrWithStack(errorsx.Wrap(err))
		}

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
		err := setupFoldersAndDBAction(*filePath)
		return errorsx.ErrWithStack(err)
	})
}

func setupFoldersAndDBAction(filePath string) errorsx.Error {
	err := createDirStructure(filePath)
	if err != nil {
		return errorsx.Wrap(err)
	}

	dbFilePath := filepath.Join(filePath, dal.DataFolderName, "taskmaster-db.sqlite3")

	dbc, err := db.OpenDB(dbFilePath)
	if err != nil {
		return errorsx.Wrap(err)
	}

	err = db.RunMigrations(dbc.DB)
	if err != nil {
		return errorsx.Wrap(err)
	}
	return nil
}

type createDirStructureTask func() error

func createDirStructure(baseDir string) errorsx.Error {
	tasks := []createDirStructureTask{
		func() error { return os.MkdirAll(filepath.Join(baseDir, "tasks"), 0755) },
		func() error { return os.MkdirAll(filepath.Join(baseDir, dal.DataFolderName, "results"), 0755) },
	}

	for _, task := range tasks {
		err := task()
		if err != nil {
			return errorsx.Wrap(err)
		}
	}

	return nil
}

func MustJSONPrettyPrint(writer io.Writer, obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		panic(fmt.Sprintf("couldn't pretty print JSON. Error: %s", err))
	}

	_, err = writer.Write(append(b, byte('\n')))
	if err != nil {
		panic(fmt.Sprintf("couldn't write to writer. Error: %s", err))
	}
}

func makeHttpLink(s string) string {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}

	return "http://" + s
}

func getDBConn(filePath string) (db.DBConn, errorsx.Error) {
	dbFilePath := filepath.Join(filePath, dal.DataFolderName, "taskmaster-db.sqlite3")

	dbConn, err := db.OpenDB(dbFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err)
	}

	return dbConn, nil
}

func getTaskDuration(taskRun *taskrunner.TaskRun) time.Duration {
	return time.Time(*taskRun.EndTimestamp).Sub(time.Time(taskRun.StartTimestamp))
}
