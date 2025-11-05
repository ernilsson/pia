BIN=bin
INSTALL_DIR=/usr/local/bin

$(BIN):
	mkdir -p $@

$(BIN)/pia: $(BIN) $(wildcard **/*.go)
	go build -o $@ cmd/pia/main.go

.PHONY: install
install: $(BIN)/pia
	cp -rf $(BIN)/pia $(INSTALL_DIR)/pia

.PHONY: clean
clean: 
	rm -f $(BIN)/*
