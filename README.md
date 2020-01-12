## Install
On linux set your host to the following:

    127.0.0.1 localhost lin.ks # http only

typically `/etc/hosts`. After run `docker-compose up --build --remove-orphans`.

The following occurs:
   1) nginx-proxy is loaded and creates `VIRTUAL_HOST`s based on the environmental variables found and this occurs automatically 
   2) The linkshortener app loads with the`VIRTUAL_HOST` of `lin.ks`
   3) postgres is accessed and several tables are initialized

### Available Endpoints

The endpoints can be found [here](https://documenter.getpostman.com/view/9113626/SWLiZ66m?version=latest)

### Development Dependencies
```bash
# turn off modules
$ GO111MODULE=off
$ go get github.com/oxequa/realize
# turn on modules
$ GO111MODULE=on
```

### Development
```bash
$ ./scripts/dev-run.sh
```

### IMPORTANT
To be included: 
   1) tests
   2) monitor of link usage, 
