# urlshortener
URL shortener service implemented in Go. 

It allows you to shorten long URLs into shorter, more manageable ones and then redirect users to the original URLs.

It uses Redis as a data storage.

# Features
- Shorten long URLs.
- Redirect to the original URLs.
- Delete short urls.

# Pre-requisite
- docker
- docker-compose

# Management
1. `make build` - to build service
2. `make up` - to run service
3. `make down` - teardown
4. `make test` - run tests

# Usage
Server runs on port 8090 by default.
1. `curl -X POST http://localhost:8090/ -i -H "Content-Type: application/json" -d '{"url": "https://google.com"}'` - will generate shorten url for https://google.com
2. `curl http://localhost:8090/FtKtW5L9VF` - assuming `http://localhost:8090/FtKtW5L9VF` is short url generated in step one, redirect to original url should happen.
3. `curl -X DELETE http://localhost:8090/FtKtW5L9VF` - delete short url.
