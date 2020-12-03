install:
	go get -v | cd web && yarn install
watch-go:
	air -c .air.toml
watch-web:
	cd web && yarn watch
build-docker:
	docker build -t gomodest .
run-docker:
	docker run -it --rm -p 4000:4000 gomodest:latest
mailhog:
	docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog