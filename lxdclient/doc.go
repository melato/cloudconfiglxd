/*
	Package lxdclient connects to an LXD server.

It connects via:
- A unix socket, defaults to /var/snap/lxd/common/lxd/unix.socket
- or via HTTP, using the url found in the configuration files below.

It uses the same configuration files used by the "lxc" command.
It looks for configuration files at the following directories:
- LxdClient.ConfigDir, if non-empty (set it via command-line options)
- The value of the LXD_CONF environment variable, if it exists.

The first of the following directories that exists:
- $HOME/snap/lxd/common/config
- $HOME/snap/lxd/current/.config/lxc
- $CONFIGDIR/lxc

Where:

	$HOME = os.UserHomeDir()
	$CONFIGDIR = os.UserConfigDir() (usually $HOME/.config, on Unix systems)
*/
package lxdclient
