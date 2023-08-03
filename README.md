# Request analyser

> A tool that takes a set of http requests from a source, runs it against an API and analyses the best it can.

---

## Install / build

```bash
make build
```

## Parse

Converts and validates a source to be used with the analyser

```bash
# Running with a json file source
./bin/request_analyser parse -s "<file_path>" -o "<output_file_path>"

# Running with a raw file source
./bin/request_analyser parse -s "<file_path>" -o "<output_file_path>"

# Running with a redis source
./bin/request_analyser parse -s "<redis|rediss>://<redis_connect_url>;<pattern>" -o "<output_file_path>"
```

### Source examples

#### JSON

`requestUrl` is required, all the other properties will be defaulted. `requestMethod` defaults to `GET`

```json
[
  {
    "unix": 1690979263,
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
./bin/request_analyser parse -s "requests.json" -o "records_output"

# relative path
./bin/request_analyser parse -s "./requests.json" -o "./records_output"

# absolute path
./bin/request_analyser parse -s "/var/log/requests.json" -o "/var/log/records_output"
```

#### Raw

Separate each requests with a `\n` and each property is separated by `;;` (so it doesn't colide with the single `;` of for example the headers). Values under properties will be separated by the first `:`.
`requestUrl` is required, all the other properties will be defaulted. `requestMethod` defaults to `GET`.
You can setup comments using `#` on the first character.

```
# request status defaulting to "GET" method
requestUrl:/status;;unix:1690979263

# same status request but with the method in there
requestUrl:/users/list;;requestMethod:GET

# a post request with a body
requestUrl:/users/login;;requestMethod:POST;;requestBody:{"username":"amazing@email.com","password":"a_very_strong_password"}

# a post request with body and headers
requestUrl:/notifications/count;requestMethod:POST;;requestHeaders:{"Content-Type":"application/json","Authorization":"Bearer super_token"};requestBody:{"type": "feed"}
```

```bash
# source saved as requests.json on the current directory
./bin/request_analyser parse -s "requests.txt" -o "records_output"

# relative path
./bin/request_analyser parse -s "./requests.txt" -o "./records_output"

# absolute path
./bin/request_analyser parse -s "/var/log/requests.txt" -o "/var/log/records_output"
```

#### Redis

```bash
# example returning all keys
./bin/request_analyser parse -s "redis://url_for_the_redis;*" -o "records_output"

# the pattern defaults to the all keys wildcard
./bin/request_analyser parse -s "redis://url_for_the_redis" -o "records_output"

# retrieve a specific key pattern
./bin/request_analyser parse -s "redis://url_for_the_redis;req_record_*" -o "records_output"

# example using secure protocols
./bin/request_analyser parse -s "rediss://url_for_the_redis" -o "records_output"
```

## Stats

Retrieve a count statistic of the requests

```bash
./bin/request_analyser stats -i "<file_path>"
```

## Run requests

Runs the requests from the parsed file

```bash
# run with 1 concurrent job with a base url
./bin/request_analyser run -i "<file_path>" -b "http://localhost:4040"

# run with 100 concurrent jobs without base url (if http|https provided)
./bin/request_analyser run -i "<file_path>" -c 100

# set a new request every 10 ms (does not work with -s)
./bin/request_analyser run -i "<file_path>" -t 10

# if "unix" is provided and valid unix, speeds up by 10 (does not work with -t)
./bin/request_analyser run -i "<file_path>" -s 10

# filters a pattern of endpoints / method
# wildcards acepted on endpoint and method, endpoints are regex based
./bin/request_analyser run -i "<file_path>" -f "['POST:*', *:users\/create]"
```
