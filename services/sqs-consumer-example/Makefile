build:
	export GO111MODULE=on
	export CGO_ENABLED=1

	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/api/bootstrap api/handler/main.go
	chmod +x bin/api/bootstrap
	zip -j bin/api/handler.zip bin/api/bootstrap

	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/api-dlq/bootstrap api/handler-dlq/main.go
	chmod +x bin/api-dlq/bootstrap
	zip -j bin/api/handler-dlq.zip bin/api-dlq/bootstrap
	
clean:
	rm -rf ./bin

deploy: clean build
	go test -v ./...
	cdk bootstrap
	cdk deploy --all --ci --require-approval never --trace

seek: clean build
	go test -v ./...
	cdk synth --all
	cdk doctor
	cdk diff --all