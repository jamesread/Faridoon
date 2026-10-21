# Regenerates marketing screenshots against a throwaway stack seeded with
# fictional data. Run `make` here; everything is torn down afterwards.
#
# The Go service serves the built SPA itself (FARIDOON_STATIC_DIR), so the
# browser and API share one origin and no Vite dev server is needed.

ROOT := $(abspath ../../..)
HERE := $(abspath .)
TMP := $(HERE)/tmp

CONTAINER ?= podman
MYSQL_CTR := faridoon-mkt-mysql
MYSQL_PORT := 13306

BACKEND_URL := http://127.0.0.1:18480

DOCS_ASSETS := $(ROOT)/docs/assets/marketing
DOCS_IMAGES := quotes.png

.PHONY: all build up seed shoot manifest publish down clean

all:
	@$(MAKE) build up seed shoot manifest publish; rc=$$?; $(MAKE) down; exit $$rc

build:
	$(MAKE) -C $(ROOT) build

up:
	mkdir -p $(TMP)/service
	cp service.yaml $(TMP)/service/config.yaml
	$(CONTAINER) run -d --rm --name $(MYSQL_CTR) -p 127.0.0.1:$(MYSQL_PORT):3306 \
		-e MYSQL_ROOT_PASSWORD=root-demo -e MYSQL_DATABASE=faridoon \
		-e MYSQL_USER=faridoon -e MYSQL_PASSWORD=faridoon-demo \
		docker.io/library/mysql:8.4
	@echo "Waiting for MySQL..."
	@for i in $$(seq 120); do \
		$(CONTAINER) exec $(MYSQL_CTR) mysql -ufaridoon -pfaridoon-demo -h127.0.0.1 -e 'SELECT 1' faridoon >/dev/null 2>&1 && exit 0; \
		sleep 1; done; echo "MySQL did not start"; exit 1
	cd $(ROOT)/database/mysql && DB_HOST=127.0.0.1:$(MYSQL_PORT) DB_NAME=faridoon \
		DB_USER=faridoon DB_PASS=faridoon-demo sql-migrate up
	cd $(TMP)/service && { env -u DB_DATABASE -u DB_USERNAME -u DB_PASSWORD -u PORT -u SITE_TITLE \
		DB_HOST=127.0.0.1 DB_NAME=faridoon DB_USER=faridoon DB_PASS=faridoon-demo \
		FARIDOON_STATIC_DIR=$(ROOT)/frontend/dist \
		$(ROOT)/service/faridoon-service -configdir . > $(TMP)/service.log 2>&1 & echo $$! > $(TMP)/service.pid; }
	# Init touches the database (cvars), so it only succeeds once the DB is usable.
	@for i in $$(seq 60); do curl -fsS -X POST -H 'Content-Type: application/json' -d '{}' \
		$(BACKEND_URL)/faridoon.v1.FaridoonService/Init >/dev/null 2>&1 && exit 0; sleep 1; done; \
		echo "Backend did not become healthy; see $(TMP)/service.log"; exit 1

seed:
	./seed.sh $(BACKEND_URL) $(TMP)
	$(CONTAINER) exec -i $(MYSQL_CTR) mysql -ufaridoon -pfaridoon-demo faridoon < seed/quote-history.sql

shoot:
	repo-helper screenshot --config screenshots.ini

manifest:
	./generate-manifest.py

publish:
	mkdir -p $(DOCS_ASSETS)
	cp $(DOCS_IMAGES) $(DOCS_ASSETS)/

down:
	-@for p in service; do \
		[ -f $(TMP)/$$p.pid ] && kill $$(cat $(TMP)/$$p.pid) 2>/dev/null; \
		rm -f $(TMP)/$$p.pid; done
	-@$(CONTAINER) rm -f $(MYSQL_CTR) >/dev/null 2>&1

clean: down
	rm -rf $(TMP)
