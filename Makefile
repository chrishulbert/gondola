help:
	cat Makefile
	
run:
	go run *.go

build:
	go build

serve:
	python -m SimpleHTTPServer

try: build
