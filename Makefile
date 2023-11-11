install: 
	set -e
	go build -o pia cmd/cli/*.go
	mv pia /usr/local/bin/