all:
	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-server-darwin-arm64 ./autobahn-server.go 
	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -tags=goexperiment.arenas -o autobahn-server-darwin-arm64-arena ./autobahn-server.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-server-linux-amd64 ./autobahn-server.go 

	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-client-darwin-arm64 ./autobahn-client.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-client-linux-amd64 ./autobahn-client.go 

key:
	openssl genrsa 2048 > privatekey.pem
	openssl req -new -key privatekey.pem -out csr.pem
	openssl x509 -req -days 36500 -in csr.pem -signkey privatekey.pem -out public.crt
