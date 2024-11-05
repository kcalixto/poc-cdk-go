define build_go_binary
	@if [ -z $(1) ]; then \
		echo "executable path is required"; \
		exit 1; \
	fi
	@if [ -z $(2) ]; then \
		echo "zip name is required"; \
		exit 1; \
	fi
	BOOTSTRAP_PATH="bin/$(1)/bootstrap"; \
	export GO111MODULE=on; \
	export CGO_ENABLED=1; \
	export GOARCH=arm64; \
	if [ `uname` = "Darwin" ]; then \
		export GOOS=darwin; \
	else \
		export GOOS=linux; \
	fi; \
	go build -ldflags="-s -w" -o $$BOOTSTRAP_PATH $(1)/main.go; \
	chmod +x $$BOOTSTRAP_PATH; \
	zip -j bin/$(2).zip $$BOOTSTRAP_PATH
endef