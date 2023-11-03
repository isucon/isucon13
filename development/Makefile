
up/%:
	docker compose -f docker-compose-common.yml -f docker-compose-$*.yml up $(OPT)

down/%:
	docker compose -f docker-compose-common.yml -f docker-compose-$*.yml down --volumes
