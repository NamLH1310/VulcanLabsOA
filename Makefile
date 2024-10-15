fmt:
	go fmt ./... && sort-imports . && echo '\n'

run:
	CONFIG_PATH=config/config_local.yaml go run github.com/namlh/vulcanLabsOA/cmd

test:
	go test ./...