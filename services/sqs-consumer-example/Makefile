build:
	go_build --path api/handler --zip handler
	go_build --path api/handler-dlq --zip handler-dlq

clean:
	rm -rf ./bin

test:
	go test -v ./...

deploy: test clean build
	cd stack; \
	cdk bootstrap; \
	cdk deploy --all --ci --require-approval never --trace;

seek: test clean build
	cd stack; \
	cdk synth --all; \
	cdk doctor;