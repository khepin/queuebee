-include Makefile_local

default:
	go run *.go || (go get && go run *.go)

watch:
	(watchexec -e sql -- make sqlc) & \
	(watchexec -w db/sqlc.yaml -- make sqlc) & \
	(watchexec -e go -- go get ./...) & \
	(watchexec -e go -c -r -- go run *.go) & \
	wait

sqlc:
	sqlc -f db/sqlc.yaml generate

.PHONY: migrate
migrate:
	atlas schema apply -u "sqlite://data/queuebee.db" --to="file://db/schema.sql" --dev-url="sqlite://db/test.db"
	rm db/test.db
	echo "PRAGMA journal_mode=wal;" | sqlite3 data/queuebee.db
