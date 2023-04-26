cloudinitlxd configures LXD instances (containers, VMs), by
applying cloud-init config files, using the LXD API.

It does not require that the guest OS has cloud-init installed.

It can be used as a Go library or as an executable.

# Supported cloud-init features
The following cloud-init modules (sections) are supported and applied in this order:

- packages
- write_files (only plain text, without any encoding)
- users
- runcmd

# Standalone executable
cloudconfig-lxd will connect to the LXD server
using either the UNIX socket or HTTPS.
For HTTPS, it uses configuration information found in the lxc user configuration files,
in ~/snap/lxd/common/config/

# Usage
```
cloudconfig-lxd apply [-ostype <ostype>] -i <instance> <cloud-config-file>...
```
ostype is needed to for packages and users, since different distributions have
different package systems and may have differences in how they create users.

Supported ostypes are: alpine, debian.  Others can be added.
  
# compile

```
cd main
date > version
go install cloudconfig-lxd.go
```
