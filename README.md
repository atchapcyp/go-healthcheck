## Building

Make sure `GOPATH` is defined and run

```shell
LOOM_SRC=$GOPATH/src/github.com/atchapcyp/healthcheck-cmd
# clone into gopath
git clone git@github.com:loomnetwork/loomchain.git $LOOM_SRC
# install deps
cd $LOOM_SRC
make deps
make
```

## Running

```shell
# init the blockchain with builtin contracts
./loom init
# run the node
./loom run
```
