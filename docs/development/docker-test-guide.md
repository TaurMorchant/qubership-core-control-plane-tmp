# Docker Test Guide

* [Running tests](#running-tests)
* [Writing tests](#writing-tests)

## Running tests
To run tests with docker you need to set environment variable `TEST_DOCKER_URL` with proper Docker API address. 

If using Docker Desktop on Windows: 
1. Go to **Settings** - **General** and enable **Expose daemon on tcp://localhost:2375 without TLS**. 
2. Set up environment variable `TEST_DOCKER_URL=tcp://localhost:2375`. 

When `TEST_DOCKER_URL` is set, tests will be executed during maven build:

`mvn clean install`

## Writing tests
1. All tests with docker located in `main` package of `contron-plane-service` module. 
2. Source files related to tests named starting with `it_` prefix to indicate that these are integration tests
and ending with `_test.go` suffix to avoid including in delivery.
3. Each **test must start with line** `skipTestIfDockerDisabled(t)`. 
4. In the end of each test you **need to restore the initial state of control-plane**: remove all test routes and all versions except v1. 
5. **Test microservice** to verify traffic flow through gateway can be started with code:
    ```go
    traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", true)
    defer traceSrvContainer.Purge()
    ``` 
    You can run several test microservices with different names and/or deployment versions. 
6. Global variables in `it_engine_test.go` store many useful thing such as Docker pool, gateway addresses, idpStub and so on. 
7. **Custom docker containers** should be run using `containerManager` global variable. 
8. `it_utils_test.go` contains useful functions such as registration and deletion of routes, sending requests to test service through gateway and etc. 
9. `it_idp_stub.go` contains **identity-provider stub** implementation: it can generate m2m and cloud-admin tokens that can be used in tests: control-plane will successfully parse them. 
    See public functions in `it_idpstub_test.go`. 
10. **Test microservice implementation** located in `control-plane-test-service` module and can be changed for any testing purposes. 
11. In case Docker resources haven't been purged after running tests you can use script `./control-plane-test-service/cleanup_dockertest.sh`.
    > :warning: This script will delete **ALL** docker containers on your machine, not just created by tests. 
