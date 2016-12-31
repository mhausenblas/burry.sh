# burry

This is the `burry`, the BackUp & RecoveRY tool for cloud native infrastructure services. Use `burry` to back up and restore
critical infrastructure base services such as ZooKeeper and etcd.

`burry` support back up/restore the following infrastructure services with the respective storage targets:

|to/from         |ZooKeeper   |etcd        |
| --------------:| ---------- | ---------- |
| Amazon S3      | yes        | backlog    |
| Azure Storage  | backlog    | backlog    |
| Google Storage | backlog    | backlog    |
| Local          | yes        | WIP        |
| Minio*         | yes        | backlog    |
| TTY**          | yes        | WIP        |

```
 *) Minio can be either on-premises or in the cloud, but always self-hosted. See also https://www.minio.io
**) TTY effectively means it's not stored at all but rather dumped on the screen; useful for debugging, though.
```

## Architecture

The essence of burry's algorithm is:

- Until user cancels
  - Either on changes or every `AT_LEAST_SEC`
  - Walk the tree from root
  - Retrieve data and metadata from each non-ephemeral node
  - Write all data and metadata to storage target

## Install

TBD.

## Use

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

Policy is:

1. If no manifest `.burryfest` exists in the current directory, the command line parameters passed are used to create a manifest file.
1. If a manifest `.burryfest` exists in the current directory it will be used, use `--overwrite` to temporarily overwrite parameters.
1. For every storage target other than `tty` a new manifest in the timestamped archive file will be created.

Examples usages of `burry` follow.

### Back up DC/OS system ZooKeeper to Amazon S3

```bash
# let's first do a dry run, that is, only dump to screen.
# this works because the default value of --target is 'tty'
$ ./burry.sh --endpoint leader.mesos:2181
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
$ ./burry.sh --endpoint leader.mesos:2181 --target s3 --overwrite
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





