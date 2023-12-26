install: 
	set -e
	go build -o pia cmd/cli/*.go
	mv pia /usr/local/bin/
	mkdir -p /etc/pia/plugins

varsub:
	set -e
	go build -buildmode=plugin -o varsub.so cmd/plugins/varsub/*.go
	mkdir -p /etc/pia/plugins
	mv varsub.so /etc/pia/plugins