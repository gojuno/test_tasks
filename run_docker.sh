#!/bin/bash
set -xeuo pipefail
echo 1 > /proc/sys/vm/drop_caches
docker-compose build
docker-compose up -d kv_server
docker-compose run --rm --entrypoint=/juno-test/bin/kv_server_bench kv_client -test.v -test.bench . -test.run ^$ -test.benchtime 1m
# docker-compose run --rm --entrypoint=/juno-test/bin/kv_server_bench_race kv_client -test.v -test.bench . -test.run ^$ -test.benchtime 1m
# docker-compose run linter
docker-compose push kv_server

