# dump:
	#go build -o /dev/null -gcflags -S *

all:
	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-server-darwin-arm64 ./autobahn/server/autobahn-server.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-server-linux-amd64 ./autobahn/server/autobahn-server.go 

	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-client-darwin-arm64 ./autobahn/client/autobahn-client.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-client-linux-amd64 ./autobahn/client/autobahn-client.go 

key:
	openssl genrsa 2048 > privatekey.pem
	openssl req -new -key privatekey.pem -out csr.pem
	openssl x509 -req -days 36500 -in csr.pem -signkey privatekey.pem -out public.crt