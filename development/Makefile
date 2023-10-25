.PHONY: install-initialdata init-mysql run-bench run-bench-noload clean

up/%:
	docker compose -f docker-compose-$*.yml -f docker-compose-common.yml up $(OPT)

down/%:
	docker compose -f docker-compose-$*.yml -f docker-compose-common.yml down

logs-webapp/%:
	docker compose -f docker-compose-$*.yml -f docker-compose-common.yml logs -f webapp

init-mysql:
	docker compose -f docker-compose-common.yml exec --workdir /docker-entrypoint-initdb.d mysql \
		bash -c 'cat 10_schema.sql 90_data.sql | mysql -uisucon -pisucon isuports'

run-bench:
	docker compose -f docker-compose-common.yml exec bench ./run_local.sh -target-addr nginx:443

run-bench-noload:
	docker compose -f docker-compose-common.yml exec bench ./run_local.sh -prepare-only -strict-prepare=false -target-addr nginx:443

run-bench-ci:
	timeout 60 ./wait-for mysql
	make init-mysql
	timeout 60 ./wait-for app
	docker compose -f docker-compose-common.yml exec bench make ci-with-comment

../initial_data.tar.gz:
	gh release list | awk '/Latest/{print $$3}' | xargs gh release download --dir ..

install_initial_data:
	docker compose -f docker-compose-common.yml  cp ../initial_data/ bench:/home/isucon/

initial_data: ../initial_data.tar.gz
	tar -C .. -xvf ../initial_data.tar.gz

clean:
	rm -rf ../initial_data*
	git checkout ../initial_data/
