VERSION=3.0.0
PREFIX=target/pagent-$(VERSION)

# go command for linux and windows.
GO=CGO_ENABLED=0 go
PARAMS=-ldflags '-s -w -extldflags "-static"'

# upx is a tool to compress executable program.
UPX=upx

PRGS=pagent pagentd


all:    $(PRGS)

pagent:
	$(GO) build $(PARAMS) -o $@ ./bin/pagent

pagentd:
	$(GO) build $(PARAMS) -o $@ ./bin/pagentd

clean:
	rm -f $(PRGS)

.PHONY: ./test
test:
	$(GO) test ./test

install:
	$(UPX) $(PRGS) || echo $?
	mkdir -p $(PREFIX)/etc
	cp -a etc/*.tpl $(PREFIX)/etc
	cp -a  Changelog.md $(PRGS) $(PREFIX)

	cd `dirname $(PREFIX)` && tar cvfz `basename $(PREFIX)`.tar.gz `basename $(PREFIX)`

