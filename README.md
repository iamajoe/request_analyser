# Request analyser

> A tool that takes a set of http requests from a source, runs it against an API and analyses the best it can.

---

## Usage

```bash
make build
```

### Stats

Retrieve a count statistic of the requests

```bash
# Running with a json file source
./bin/request_analyser stats -s "<file_path>"

# Running with a raw file source
./bin/request_analyser stats -s "<file_path>"

# Running with a redis source
./bin/request_analyser stats -s "<redis|rediss>://<redis_connect_url>;<pattern>"
```

## Source examples

### JSON

`requestUrl` is required, all the other properties will be defaulted. `requestMethod` defaults to `GET`

```json
[
  {
    "requestUrl": "/status"
  },
  {
    "requestUrl": "/users/list",
    "requestMethod": "GET"
  },
  {
    "requestUrl": "/users/login",
    "requestMethod": "POST",
    "requestBody": {
      "username": "amazing@email.com",
      "password": "a_very_strnong_password"
    }
  },
  {
    "requestUrl": "/notifications/count",
    "requestMethod": "POST",
    "requestHeaders": {
      "Content-Type": "application/json",
      "Authorization": "Bearer super_token"
    },
    "requestBody": {
      "type": "feed"
    }
  }
]
```

```bash
# source saved as requests.json on the current directory
./bin/request_analyser stas -s "requests.json"

# relative path
./bin/request_analyser stats -s "./requests.json"

# absolute path
./bin/request_analyser stats -s "/var/log/requests.json"
```

### Raw

Separate each requests with a `\n` and each property is separated by a `;`. Values under properties will be separated by the first `:`.
`requestUrl` is required, all the other properties will be defaulted. `requestMethod` defaults to `GET`.
You can setup comments using `#` on the first character.

```
# request status defaulting to "GET" method
requestUrl:/status

# same status request but with the method in there
requestUrl:/users/list;requestMethod:GET

# a post request with a body
requestUrl:/users/login;requestMethod:POST;requestBody:{"username":"amazing@email.com","password":"a_very_strong_password"}

# a post request with body and headers
requestUrl:/notifications/count;requestMethod:POST;requestHeaders:{"Content-Type":"application/json","Authorization":"Bearer super_token"};requestBody:{"type": "feed"}
```

```bash
# source saved as requests.txt on the current directory
./bin/request_analyser stats -s "requests.txt"

# relative path
./bin/request_analyser stats -s "./requests.txt"

# absolute path
./bin/request_analyser stats -s "/var/log/requests.txt"
```

### Redis

```bash
# example returning all keys
./bin/request_analyser stats -s "redis://url_for_the_redis;*"

# the pattern defaults to the all keys wildcard
./bin/request_analyser stats -s "redis://url_for_the_redis"

# retrieve a specific key pattern
./bin/request_analyser stats -s "redis://url_for_the_redis;req_record_*"

# example using secure protocols
./bin/request_analyser stats -s "rediss://url_for_the_redis"
```
