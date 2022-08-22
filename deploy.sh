GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -v -ldflags="-s -w" -o ${program_name} .
rsync -uvcr ./ scott@10.0.0.101:~/controller/