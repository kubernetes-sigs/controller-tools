test-all:
	TRACE=1 ./test.sh

generate-modules:
	go mod tidy
