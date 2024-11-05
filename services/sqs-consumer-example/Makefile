include ../../scripts/go.mk

build:
	$(call build_go,api/handler,handler.zip)
	$(call build_go,api/handler-dlq,handler-dlq.zip)

clean:
	rm -rf ./bin

test:
	go test -v ./...

deploy: clean build test
	cd stack; \
	cdk bootstrap; \
	cdk deploy --all --ci --require-approval never --trace;

seek: clean build test
	cd stack; \
	cdk synth --all; \
	cdk doctor;

diff:
	cd stack; \
	cdk diff --all;