# call-of-telegram
Send telegram message to wide range of users

## Install telegram lib
https://tdlib.github.io/td/build.html

## Build
```shell
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o ./bin/call-of-telegram ./cmd/call-of-telegram/main.go
```

## Environment setup
```shell
API_ID=<your_api_id>
API_HASH=<your_api_hash>
```

## Run
```shell
./call-of-telegram broadcast --group-name "Example Group name" --msg-file "/path/to/file/with/message"
```