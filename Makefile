run: vet generate migrate
	nix run

build: vet generate migrate
	nix build

generate:
	tbls doc --rm-dist
	sqlc generate
	buf generate
	buf format -w

migrate:
	atlas schema apply --env prod

clean:
	@printf "mariadb $(DB_DEV_NAME) -e %s\n" \
		'"DROP DATABASE $(DB_DEV_NAME); CREATE DATABASE $(DB_DEV_NAME);"'
	@mariadb -h $(DB_DEV_HOST) -u $(DB_DEV_USERNAME) -p$(DB_DEV_PASSWORD) \
		--skip-ssl $(DB_DEV_NAME) -e "DROP DATABASE $(DB_DEV_NAME);" \
		-e "CREATE DATABASE $(DB_DEV_NAME);"

vet:
	go vet
	sqlc vet
	buf lint
	tbls lint

.PHONY: run build generate migrate clean vet
