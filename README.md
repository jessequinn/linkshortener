Endpoints

```bash
$ curl -X GET "http://localhost:8000/health" # returns server time and 200 responseCode

$ curl -X POST \
  http://localhost:8000/v1/urls \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{"url":"http://jessequinn.info"}'

```




Development Dependencies

```bash
# turn off modules
$ GO111MODULE=off
$ go get github.com/oxequa/realize
# turn on modules
$ GO111MODULE=on
```


Development

```bash
$ ./scripts/dev-run.sh
```
