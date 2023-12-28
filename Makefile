install: 
	set -e
	go build -o pia cmd/cli/*.go
	mv pia /usr/local/bin/
	mkdir -p /etc/pia/plugins

varsub:
	set -e
	go build -buildmode=plugin -o varsub.so cmd/plugins/varsub/*.go
	mv varsub.so /etc/pia/plugins

jshooks:
	set -e
	go build -buildmode=plugin -o jshooks.so cmd/plugins/jshooks/*.go
	mv jshooks.so /etc/pia/plugins

all: install varsub jshooks

purge-plugins:
	set -e
	rm /etc/pia/plugins/*.so