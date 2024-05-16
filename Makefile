VERSION=3.0.0
PREFIX=target/pagent-$(VERSION)

# go command for linux and windows.
GO=CGO_ENABLED=0 go
PARAMS=-ldflags '-s -w -extldflags "-static"'

# if want to link static libpcap.a then 
# go build -ldflags '-s -w -extldflags "-L/usr/local/libpcap-1.10.4/lib -lpcap"' ./bin/pagent
# remember setcap 'CAP_NET_RAW,CAP_NET_ADMIN,CAP_DAC_OVERRIDE+ep' pagent

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
	# setcap 'CAP_NET_RAW,CAP_NET_ADMIN,CAP_DAC_OVERRIDE+ep' pagent
	# setcap 'CAP_NET_RAW,CAP_NET_ADMIN,CAP_DAC_OVERRIDE+ep' pagentd
	$(UPX) $(PRGS) || echo $?
	mkdir -p $(PREFIX)/etc
	cp -a etc/*.tpl $(PREFIX)/etc
	cp -a  Changelog.md $(PRGS) $(PREFIX)

	cd `dirname $(PREFIX)` && tar cvfz `basename $(PREFIX)`.tar.gz `basename $(PREFIX)`

