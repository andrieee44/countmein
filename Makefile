run: clean migrate generate format vet
	go run .

build: clean migrate generate format vet
	nix build .#appimage

format:
	buf format -w
	go fmt

generate:
	tbls doc -f
	sqlc generate
	buf generate

migrate:
	@printf 'find ./store/sfuncs -name "*.sql" | %s | %s\n' \
		"xargs cat" \
		'mariadb "$(DB_NAME)"'
	@find ./store/sfuncs -name "*.sql" | xargs cat | \
		mariadb -h "$(DB_HOST)" -u "$(DB_USERNAME)" \
		"-p$(DB_PASSWORD)" --skip-ssl "$(DB_NAME)"
	@printf 'find ./store/sprocs -name "*.sql" | %s | %s\n' \
		"xargs cat" \
		'mariadb "$(DB_NAME)"'
	@find ./store/sprocs -name "*.sql" | xargs cat | \
		mariadb -h "$(DB_HOST)" -u "$(DB_USERNAME)" \
		"-p$(DB_PASSWORD)" --skip-ssl "$(DB_NAME)"
	atlas schema apply --env prod

deploy: build
	@echo "rsync -avz ./result $(DEPLOY_BIN_PATH)/countmein"
	@sshpass -p "$(SSH_PASSWORD)" \
		rsync -avz -e "ssh -p $(SSH_PORT)" --chmod=u+w --copy-links \
		--checksum ./result \
		"$(SSH_USER)@$(SSH_HOST):/$(DEPLOY_BIN_PATH)/countmein"
	@echo "rsync -avz --delete ./gen/docs/ $(DEPLOY_PUBLIC_PATH)"
	@sshpass -p "$(SSH_PASSWORD)" \
		rsync -avz --delete -e "ssh -p $(SSH_PORT)" --chmod=u+w  \
		--checksum ./gen/docs/ \
		"$(SSH_USER)@$(SSH_HOST):/$(DEPLOY_PUBLIC_PATH)"
	@echo 'ssh "systemctl --user restart countmein-api.service"'
	@sshpass -p "$(SSH_PASSWORD)" ssh -p $(SSH_PORT) \
		"$(SSH_USER)@$(SSH_HOST)" \
		"systemctl --user restart countmein-api.service"
	rm -f ./result

clean:
	rm -f ./result ./store/v2/*
	rm -rf ./gen
	@printf 'mariadb "$(DB_DEV_NAME)" -e %s\n' \
		'"DROP DATABASE $(DB_DEV_NAME); CREATE DATABASE $(DB_DEV_NAME);"'
	@mariadb -h "$(DB_DEV_HOST)" -u "$(DB_DEV_USERNAME)" \
		"-p$(DB_DEV_PASSWORD)" --skip-ssl "$(DB_DEV_NAME)" \
		-e "DROP DATABASE $(DB_DEV_NAME);" \
		-e "CREATE DATABASE $(DB_DEV_NAME);"

vet:
	sqlc vet
	buf lint
	tbls lint
	go vet

.PHONY: run build format generate migrate deploy clean vet
