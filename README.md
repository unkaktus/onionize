onionize
===========
Make static onion site up and running from any directory.
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
