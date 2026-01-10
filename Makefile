build:
	go build

test:
	go test ./...

lint:
	bash maint/lint.sh

prettier-fix:
	prettier -w .
