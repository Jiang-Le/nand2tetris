GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)
GO      = ${GOROOT}/bin/go
MakeFileDir = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

JACK_FILES := $(wildcard $(MakeFileDir)../ArrayTest/*.jack $(MakeFileDir)../ExpressionLessSquare/*.jack $(MakeFileDir)../Square/*.jack)
TOKEN_XML_FILES := $(JACK_FILES:.jack=T.xml)
GENERATED_TOKEN_XML_FILES := $(addsuffix .generated,$(TOKEN_XML_FILES))

LEX_XML_FILES := $(JACK_FILES:.jack=.xml)
GENERATED_LEX_XML_FILES := $(addsuffix .generated,$(LEX_XML_FILES))


all: compare_token compare_lex
	
%T.xml.generated: %.jack
	$(GO) test -run TestTokenizer -v -args $< $@

%.xml.generated: %.jack
	$(GO) test -run CompilationEngine -v -args -if $< -of $@
	
compare_token: $(TOKEN_XML_FILES) $(GENERATED_TOKEN_XML_FILES)
	@for xml in $(TOKEN_XML_FILES); do \
		if diff -w $$xml $${xml}.generated >/dev/null; then \
        	echo $$xml "are identical"; \
		else \
			echo $$xml "are different"; \
		fi \
    done

compare_lex: $(LEX_XML_FILES) $(GENERATED_LEX_XML_FILES)
	@for xml in $(LEX_XML_FILES); do \
		if diff -w $$xml $${xml}.generated >/dev/null; then \
        	echo $$xml "are identical"; \
		else \
			echo $$xml "are different"; \
		fi \
    done

clean:
	rm -rf $(GENERATED_TOKEN_XML_FILES) $(GENERATED_LEX_XML_FILES)
