run: vet generate migrate
	go run .

build: vet generate migrate
	nix build

generate:
	tbls doc --rm-dist
	sqlc generate
	buf generate
	buf format -w

migrate:
	atlas schema apply --env prod

deploy: build
	@echo "rsync -avz result/bin/countmein $(DEPLOY_BIN_PATH)"
	@sshpass -p "$(SSH_PASSWORD)" \
		rsync -avz -e "ssh -p $(SSH_PORT)" --chmod=u+w \
		result/bin/countmein \
		"$(SSH_USER)@$(SSH_HOST):/$(DEPLOY_BIN_PATH)"
	@echo "rsync -avz --delete ./gen/docs $(DEPLOY_PUBLIC_PATH)"
	@sshpass -p "$(SSH_PASSWORD)" \
		rsync -avz --delete -e "ssh -p $(SSH_PORT)" --chmod=u+w  \
		./gen/docs/ \
		"$(SSH_USER)@$(SSH_HOST):/$(DEPLOY_PUBLIC_PATH)"
	@echo 'ssh "systemctl --user restart countmein-api.service"'
	@sshpass -p "$(SSH_PASSWORD)" ssh -p $(SSH_PORT) \
		"$(SSH_USER)@$(SSH_HOST)" \
		"systemctl --user restart countmein-api.service"
	rm -f result

clean:
	rm -f result
	@printf 'mariadb "$(DB_DEV_NAME)" -e %s\n' \
		'"DROP DATABASE $(DB_DEV_NAME); CREATE DATABASE $(DB_DEV_NAME);"'
	@mariadb -h "$(DB_DEV_HOST)" -u "$(DB_DEV_USERNAME)" \
		"-p$(DB_DEV_PASSWORD)" \ --skip-ssl "$(DB_DEV_NAME)" \
		-e "DROP DATABASE $(DB_DEV_NAME);" \
		-e "CREATE DATABASE $(DB_DEV_NAME);"

vet:
	sqlc vet
	buf lint
	tbls lint
	go vet

.PHONY: run build generate migrate deploy clean vet
