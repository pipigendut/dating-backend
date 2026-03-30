.PHONY: dev prod build

dev:
	APP_ENV=development go run cmd/app/main.go

prod:
	APP_ENV=production go run cmd/app/main.go

build:
	go build -o bin/app cmd/app/main.go
