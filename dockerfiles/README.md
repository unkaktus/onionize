Building
-------
You have to have docker 17.05+ installed.

To build `onionize` binary with GUI for Ubuntu Xenial to `$HOME/go/bin` do:
```shell
$ sudo docker build -t onionize-build:xenial -f Dockerfile.xenial .
$ sudo docker run -ti --rm -v $HOME/go/bin:/go/bin onionize-build:xenial
```

To build `onionize` binary with GUI for Debian Stretch to `$HOME/go/bin` do:
```shell
$ sudo docker build -t onionize-build:stretch -f Dockerfile.stretch .
$ sudo docker run -ti --rm -v $HOME/go/bin:/go/bin onionize-build:stretch
```
