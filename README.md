onionize
===========
Make an onion site (aka HTTP over onion services) up and running from any a
directory/file/zip archive.

Onion services are end-to-end encrypted, metadata-free and forward-secure (see design overview [1](https://www.torproject.org/docs/hidden-services.html.en)).
Much love to onion services.

Install
-------
CLI-only version:
```
$ go get github.com/nogoegst/onionize
```
CLI + GUI version (reqiures GTK3 installed):
```
$ go get -tags gui github.com/nogoegst/onionize
```

Usage
-----
To onionize a [directory|file|zip archive]:

```
$ onionize [||-zip] /path/to/my/[directory|file|archive.zip]
```

Grab the onion link from `stdout` and errors/info from `stderr`.
 
That's it.

GUI mode
--------
To run `onionize` in GUI mode just don't specify any path.
Select target type, open your file/directory and click `onionize`:

![onionize GUI screenshot](docs/onionize-dir-1.png)

Then you can grab the link from a text field:

![onionize GUI screenshot](docs/onionize-dir-2.png)

You can find more screenshots in `docs`.

Identity passphrase
-------------------

You may also specify a passphrase from which onion service identity key
will be derived. Thus you can preserve same .onion address across setups.

Identity passphrase can be specified on `stdin` by setting `-p` flag in CLI
or in corresponding field in GUI.
