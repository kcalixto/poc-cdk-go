define build_go
	env GOARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o bin/$1/bootstrap $1/main.go
	chmod +x bin/$1/bootstrap
	zip -j bin/$2 bin/$1/bootstrap
endef