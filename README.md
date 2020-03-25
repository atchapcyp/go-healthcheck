## Environment Setup

Requirements

- Go 1.12+

## Building

Make sure `GOPATH` is defined and run

```shell
REPO_SRC=$GOPATH/src/github.com/atchapcyp/go-healthcheck
# clone into gopath
git clone git@github.com:atchapcyp/go-healthcheck.git $REPO_SRC
# install dependencies
cd $REPO_SRC
make build
```

## Running

```shell
# run with specific file and set request sender amount and request timeout limit.
./health -f weblist.csv -n 100 -t 30
```

or

```
go run `ls cmd/*.go | grep -v _test.go` -f weblist.csv -n 100
```
