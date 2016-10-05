onionize
===========
Make static onion site up and running from any directory.
Or share a file via one-shot onion site.
Onion sites are end-to-end encrypted, metadata-free and forward-secure.
Much love to onion services.

Install
-------
```
$ torsocks go get github.com/nogoegst/onionize
```

Usage
-----
```
$ onionize /path/to/my/tiny/webroot
```

Grab onion address from `stdout` and errors/info from `stderr`.
 
That's it.

Also you can `onionize` contents of a `zip` file:

```
$ onionize /path/to/my/tiny/webroot.zip
```

You also can share a file as `oignonshare` does:

```
$ onionize /path/to/my/secret.file
```
