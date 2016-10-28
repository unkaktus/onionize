onionize
===========
Make an onion site up and running from any a
directory/file/zip archive.

Onion sites are end-to-end encrypted, metadata-free and forward-secure.
Much love to onion services.

Install
-------
```
$ go get github.com/nogoegst/onionize
```

Usage
-----
To onionize a [directory|file|zip archive]:

```
$ onionize [||-z] /path/to/my/[directory|file|archive.zip]
```

Grab the onion link from `stdout` and errors/info from `stderr`.
 
That's it.
