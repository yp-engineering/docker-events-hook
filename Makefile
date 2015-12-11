BIN_ARGS?=

run: examples/echo docker-events-hook
	./docker-events-hook $(BIN_ARGS)

examples/echo: examples/echo.go
	cd examples && go build -i echo.go

docker-events-hook: docker-events-hook.go
	go build -i docker-events-hook.go
