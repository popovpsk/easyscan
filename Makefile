up:
	docker-compose up -d --build

recreate-db:
	docker-compose exec database dropdb --if-exists -Upostgres easyscan
	docker-compose exec database createdb -Upostgres easyscan

test:
	go test -v -race -cover -count 1 -p 1 -cpu 1,4,16 ./...

bench:
	go test -bench=. -benchtime=5s -benchmem ./benchmarks
