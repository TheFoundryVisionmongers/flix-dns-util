# Flix DNS Util
This utility application can be used to test the DNS resolution; designed to check compatibility of Flix in a
split-tunnel VPN environment.

## Directions
Download the appropriate binary from the releases page for your operating system.  This binary should be run from where
you expect to run the Flix client.

The utility binary requires a command line flag to be set `hostname`.  This should be the hostname you would log into
using the Flix client, excluding the protocol (the `http://` or `https://` at the beginning).

The second required command line flag to be set is `port`.  This should be the port you would connect to the Flix server
with.

Optionally, you can include the flag `--use-tls`.  Set this flag if you connect with `https://`, and not `http://`.

If you would log into `http://flix.foundry.com:8181`, you should run the utility with this command:
```bash
./flix-dns-util --hostname=flix.foundry.com --port=8181
```

## Output
The Flix DNS utility will create an output log in the terminal.  You may find that a large number of the tests fail;
this is not a cause for concern; the utility checks a variety of things, many of which may not be needed for your setup.

Copy the output from the terminal and provide this to the Foundry support representative. 
