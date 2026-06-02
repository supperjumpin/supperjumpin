.PHONY: db-up db-migrate api-dev api-test

db-up:
	npm run db:up

db-migrate:
	npm run db:migrate

api-dev:
	npm run api:dev

api-test:
	npm run api:test
