# Workshop

## Task 1

The first challenge is to setup the environment by running 

```bash
docker compose up --detach
```
and answering the following questions

1. For the running envoy instance, what is the port of the control-plane the envoy instance is
connected to?
2. Did envoy connect to the control-plane before it was ready?

You may find the envoy admin panel helpful in answering these questions.
- **Envoy Admin Panel**: http://localhost:60001


## Task 2


Run the following script

```
scripts/make_request.sh 10000
```

It should fail because at this point you have no resources.

To configure envoy via the control-plane run to generate a cli to propagate envoy resources

```
go build -o bin/cli ./cli
```

the run

```
./bin/cli --action update -config workshop-resources/basic-grpc.json
```

to add some configuration to envoy.

Observe that there are resources in your cache at `http://localhost:8734/`

Run

```
scripts/make_request.sh 10000
```

You should get

```
> scripts/make_request.sh 10000

Sending request via grpcurl to Envoy (localhost:10000)...
{
  "message": "Hello Test User from Docker-gRPC server"
}
grpcurl request successful!
```

## Task 3

Run

```
./bin/cli --action update -config workshop-resources/basic-http.json
```

and then run

```
scripts/make_http_request.sh 10001
```

You should get a 503 error

```
> scripts/make_http_request.sh 10001
Sending HTTP request via curl to Envoy HTTP/1.1 listener (localhost:10001)...
Note: Unnecessary use of -X or --request, POST is already inferred.
* Host localhost:10001 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:10001...
* Connected to localhost (::1) port 10001
> POST /test/sayhello HTTP/1.1
> Host: localhost:10001
> User-Agent: curl/8.7.1
> Accept: */*
> Content-Type: application/json
> Content-Length: 27
>
* upload completely sent off: 27 bytes
< HTTP/1.1 503 Service Unavailable
< content-length: 167
< content-type: text/plain
< date: Tue, 19 Aug 2025 20:55:18 GMT
< server: envoy
<
* The requested URL returned error: 503
* Closing connection
curl: (22) The requested URL returned error: 503
```

The task is modify basic-http.json so that works

```bash
scripts/make_http_request.sh 10001
```

You may find it helpful to exec into the debug container
or read the docker-compose.yml file to understand how the servers are setup
```
docker compose exec debug bash
```

# General Tips and Tricks
## Stop Services

```bash
docker compose down
```

## Development Tips

```bash
# Rebuild images when code changes
docker compose up --build

# View logs
docker compose logs -f

# Rebuild specific service
docker compose build control-plane
```

## Debugging

Access the debug container with network troubleshooting tools:

```bash
# Shell into debug container
docker compose exec debug bash

# Inside the debug container, you can use:
# Test connectivity
ping control-plane
ping 172.20.0.10

# Test HTTP endpoints
curl http://control-plane:8734
curl http://envoy:60001

# Test gRPC services using server reflection
grpcurl -plaintext -d '{"name": "Debug User"}' test-server-grpc:50051 test.TestService/SayHello

# Test via Envoy proxy (requires xDS configuration be applied)
grpcurl -plaintext -d '{"name": "Via Envoy"}' envoy:10000 test.TestService/SayHello
```
