Endpoints

```bash
$ curl -X GET "http://localhost:8000/health" # returns server time and 200 responseCode

# login to receive JWT 
$ curl -X POST \                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  
  http://localhost:8000/login \       
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{"username":"admin", "password":"admin"}'

# refresh token
$ curl -X GET \                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  
    http://localhost:8000/auth/refresh_token \
    -H 'cache-control: no-cache' \
    -H 'content-type: application/json' \
    -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzg3NDU5MDUsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTU3ODc0MjMwNX0.CEX5LX_9ubcWYCme1qBMJDYIx4RArH7AHyRHCGhbMpg'

$ curl -X POST \
  http://localhost:8000/auth/v1/urls \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzg3NDYwNzAsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTU3ODc0MjQ3MH0.vO_f7oO-FYtst3m6Cv3JYLF31blIPIAKJwXy_ui6QB0' 
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
