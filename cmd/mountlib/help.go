package mountlib

// "@" will be replaced by the command name, "|" will be replaced by backticks
var mountHelp = `
ipfsdrive @ allows Linux, FreeBSD, macOS and Windows to
mount any of ipfsdrive's cloud storage systems as a file system with
FUSE.

First set up your remote using |ipfsdrive config|.  Check it works with |ipfsdrive ls| etc.

On Linux and macOS, you can run mount in either foreground or background (aka
daemon) mode. Mount runs in foreground mode by default. Use the |--daemon| flag
to force background mode. On Windows you can run mount in foreground only,
the flag is ignored.

In background mode ipfsdrive acts as a generic Unix mount program: the main
program starts, spawns background ipfsdrive process to setup and maintain the
mount, waits until success or timeout and exits with appropriate code
(killing the child process if it fails).

On Linux/macOS/FreeBSD start the mount like this, where |/path/to/local/mount|
is an **empty** **existing** directory:

    ipfsdrive @ remote:path/to/files /path/to/local/mount

On Windows you can start a mount in different ways. See [below](#mounting-modes-on-windows)
for details. If foreground mount is used interactively from a console window,
ipfsdrive will serve the mount and occupy the console so another window should be
used to work with the mount until ipfsdrive is interrupted e.g. by pressing Ctrl-C.

The following examples will mount to an automatically assigned drive,
to specific drive letter |X:|, to path |C:\path\parent\mount|
(where parent directory or drive must exist, and mount must **not** exist,
and is not supported when [mounting as a network drive](#mounting-modes-on-windows)), and
the last example will mount as network share |\\cloud\remote| and map it to an
automatically assigned drive:

    ipfsdrive @ remote:path/to/files *
    ipfsdrive @ remote:path/to/files X:
    ipfsdrive @ remote:path/to/files C:\path\parent\mount
    ipfsdrive @ remote:path/to/files \\cloud\remote

When the program ends while in foreground mode, either via Ctrl+C or receiving
a SIGINT or SIGTERM signal, the mount should be automatically stopped.

When running in background mode the user will have to stop the mount manually:

    # Linux
    fusermount -u /path/to/local/mount
    # OS X
    umount /path/to/local/mount

The umount operation can fail, for example when the mountpoint is busy.
When that happens, it is the user's responsibility to stop the mount manually.

The size of the mounted file system will be set according to information retrieved
from the remote, the same as returned by the [ipfsdrive about](https://ipfsdrive.org/commands/ipfsdrive_about/)
command. Remotes with unlimited storage may report the used size only,
then an additional 1 PiB of free space is assumed. If the remote does not
[support](https://ipfsdrive.org/overview/#optional-features) the about feature
at all, then 1 PiB is set as both the total and the free size.

### Installing on Windows

To run ipfsdrive @ on Windows, you will need to
download and install [WinFsp](http://www.secfs.net/winfsp/).

[WinFsp](https://github.com/billziss-gh/winfsp) is an open source
Windows File System Proxy which makes it easy to write user space file
systems for Windows.  It provides a FUSE emulation layer which ipfsdrive
uses combination with [cgofuse](https://github.com/billziss-gh/cgofuse).
Both of these packages are by Bill Zissimopoulos who was very helpful
during the implementation of ipfsdrive @ for Windows.

#### Mounting modes on windows

Unlike other operating systems, Microsoft Windows provides a different filesystem
type for network and fixed drives. It optimises access on the assumption fixed
disk drives are fast and reliable, while network drives have relatively high latency
and less reliability. Some settings can also be differentiated between the two types,
for example that Windows Explorer should just display icons and not create preview
thumbnails for image and video files on network drives.

In most cases, ipfsdrive will mount the remote as a normal, fixed disk drive by default.
However, you can also choose to mount it as a remote network drive, often described
as a network share. If you mount an ipfsdrive remote using the default, fixed drive mode
and experience unexpected program errors, freezes or other issues, consider mounting
as a network drive instead.

When mounting as a fixed disk drive you can either mount to an unused drive letter,
or to a path representing a **non-existent** subdirectory of an **existing** parent
directory or drive. Using the special value |*| will tell ipfsdrive to
automatically assign the next available drive letter, starting with Z: and moving backward.
Examples:

    ipfsdrive @ remote:path/to/files *
    ipfsdrive @ remote:path/to/files X:
    ipfsdrive @ remote:path/to/files C:\path\parent\mount
    ipfsdrive @ remote:path/to/files X:

Option |--volname| can be used to set a custom volume name for the mounted
file system. The default is to use the remote name and path.

To mount as network drive, you can add option |--network-mode|
to your @ command. Mounting to a directory path is not supported in
this mode, it is a limitation Windows imposes on junctions, so the remote must always
be mounted to a drive letter.

    ipfsdrive @ remote:path/to/files X: --network-mode

A volume name specified with |--volname| will be used to create the network share path.
A complete UNC path, such as |\\cloud\remote|, optionally with path
|\\cloud\remote\madeup\path|, will be used as is. Any other
string will be used as the share part, after a default prefix |\\server\|.
If no volume name is specified then |\\server\share| will be used.
You must make sure the volume name is unique when you are mounting more than one drive,
or else the mount command will fail. The share name will treated as the volume label for
the mapped drive, shown in Windows Explorer etc, while the complete
|\\server\share| will be reported as the remote UNC path by
|net use| etc, just like a normal network drive mapping.

If you specify a full network share UNC path with |--volname|, this will implicitely
set the |--network-mode| option, so the following two examples have same result:

    ipfsdrive @ remote:path/to/files X: --network-mode
    ipfsdrive @ remote:path/to/files X: --volname \\server\share

You may also specify the network share UNC path as the mountpoint itself. Then ipfsdrive
will automatically assign a drive letter, same as with |*| and use that as
mountpoint, and instead use the UNC path specified as the volume name, as if it were
specified with the |--volname| option. This will also implicitely set
the |--network-mode| option. This means the following two examples have same result:

    ipfsdrive @ remote:path/to/files \\cloud\remote
    ipfsdrive @ remote:path/to/files * --volname \\cloud\remote

There is yet another way to enable network mode, and to set the share path,
and that is to pass the "native" libfuse/WinFsp option directly:
|--fuse-flag --VolumePrefix=\server\share|. Note that the path
must be with just a single backslash prefix in this case.


*Note:* In previous versions of ipfsdrive this was the only supported method.

[Read more about drive mapping](https://en.wikipedia.org/wiki/Drive_mapping)

See also [Limitations](#limitations) section below.

#### Windows filesystem permissions

The FUSE emulation layer on Windows must convert between the POSIX-based
permission model used in FUSE, and the permission model used in Windows,
based on access-control lists (ACL).

The mounted filesystem will normally get three entries in its access-control list (ACL),
representing permissions for the POSIX permission scopes: Owner, group and others.
By default, the owner and group will be taken from the current user, and the built-in
group "Everyone" will be used to represent others. The user/group can be customized
with FUSE options "UserName" and "GroupName",
e.g. |-o UserName=user123 -o GroupName="Authenticated Users"|.

The permissions on each entry will be set according to
[options](#options) |--dir-perms| and |--file-perms|,
which takes a value in traditional [numeric notation](https://en.wikipedia.org/wiki/File-system_permissions#Numeric_notation),
where the default corresponds to |--file-perms 0666 --dir-perms 0777|.

Note that the mapping of permissions is not always trivial, and the result
you see in Windows Explorer may not be exactly like you expected.
For example, when setting a value that includes write access, this will be
mapped to individual permissions "write attributes", "write data" and "append data",
but not "write extended attributes". Windows will then show this as basic
permission "Special" instead of "Write", because "Write" includes the
"write extended attributes" permission.

If you set POSIX permissions for only allowing access to the owner, using
|--file-perms 0600 --dir-perms 0700|, the user group and the built-in "Everyone"
group will still be given some special permissions, such as "read attributes"
and "read permissions", in Windows. This is done for compatibility reasons,
e.g. to allow users without additional permissions to be able to read basic
metadata about files like in UNIX. One case that may arise is that other programs
(incorrectly) interprets this as the file being accessible by everyone. For example
an SSH client may warn about "unprotected private key file".

WinFsp 2021 (version 1.9) introduces a new FUSE option "FileSecurity",
that allows the complete specification of file security descriptors using
[SDDL](https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-string-format).
With this you can work around issues such as the mentioned "unprotected private key file"
by specifying |-o FileSecurity="D:P(A;;FA;;;OW)"|, for file all access (FA) to the owner (OW).

#### Windows caveats

Drives created as Administrator are not visible to other accounts,
not even an account that was elevated to Administrator with the
User Account Control (UAC) feature. A result of this is that if you mount
to a drive letter from a Command Prompt run as Administrator, and then try
to access the same drive from Windows Explorer (which does not run as
Administrator), you will not be able to see the mounted drive.

If you don't need to access the drive from applications running with
administrative privileges, the easiest way around this is to always
create the mount from a non-elevated command prompt.

To make mapped drives available to the user account that created them
regardless if elevated or not, there is a special Windows setting called
[linked connections](https://docs.microsoft.com/en-us/troubleshoot/windows-client/networking/mapped-drives-not-available-from-elevated-command#detail-to-configure-the-enablelinkedconnections-registry-entry)
that can be enabled.

It is also possible to make a drive mount available to everyone on the system,
by running the process creating it as the built-in SYSTEM account.
There are several ways to do this: One is to use the command-line
utility [PsExec](https://docs.microsoft.com/en-us/sysinternals/downloads/psexec),
from Microsoft's Sysinternals suite, which has option |-s| to start
processes as the SYSTEM account. Another alternative is to run the mount
command from a Windows Scheduled Task, or a Windows Service, configured
to run as the SYSTEM account. A third alternative is to use the
[WinFsp.Launcher infrastructure](https://github.com/billziss-gh/winfsp/wiki/WinFsp-Service-Architecture)).
Note that when running ipfsdrive as another user, it will not use
the configuration file from your profile unless you tell it to
with the [|--config|](https://ipfsdrive.org/docs/#config-config-file) option.
Read more in the [install documentation](https://ipfsdrive.org/install/).

Note that mapping to a directory path, instead of a drive letter,
does not suffer from the same limitations.

### Limitations

Without the use of |--vfs-cache-mode| this can only write files
sequentially, it can only seek when reading.  This means that many
applications won't work with their files on an ipfsdrive mount without
|--vfs-cache-mode writes| or |--vfs-cache-mode full|.
See the [VFS File Caching](#vfs-file-caching) section for more info.

The bucket based remotes (e.g. Swift, S3, Google Compute Storage, B2,
Hubic) do not support the concept of empty directories, so empty
directories will have a tendency to disappear once they fall out of
the directory cache.

When |ipfsdrive mount| is invoked on Unix with |--daemon| flag, the main ipfsdrive
program will wait for the background mount to become ready or until the timeout
specified by the |--daemon-wait| flag. On Linux it can check mount status using
ProcFS so the flag in fact sets **maximum** time to wait, while the real wait
can be less. On macOS / BSD the time to wait is constant and the check is
performed only at the end. We advise you to set wait time on macOS reasonably.

Only supported on Linux, FreeBSD, OS X and Windows at the moment.

### ipfsdrive @ vs ipfsdrive sync/copy

File systems expect things to be 100% reliable, whereas cloud storage
systems are a long way from 100% reliable. The ipfsdrive sync/copy
commands cope with this with lots of retries.  However ipfsdrive @
can't use retries in the same way without making local copies of the
uploads. Look at the [VFS File Caching](#vfs-file-caching)
for solutions to make @ more reliable.

### Attribute caching

You can use the flag |--attr-timeout| to set the time the kernel caches
the attributes (size, modification time, etc.) for directory entries.

The default is |1s| which caches files just long enough to avoid
too many callbacks to ipfsdrive from the kernel.

In theory 0s should be the correct value for filesystems which can
change outside the control of the kernel. However this causes quite a
few problems such as
[ipfsdrive using too much memory](https://github.com/ipfsdrive/ipfsdrive/issues/2157),
[ipfsdrive not serving files to samba](https://forum.ipfsdrive.org/t/ipfsdrive-1-39-vs-1-40-mount-issue/5112)
and [excessive time listing directories](https://github.com/ipfsdrive/ipfsdrive/issues/2095#issuecomment-371141147).

The kernel can cache the info about a file for the time given by
|--attr-timeout|. You may see corruption if the remote file changes
length during this window.  It will show up as either a truncated file
or a file with garbage on the end.  With |--attr-timeout 1s| this is
very unlikely but not impossible.  The higher you set |--attr-timeout|
the more likely it is.  The default setting of "1s" is the lowest
setting which mitigates the problems above.

If you set it higher (|10s| or |1m| say) then the kernel will call
back to ipfsdrive less often making it more efficient, however there is
more chance of the corruption issue above.

If files don't change on the remote outside of the control of ipfsdrive
then there is no chance of corruption.

This is the same as setting the attr_timeout option in mount.fuse.

### Filters

Note that all the ipfsdrive filters can be used to select a subset of the
files to be visible in the mount.

### systemd

When running ipfsdrive @ as a systemd service, it is possible
to use Type=notify. In this case the service will enter the started state
after the mountpoint has been successfully set up.
Units having the ipfsdrive @ service specified as a requirement
will see all files and folders immediately in this mode.

Note that systemd runs mount units without any environment variables including
|PATH| or |HOME|. This means that tilde (|~|) expansion will not work
and you should provide |--config| and |--cache-dir| explicitly as absolute
paths via ipfsdrive arguments.
Since mounting requires the |fusermount| program, ipfsdrive will use the fallback
PATH of |/bin:/usr/bin| in this scenario. Please ensure that |fusermount|
is present on this PATH.

### ipfsdrive as Unix mount helper

The core Unix program |/bin/mount| normally takes the |-t FSTYPE| argument
then runs the |/sbin/mount.FSTYPE| helper program passing it mount options
as |-o key=val,...| or |--opt=...|. Automount (classic or systemd) behaves
in a similar way.

ipfsdrive by default expects GNU-style flags |--key val|. To run it as a mount
helper you should symlink ipfsdrive binary to |/sbin/mount.ipfsdrive| and optionally
|/usr/bin/ipfsdrivefs|, e.g. |ln -s /usr/bin/ipfsdrive /sbin/mount.ipfsdrive|.
ipfsdrive will detect it and translate command-line arguments appropriately.

Now you can run classic mounts like this:
|||
mount sftp1:subdir /mnt/data -t ipfsdrive -o vfs_cache_mode=writes,sftp_key_file=/path/to/pem
|||

or create systemd mount units:
|||
# /etc/systemd/system/mnt-data.mount
[Unit]
After=network-online.target
[Mount]
Type=ipfsdrive
What=sftp1:subdir
Where=/mnt/data
Options=rw,allow_other,args2env,vfs-cache-mode=writes,config=/etc/ipfsdrive.conf,cache-dir=/var/ipfsdrive
|||

optionally accompanied by systemd automount unit
|||
# /etc/systemd/system/mnt-data.automount
[Unit]
After=network-online.target
Before=remote-fs.target
[Automount]
Where=/mnt/data
TimeoutIdleSec=600
[Install]
WantedBy=multi-user.target
|||

or add in |/etc/fstab| a line like
|||
sftp1:subdir /mnt/data ipfsdrive rw,noauto,nofail,_netdev,x-systemd.automount,args2env,vfs_cache_mode=writes,config=/etc/ipfsdrive.conf,cache_dir=/var/cache/ipfsdrive 0 0
|||

or use classic Automountd.
Remember to provide explicit |config=...,cache-dir=...| as a workaround for
mount units being run without |HOME|.

ipfsdrive in the mount helper mode will split |-o| argument(s) by comma, replace |_|
by |-| and prepend |--| to get the command-line flags. Options containing commas
or spaces can be wrapped in single or double quotes. Any inner quotes inside outer
quotes of the same type should be doubled.

Mount option syntax includes a few extra options treated specially:

- |env.NAME=VALUE| will set an environment variable for the mount process.
  This helps with Automountd and Systemd.mount which don't allow setting
  custom environment for mount helpers.
  Typically you will use |env.HTTPS_PROXY=proxy.host:3128| or |env.HOME=/root|
- |command=cmount| can be used to run |cmount| or any other ipfsdrive command
  rather than the default |mount|.
- |args2env| will pass mount options to the mount helper running in background
  via environment variables instead of command line arguments. This allows to
  hide secrets from such commands as |ps| or |pgrep|.
- |vv...| will be transformed into appropriate |--verbose=N|
- standard mount options like |x-systemd.automount|, |_netdev|, |nosuid| and alike
  are intended only for Automountd and ignored by ipfsdrive.
`
