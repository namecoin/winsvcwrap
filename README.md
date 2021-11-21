winsvcwrap
==========

This is an open source replacement for `SRVANY.EXE`.

Description of SRVANY
---------------------

`SRVANY.EXE` was a proprietary tool included in the Windows 2003 Resource Kit
and no longer maintained by Microsoft. It also came with a utility
`INSTSRV.EXE` which installs services which use `SRVANY.EXE`.

The `INSTSRV` utility used the following arguments:

    instsrv <service-name> <path-to-SRVANY.EXE> [-a <account-name>] [-p <account-password>]
    instsrv <service-name> remove

This creates a service using `SRVANY` as the service executable.

Windows stores service configuration information in `HKLM\SYSTEM\CurrentControlSet\Services\<service-name>`. `SRVANY` needs these configuration parameters to be added to its key, beyond the standard Windows ones:

  - `Parameters\Application`: String. Required. The actual service executable to launch and be supervised by `SRVANY`.
  - `Parameters\AppParameters`: String. Optional. Parameters to pass to the actual service executable.
  - `Parameters\AppEnvironment`: Multi-String. Optional. Environment variables to pass to the actual service executable.
  - `Parameters\AppDirectory`: String. Optional. Working directory to run actual service executable under.

Description of winsvcwrap
-------------------------

winsvcwrap is a simple Go daemon which can be hosted as a Windows service and
which spawns and supervises one other process. It uses
[hlandau/service](https://github.com/hlandau/service) and inherits its Windows
service support from there.

    winsvcwrap -service.*=...

For simplicity, arguments are taken from the service command line rather than
separate registry keys as in the `SRVANY` case. The following command line
arguments are supported:

    winsvcwrap
      -winsvcwrap.run=...            Windows commandline to spawn (quote both EXE and arguments).
      -winsvcwrap.arg=...            Add an argument to the command to spawn. May be specified multiple times.
      -winsvcwrap.cwd=...            Set current working directory to this before spawning (optional).

winsvcwrap will propagate failures. That is, if the spawned process fails for
any reason, winsvcwrap will also fail and exit so that the Windows service
management system can detect this failure, log it, and restart the entire
hierarchy, winsvcwrap and all.

winsvcwrap generates log output using xlog, so you can logs its (small number
of) log messages using any method supported by xlog. This is separate from the
logging of the daemon it supervises.

TODO
----

- Send the Windows equivalent of SIGINT instead of just terminating the child process.
  This appears to be hard to do.

Licence
-------

Licenced under the GPLv3 or later.

© 2021 [Hugo Landau](https://www.devever.net/~hl/)
