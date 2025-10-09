-- +goose Up
-- +goose StatementBegin
CREATE TABLE task_runs (
    id BIGINT NOT NULL PRIMARY KEY,
    task_name TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    status INT NOT NULL
);

CREATE TABLE task_runs_results (
    task_run_id BIGINT NOT NULL PRIMARY KEY,
    end_time TIMESTAMP NOT NULL,
    exit_code INT NOT NULL,
    logs_compression TEXT NOT NULL,
    logs_raw_size BIGINT NOT NULL,
    logs_compressed_size BIGINT NOT NULL,
    FOREIGN KEY(task_run_id) REFERENCES task_runs(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE task_runs_results;
DROP TABLE task_runs;
-- +goose StatementEnd
