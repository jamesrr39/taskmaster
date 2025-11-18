OPENAPI_SPEC_OUTDIR=web-client/docs
OPENAPI_SPEC_PATH=${OPENAPI_SPEC_OUTDIR}/openapi3-spec.yml

.PHONY: help
help:
	echo "see Makefile"

.PHONY: generate_openapi_typescript
generate_openapi_typescript:
	mkdir -p ${OPENAPI_SPEC_OUTDIR}
	go run taskmaster-main.go generate-openapi-spec --output ${OPENAPI_SPEC_PATH}
	sed -i 's/Domain//g' ${OPENAPI_SPEC_PATH}
	sed -i 's/Webservices//g' ${OPENAPI_SPEC_PATH}
	cd web-client && mkdir -p src/openapi/generated && rm -f src/openapi/generated/* && echo "running yarn generate-openapi" && yarn generate-openapi

.PHONY: run_dev_server
run_dev_server:
	go run taskmaster-main.go serve --path data/localdev

.PHONY: run_dev_client
run_dev_client:
	cd web-client && yarn dev

.PHONY: release
release:
	goreleaser release --clean

.PHONY: release_snapshot
release_snapshot:
	goreleaser build --snapshot --clean

.PHONY: check_modernc_libc_version
check_modernc_libc_version:
# https://pkg.go.dev/modernc.org/sqlite#section-readme
# When you import this package you should use in your go.mod file the exact same version of modernc.org/libc as seen in the go.mod file of this repository.
# grep returns exit code 0 if found, 1 if not found
	grep "modernc.org/libc v1.66.3" go.mod

.PHONY: run_test_job
run_test_job:
	go run taskmaster-main.go run-task --path data/localdev test-job

.PHONY: reset_db
reset_db:
	rm data/localdev/data/taskmaster-db.sqlite3
	go run taskmaster-main.go upgrade --path data/localdev

.PHONY: upgrade
upgrade:
	go run taskmaster-main.go upgrade --path data/localdev