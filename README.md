onionize
===========
Make an onion site (aka HTTP over onion services) up and running from any a
directory/file/zip archive.

Onion services are end-to-end encrypted, metadata-free and forward-secure (see design overview [1](https://www.torproject.org/docs/hidden-services.html.en)).
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
$ onionize [||-zip] /path/to/my/[directory|file|archive.zip]
```

Grab the onion link from `stdout` and errors/info from `stderr`.
 
That's it.
