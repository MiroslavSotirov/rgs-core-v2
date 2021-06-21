RGS Core V2


Requirements:
```cassandraql
go v 1.13.1 (gomodule should be set to on, auto should handle this in most cases, you can also export GO111MODULE=on to be certain)
memcached server -- see setup below

```
This directory should live inside your GOROOT/src directory (by default GOROOT is $HOME/go)
Dependency: memcached
Set up local memcached server with:
```
memcached -l 127.0.0.1 -m 64 -vv
or
docker run --name memcached1 -p 11211:11211 -d memcached
```
Find address of memcached container and set env variable for rgs to lookup
```
docker inspect memcached1
export MCROUTER=127.0.0.1:11211 (or whichever address your memcached container has)

todo: add in config file to set address
```

To run locally:
```cassandraql
make build
make run
```

Build and run Docker rgs container (this needs updating after restructuring)
```
export GO111MODULE=on (
cd go/src/rgs-core-v2
docker build -t rgs-core-v2.latest .
docker run -p 3000:3000 rgs-core-v2.latest

```


To run tests: 
`make test`

To run on only one module:
``` go test -v rgs-core-v2/internal/engine```
(remove -v flag to remove verbosity)
`log.Printf` functions can be useful for debugging but will not be displayed on testing unless the -v flag is present

To run the server: `make run`

Due to Chrome security and CORS flights, OPTIONS calls are made to the RGS. In production this will be handled by the load balancer so to avoid creating unwanted endpoints on the RGS, run a a security-free chrome window
```cassandraql
open -n -a /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --args --user-data-dir="/tmp/chrome_dev_test" --disable-web-security
```


You may need to update the certificate/key for your localhost, add new files to config or change the paths in rgs.go lines 132/133.

From your browser, try:
```localhost:3000/v2/rgs/play/test-engine
localhost:3000/v2/rgs/spin (needs to be updated for new Spin method)
```
Running against a local client
==============================
To run a version of the game client locally and point it to the correct version of the rgs, clone the following repo:
```https://gitlab.maverick-ops.com/cobalt7/the-year-of-zhu```
- If the version of mvwrapper in package.json is lower than 2.1.37, update this before building the client locally.
- Follow the README to build and run locally a version of the client
- Open a security-free chrome window (see above) and try to access the following urls:
	- http://localhost:3000/v2/rgs/initall
	- http://127.0.0.1:8080/index.html?operator=local2&gameName=the-year-of-zhu&language=en&mode=realplay&token=testToken&currency=USD&wallet=demo
- NB: you may change the token to anything, if it doesn't exist already, a new entry will be added into memcached


Running against a docker image of the client
============================================
(can only be done once client version with proper mvwrapper is pushed)
- pull docker image from harbor
- run docker image

Sending images to harbor
=======================
```	
docker build -t 'harbor.maverick-ops.com/maverick/mvg_rgs:v2.1' rgs-core-v2
docker push harbor.maverick-ops.com/maverick/mvg_rgs:v2.1
```


Protobuf
========

Updates to the datatypes in the protobuf file should only be additive.
To recompile protobuf types, run the following from `/internal/engine`: 

`protoc --go_out=. *.proto`



Editing the Makefile
```sh
jetbrains IDEs by default turn a tab into 4 spaces. Makefiles are sensitive to spaces vs tabs *(must use tabs!) so set your IDE to use tabs in Other file types under Preferences | Editor | Code Style | Other file types
```

Make commands
=============
```sh
make start - build and run rgs-core-v2 locally
make test - runs unit test

make build - build docker image and tag with current version
make push - push current docker image to harbor registry
make latest - tag a current docker image as latest and push to harbor registry (warning: this might override RGSv1 if tagged with `:latest` as they share the same repository)

make run - create and run a container from current docker image (requires MCROUTER environment variable)
make stop - stop and remove currently running rgs container 
```

Configuration
============
Create/Update `config.yml` in the configs/ directory
```sh
# sample config

# development mode
devmode: true
local: true

# Log levels
# debug, info, warn, error, fatal
logging: debug

# mcrouter address
mcrouter: "0.0.0.0:11211"

# server configuration
server:
  host: "0.0.0.0"
  port: 3000
```