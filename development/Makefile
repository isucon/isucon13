
up/%:
	docker compose -f docker-compose-common.yml -f docker-compose-$*.yml up -d --build

down/%:
	docker compose -f docker-compose-common.yml -f docker-compose-$*.yml down --volumes

bench:
	make linux_amd64
	./bin/bench_linux_amd64 run --dns-port=1053

