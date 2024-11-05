CLI_NAME=poccdkgo

install-cli:
	@echo "Installing $(CLI_NAME)..."
	go build -o bin/$(CLI_NAME) ./cli
	sudo install -m 755 bin/$(CLI_NAME) /usr/local/bin/$(CLI_NAME)
	rm -rf bin/$(CLI_NAME)
	@echo "Done!"