all:
	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-server-testsuite-darwin-arm64 ./autobahn-server-testsuite.go 
	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -tags=goexperiment.arenas -o autobahn-server-testsuite-darwin-arm64-arena ./autobahn-server-testsuite.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-server-testsuite-linux-amd64 ./autobahn-server-testsuite.go 

	# mac, arm64
	GOOS=darwin GOARCH=arm64 go build -o autobahn-client-testsuite-darwin-arm64 ./autobahn-client-testsuite.go 
	# linux amd64
	GOOS=linux GOARCH=amd64 go build -o autobahn-client-testsuite-linux-amd64 ./autobahn-client-testsuite.go 

key:
	openssl genrsa 2048 > privatekey.pem
	openssl req -new -key privatekey.pem -out csr.pem
	openssl x509 -req -days 36500 -in csr.pem -signkey privatekey.pem -out public.crt
