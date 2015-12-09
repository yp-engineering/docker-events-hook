run: echo docker-events-hook
	./docker-events-hook

echo: echo.go
	go build echo.go

docker-events-hook: docker-events-hook.go
	go build docker-events-hook.go
