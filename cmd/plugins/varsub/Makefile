install:
	set -e
	go build -buildmode=plugin -o varsub.so plugin.go
	mv varsub.so /usr/local/bin/pia-plug