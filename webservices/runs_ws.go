package webservices

import (
	"context"

	"github.com/jamesrr39/go-openapix"

	"github.com/jamesrr39/taskmaster/dal"
	"github.com/jamesrr39/taskmaster/db"
	"github.com/jamesrr39/taskmaster/taskrunner"

	"github.com/swaggest/rest/nethttp"
)

type ListRunsResponse struct {
	Runs []*taskrunner.TaskRun `json:"runs" nullable:"false" required:"true"`
}

func GetAllRuns(d *dal.TaskDAL, dbConn db.DBConn) *nethttp.Handler {
	return openapix.MustCreateOpenapiEndpoint(
		"Get runs",
		&openapix.HandlerOptions{},
		func(ctx context.Context, input *EmptyStruct, output *ListRunsResponse) error {

			const LIMIT = 10_000

			runs, err := d.GetTaskLatestRuns(dbConn, LIMIT)
			if err != nil {
				return err
			}

			output.Runs = runs

			return nil
		},
	)
}
