## Kamstrup Multical (601) Reader/Server

This little project runs a web server that can query the Kamstrup Multical 601's registers on demand. It'll most likely
work with other Multical meters as well, but I'm not in a position to test that.

I'm using [this USB IR-transceiver](https://de.elv.com/elv-bausatz-lesekopf-mit-usb-schnittstelle-fuer-digitale-zaehler-usb-iec-155523) (requires some soldering).

The Makefile I'm using to deploy it to the Raspberry Pi running the server will need tweaking for your situation (SSH host, remote directories, etc).  