# burry

This is the `burry`, the BackUp & RecoveRY tool for cloud native infrastructure services. Use `burry` to back up and restore
critical infrastructure base services such as ZooKeeper and etcd.

`burry` currently supports backing up the following infra services (`from`) into storage targets (`to`):

|to/from         |ZooKeeper   |etcd        |
| --------------:| ---------- | ---------- |
| Amazon S3      | yes        | yes        |
| Azure Storage  | backlog    | backlog    |
| Google Storage | backlog    | backlog    |
| Local          | yes        | yes        |
| Minio*         | yes        | yes        |
| TTY**          | yes        | yes        |

```
 *) Minio can be either on-premises or in the cloud, but always self-hosted. See also https://www.minio.io
**) TTY effectively means it's not stored at all but rather dumped on the screen; useful for debugging, though.
```

**Note that restoring infrastructure services from storage targets is NOT YET implemented.**

Contents:

- [Install](#install)
- [Use](#use)
  - Example: [Screen dump of local ZooKeeper content](#screen-dump-of-local-zookeeper-content)
  - Example: [Back up DC/OS system ZooKeeper to Amazon S3](#back-up-dcos-system-zookeeper-to-amazon-s3)
  - Example: [Back up etcd to Minio](#https://github.com/mhausenblas/burry.sh#back-up-etcd-to-minio)
- [Architecture](#architecture)

## Install

Currently, only 'build from source' install is available (note: replace `GOOS=linux` with your platform):

    $ go get github.com/mhausenblas/burry.sh
    $ GOOS=linux go build
    $ mv burry.sh burry
    $ godoc -http=":6060" &
    $ open http://localhost:6060/pkg/github.com/mhausenblas/burry.sh/

See also [GoDocs](https://godoc.org/github.com/mhausenblas/burry.sh).

## Use

The general usage is:

```bash
$ burry --endpoint IP:PORT (--isvc etcd|zk) (--target tty|local|s3) (--overwrite)
```

So, the only required parameter really is the `--endpoint`. When run the first time, `burry` creates a manifest file in the current directory called `.burryfest`, capturing all your settings. Subsequent invocations hence are simply `burry`, without any parameters. Use  `--overwrite` to temporarily overwrite parameters or remove the `.burryfest` file for permanent changes.

All parameters:

```bash
$ burry --help
Usage: burry [args]

Arguments:
  -endpoint string
        The infra service HTTP API endpoint to use. Example: localhost:8181 for Exhibitor
  -isvc string
        The type of infra service to back up or restore. Supported values are [etcd zk] (default "zk")
  -overwrite
        Command line values overwrite manifest values
  -target string
        The storage target to use. Supported values are [local tty] (default "tty")
  -version
        Display version information
```

Parameter reuse policy is  as follows:

1. If no manifest `.burryfest` exists in the current directory, the command line parameters passed are used to create a manifest file.
1. If a manifest `.burryfest` exists in the current directory it will be used, use `--overwrite` to temporarily overwrite parameters.
1. For every storage target other than `tty` a new manifest in the timestamped archive file will be created.

Next we will have a look of some examples usages of `burry`.

### Screen dump of local ZooKeeper content

See the [development and testing](dev.md#zookeeper) notes for the test setup.

```bash
$ docker ps
CONTAINER ID        IMAGE                                  COMMAND                  CREATED             STATUS              PORTS                                                                                            NAMES
9ae41a9a02f8        mbabineau/zookeeper-exhibitor:latest   "bash -ex /opt/exhibi"   2 days ago          Up 2 days           0.0.0.0:2181->2181/tcp, 0.0.0.0:2888->2888/tcp, 0.0.0.0:3888->3888/tcp, 0.0.0.0:8181->8181/tcp   amazing_kilby

$ DEBUG=true ./burry --endpoint localhost:2181
INFO[0000] Using existing burry manifest file /home/core/.burryfest  func=init
INFO[0000] My config: {InfraService:zk Endpoint:localhost:2181 StorageTarget:tty Credentials:}  func=init
INFO[0000] On node /                                     func=visitZK
2016/12/31 09:27:25 Connected to [::1]:2181
2016/12/31 09:27:25 Authenticated: id=97189781074870273, timeout=4000
DEBU[0000] / has 1 children                              func=visitZK
DEBU[0000] Next visiting child /zookeeper                func=visitZK
INFO[0000] On node /zookeeper                            func=visitZK
DEBU[0000] /zookeeper has 1 children                     func=visitZK
DEBU[0000] Next visiting child /zookeeper/quota          func=visitZK
INFO[0000] On node /zookeeper/quota                      func=visitZK
DEBU[0000] /zookeeper/quota has 0 children               func=visitZK
INFO[0000] /zookeeper/quota:                             func=reapsimple
DEBU[0000]                                               func=reapsimple
INFO[0000] Operation successfully completed.             func=main
```

### Back up DC/OS system ZooKeeper to Amazon S3

See the [development and testing](dev.md#zookeeper) notes for the test setup.

```bash
# let's first do a dry run, that is, only dump to screen.
# this works because the default value of --target is 'tty'
$ ./burry --endpoint leader.mesos:2181
INFO[0000] My config: {InfraService:zk Endpoint:leader.mesos:2181 StorageTarget:tty Credentials:}  func=init
INFO[0000] Created burry manifest file /tmp/.burryfest  func=writebf
INFO[0000] On node /                                     func=visit
2016/12/31 06:20:07 Connected to 10.0.4.185:2181
2016/12/31 06:20:07 Authenticated: id=97172902550700662, timeout=4000
INFO[0000] On node /mesos                                func=visit
...
INFO[0006] /etcd/etcd_framework_id:                      func=rznode
INFO[0006] Operation successfully completed.             func=main

$ cat .burryfest
{"svc":"zk","endpoint":"leader.mesos:2181","target":"tty","credentials":""}

# now we know we can read stuff from ZK, let's get it
# backed up into Amazon S3; you can either remove
# .burryfest or use --overwrite to specify the new storage target
$ ./burry --endpoint leader.mesos:2181 --target s3 --overwrite
INFO[0008] Using existing burry manifest file /tmp/.burryfest  func=init
INFO[0000] My config: {InfraService:zk Endpoint:leader.mesos:2181 StorageTarget:s3 Credentials:}  func=init
INFO[0000] On node /                                     func=visit
2016/12/31 06:41:46 Connected to 10.0.4.185:2181
2016/12/31 06:41:46 Authenticated: id=97172902550700682, timeout=4000
INFO[0000] On node /mesos                                func=visit
...
INFO[0006] On node /etcd/etcd_framework_id               func=visit
INFO[0006] Backup available in /tmp/1483166506.zip       func=arch
INFO[0006] Trying to back up to zk-backup-1483166506/latest.zip in Amazon S3  func=remoteS3
INFO[0008] Successfully stored zk-backup-1483166506/latest.zip (45464 Bytes) in Amazon S3  func=remoteS3
INFO[0008] Operation successfully completed.             func=main

```

### Back up etcd to Minio

See the [development and testing](dev.md#etcd) notes for the test setup.

```bash
$ ./burry --endpoint etcd.mesos:1026 --isvc etcd --target s3
INFO[0000] Using existing burry manifest file /tmp/.burryfest  func=init
INFO[0000] My config: {InfraService:etcd Endpoint:etcd.mesos:1026 StorageTarget:s3 Credentials:}  func=init
INFO[0000] On node /                                     func=visitETCD
INFO[0000] On node /foo                                  func=visitETCD
INFO[0000] On node /meh                                  func=visitETCD
INFO[0000] On node /buz                                  func=visitETCD
INFO[0000] On node /buz/meh                              func=visitETCD
INFO[0000] Adding /tmp/.burryfest to /tmp/1483173687     func=addbf
INFO[0000] Backup available in /tmp/1483173687.zip       func=arch
INFO[0000] Trying to back up to etcd-backup-1483173687/latest.zip in Amazon S3  func=remoteS3
INFO[0001] Successfully stored etcd-backup-1483173687/latest.zip (674 Bytes) in Amazon S3  func=remoteS3
INFO[0001] Operation successfully completed.             func=main
```

## Architecture

`burry` assumes that the infra service it operates on is tree-like. The essence of `burry`'s algorithm is:

- Walk the tree from the root
- For every non-leaf node: process its children
- For every leaf node, store the content (that is, the node value) 
- Depending on the storage target selected, create archive incl. metadata