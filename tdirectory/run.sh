#Assuming docker is installed in the system

docker rm -f tdirectory mongodb

docker pull mongo:latest

docker run -it -d  --name mongodb mongo:latest

make clean; make

docker run -it -d -p 8080:8080 --name tdirectory tdirectory:v1

