-- +goose Up
-- +goose StatementBegin
CREATE TABLE task_runs (
    task_name TEXT NOT NULL,
    task_run_number BIGINT NOT NULL,
    start_time BIGINT NOT NULL
);

CREATE TABLE task_runs_results (
    task_name TEXT NOT NULL,
    task_run_number BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    exit_code INT, -- can be null if the task was unable to start
    FOREIGN KEY(task_name, task_run_number) REFERENCES task_runs(task_name, task_run_number)
);

CREATE UNIQUE INDEX unique_idx__task_runs__run_number_task_name ON task_runs (task_run_number, task_name);
CREATE UNIQUE INDEX unique_idx__task_runs_results__run_number_task_name ON task_runs_results (task_run_number, task_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX unique_idx__task_runs_results__run_number_task_name;
DROP INDEX unique_idx__task_runs__run_number_task_name;
DROP TABLE task_runs_results;
DROP TABLE task_runs;
-- +goose StatementEnd
