cloudinitlxd configures LXD instances (containers, VMs), by
applying cloud-init config files, using the LXD API.

It does not require that the guest OS has cloud-init installed.

It can be used as a library or as an executable.


# Supported cloud-init features
The following fields of cloud-config files are supported and applied in this order:

- packages
- write_files (only plain text, without any encoding)
- users
- runcmd

# compile

```
cd main
date > version
go build cloudconfig-lxd 
```
