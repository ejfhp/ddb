GOOS=windows GOARCH=amd64 go build -o trh.exe .
GOOS=linux GOARCH=amd64 go build -o trh-linux .
GOOS=darwin GOARCH=arm64 go build -o trh-macm1 .
GOOS=darwin GOARCH=amd64 go build -o trh-macin .
