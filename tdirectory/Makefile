all: build docker

build:
	GOOS=linux go build

docker: build
	docker build -t tdirectory:v1 .
clean:
	go clean
	rm -rf tdirectory
	make -C docker clean
    
