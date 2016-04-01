all:
	go build -gcflags "-N -l"

docker:
	GOOS=linux go build ./ 
	docker build -t firmware ./

run:
	docker run -p 8081:8080 -it -d -v $(PWD)/package.json:/package.json -v $(PWD)/update_v0.9.4.zip:/update_v0.9.4.zip firmware
