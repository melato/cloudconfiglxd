lxd-cloudinit configures LXD instances (containers, VMs), by
applying cloud-init config files, using the LXD API.

It does not require that the guest OS has cloud-init installed.

# Supported cloud-init features
The following fields of cloud-config files are supported and applied in this order:

- bootcmd (see below)
- package_update
- package_upgrade
- packages
- write_files (only plain text, without any encoding)
- users
- timezone
- runcmd

# bootcmd
This is like runcmd, but it is applied before anything else.
Unlike cloud-config, it does not install the commands to run on every boot.
