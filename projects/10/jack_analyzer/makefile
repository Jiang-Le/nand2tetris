GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)
GO      = ${GOROOT}/bin/go
MakeFileDir = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

JACK_FILES := $(wildcard $(MakeFileDir)../ArrayTest/*.jack $(MakeFileDir)../ExpressionLessSquare/*.jack $(MakeFileDir)../Square/*.jack)
TOKEN_XML_FILES := $(JACK_FILES:.jack=T.xml)
GENERATED_TOKEN_XML_FILES := $(addsuffix .generated,$(TOKEN_XML_FILES))

LEX_XML_FILES := $(JACK_FILES:.jack=.xml)
GENERATED_LEX_XML_FILES := $(addsuffix .generated,$(LEX_XML_FILES))

COVERAGE_FILES := $(JACK_FILES:.jack=.coverage.out)

# 使用shell函数检查/dev/null是否存在
HAS_DEVNULL := $(shell if [ -e /dev/null ]; then echo 1; else echo 0; fi)

# 根据HAS_DEVNULL的值定义NULL_DEVICE变量
ifeq ($(HAS_DEVNULL),1)
NULL_DEVICE := /dev/null
else
NULL_DEVICE := NUL
endif

NULL_DEVICE := NUL

.PHONY: all compare_token compare_lex clean coverage

all: compare_token compare_lex

coverage: coverage.html
	
%T.xml.generated: %.jack
	$(GO) test -run TestTokenizer -v -args $< $@

%.xml.generated: %.jack
	$(GO) test -run CompilationEngine -coverprofile=coverage.out -v -args -if $< -of $@

%.coverage.out: %.jack
	$(GO) test -run CompilationEngine --coverprofile=$@ -v -args -if $< -of $(NULL_DEVICE)
	
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

merged_coverage.out: $(COVERAGE_FILES)
	echo "mode: atomic" > merged_coverage.out;
	@for file in $(COVERAGE_FILES); do \
		grep -v "mode: " $$file >> merged_coverage.out; \
	done

coverage.html: merged_coverage.out
	$(GO) tool cover -html=merged_coverage.out -o coverage.html
	
clean:
	rm -rf $(GENERATED_TOKEN_XML_FILES) $(GENERATED_LEX_XML_FILES) $(COVERAGE_FILES) merged_coverage.out coverage.html
