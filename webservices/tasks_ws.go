package webservices

import (
	"context"

	"github.com/jamesrr39/go-openapix"

	"github.com/jamesrr39/taskmaster/dal"
	"github.com/jamesrr39/taskmaster/taskrunner"

	"github.com/swaggest/rest/nethttp"
)

type EmptyStruct struct{}

type ListProjectsResponse struct {
	Tasks []*taskrunner.Task `json:"tasks" nullable:"false" required:"true"`
}

func GetAllProjects(d *dal.TaskDAL, baseDir string) *nethttp.Handler {
	return openapix.MustCreateOpenapiEndpoint(
		"Get tasks",
		&openapix.HandlerOptions{},
		func(ctx context.Context, input *EmptyStruct, output *ListProjectsResponse) error {

			tasks, err := d.GetAll()
			if err != nil {
				return err
			}

			output.Tasks = tasks

			return nil
		},
	)
}
