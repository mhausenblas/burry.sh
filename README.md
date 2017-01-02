# burry

This is `burry`, the BackUp & RecoveRY tool for cloud native infrastructure services. Use `burry` to back up and restore
critical infrastructure base services such as ZooKeeper and etcd. [More…](https://hackernoon.com/backup-recovery-of-infrastructure-services-200b2116930f)

`burry` currently supports the following infra services and storage targets:

|                |ZooKeeper   |etcd        |
| --------------:| ---------- | ---------- |
| Amazon S3      | B/R        | B/-        |
| Azure Storage  | []/[]      | []/[]      |
| Google Storage | []/[]      | []/[       |
| Local          | B/R        | B/-        |
| Minio*         | B/R        | B/-        |
| TTY**          | B/-        | B/-        |

```
 B  ... backups supported
 R  ... restores supported
 -  ... not applicable
 [] ... not yet implemented
 *) Minio can be either on-premises or in the cloud, but always self-hosted. See also https://www.minio.io
**) TTY effectively means it's not stored at all but rather dumped on the screen; useful for debugging, though.
```

Note: **this is WIP, please use with care. Only non-existing nodes or keys will be restored, that is, no existing data in ZK or etcd will be overwritten when attempting to restore data**.

**Contents:**

- [Install](#install)
- [Use](#use)
  - [Backups](#backups)
    - Example: [Screen dump of local ZooKeeper content](#screen-dump-of-local-zookeeper-content)
    - Example: [Back up etcd to local storage](#back-up-etcd-to-local-storage)
    - Example: [Back up DC/OS system ZooKeeper to Amazon S3](#back-up-dcos-system-zookeeper-to-amazon-s3)
    - Example: [Back up etcd to Minio](#back-up-etcd-to-minio)
  - [Restores](#restores)
    - Example: [Restore etcd from local storage](#restore-etcd-from-local-storage)
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
$ burry --help
Usage: burry [args]

Arguments:
  -c, --credentials string
        The credentials to use in format STORAGE_TARGET_ENDPOINT,KEY1=VAL1,...KEYn=VALn.
        Example: s3.amazonaws.com,AWS_ACCESS_KEY_ID=...,AWS_SECRET_ACCESS_KEY=...
  -e, --endpoint string
        The infra service HTTP API endpoint to use.
        Example: localhost:8181 for Exhibitor
  -i, --isvc string
        The type of infra service to back up or restore.
        Supported values are [etcd zk] (default "zk")
  -o, --operation string
        The operation to carry out.
        Supported values are [backup restore] (default "backup")
  -w, --overwrite
        Make command line values overwrite manifest values.
  -s, --snapshot string
        The ID of the snapshot.
        Example: 1483193387
  -t, --target string
        The storage target to use.
        Supported values are [local minio s3 tty] (default "tty")
  -v, --version
        Display version information and exit.
```

When run the first time, `burry` creates a manifest file in the current directory called `.burryfest`, capturing all your settings. If a manifest `.burryfest` exists in the current directory subsequent invocations use this and hence you can simply execute `burry`, without any parameters. Use  `--overwrite` to temporarily overwrite command line parameters or remove the `.burryfest` file for permanent changes.

An example of a burry manifest file looks like:

```json
{
    "svc": "etcd",
    "svc-endpoint": "etcd.mesos:1026",
    "target": "local",
    "credentials": {
        "target-endpoint": "",
        "params": []
    }
}
```

Note that for every storage target other than `tty` a metadata file `.burrymeta` in the (timestamped) archive file will be created, something like:

```json
{
  "snapshot-date": "2016-12-31T14:52:42Z",
  "svc": "zk",
  "svc-endpoint": "leader.mesos:2181",
  "target": "s3",
  "target-endpoint": "s3.amazonaws.com"
}
```

### Backups

In general, since `--operation backup` is the default, the only required parameter for a backup operation is the `--endpoint`, that is, the HTTP API of the ZooKeeper or etcd you want to back up.

```bash
$ burry --endpoint IP:PORT (--operation backup) (--isvc etcd|zk) (--target tty|local|s3) (--overwrite) (--credentials STORAGE_TARGET_ENDPOINT,KEY1=VAL1,KEY2=VAL2,...KEYn=VALn)
```

#### Screen dump of local ZooKeeper content

See the [development and testing](dev.md#zookeeper) notes for the test setup.

```bash
$ docker ps
CONTAINER ID        IMAGE                                  COMMAND                  CREATED             STATUS              PORTS                                                                                            NAMES
9ae41a9a02f8        mbabineau/zookeeper-exhibitor:latest   "bash -ex /opt/exhibi"   2 days ago          Up 2 days           0.0.0.0:2181->2181/tcp, 0.0.0.0:2888->2888/tcp, 0.0.0.0:3888->3888/tcp, 0.0.0.0:8181->8181/tcp   amazing_kilby

$ DEBUG=true ./burry --endpoint localhost:2181
INFO[0000] Using existing burry manifest file /home/core/.burryfest  func=init
INFO[0000] My config: {InfraService:zk Endpoint:localhost:2181 StorageTarget:tty Creds:{StorageTargetEndpoint: Params:[]}}  func=init
INFO[0000] /zookeeper/quota:                             func=reapsimple
INFO[0000] Operation successfully completed.             func=main
```

#### Back up etcd to local storage 

See the [development and testing](dev.md#etcd) notes for the test setup.

```bash
$ ./burry --endpoint etcd.mesos:1026 --isvc etcd --target local
INFO[0000] My config: {InfraService:etcd Endpoint:etcd.mesos:1026 StorageTarget:local Creds:{StorageTargetEndpoint: Params:[]}}  func=init
INFO[0000] Created burry manifest file /tmp/.burryfest  func=writebf
INFO[0000] Added metadata to /tmp/1483193387             func=addmeta
INFO[0000] Backup available in /tmp/1483193387.zip       func=arch
INFO[0000] Operation successfully completed.             func=main

$ ls -al *.zip
-rw-r--r--@ 1 mhausenblas  staff  750 31 Dec 14:22 1483194168.zip

$ unzip 1483194168.zip

$ cat 1483194168/.burrymeta | jq .
{
  "snapshot-date": "2016-12-31T14:22:48Z",
  "svc": "etcd",
  "svc-endpoint": "etcd.mesos:1026",
  "target": "local",
  "target-endpoint": "/tmp"
}
```

#### Back up DC/OS system ZooKeeper to Amazon S3

See the [development and testing](dev.md#zookeeper) notes for the test setup.

```bash
# let's first do a dry run, that is, only dump to screen.
# this works because the default value of --target is 'tty'
$ ./burry --endpoint leader.mesos:2181
INFO[0000] My config: {InfraService:zk Endpoint:leader.mesos:2181 StorageTarget:tty Creds:{StorageTargetEndpoint: Params:[]}}  func=init
INFO[0000] Created burry manifest file /tmp/.burryfest  func=writebf
INFO[0006] Operation successfully completed.             func=main

# now we know we can read stuff from ZK, let's get it
# backed up into Amazon S3; you can either remove
# .burryfest or use --overwrite to specify the new storage target
$ ./burry --endpoint leader.mesos:2181 --target s3 --credentials s3.amazonaws.com,AWS_ACCESS_KEY_ID=***,AWS_SECRET_ACCESS_KEY=***
INFO[0000] Using existing burry manifest file /tmp/.burryfest  func=init
INFO[0000] My config: {InfraService:zk Endpoint:leader.mesos:2181 StorageTarget:s3 Creds:{InfraServiceEndpoint:s3.amazonaws.com Params:[{Key:AWS_ACCESS_KEY_ID Value:***} {Key:AWS_SECRET_ACCESS_KEY Value:***}]}}}  func=init
INFO[0006] Backup available in /tmp/1483166506.zip       func=arch
INFO[0006] Trying to back up to zk-backup-1483166506/latest.zip in S3 compatible remote storage  func=remoteS3
INFO[0008] Successfully stored zk-backup-1483166506/latest.zip (45464 Bytes) in S3 compatible remote storage s3.amazonaws.com  func=remoteS3
INFO[0008] Operation successfully completed.             func=main
```

#### Back up etcd to Minio 

See the [development and testing](dev.md#etcd) notes for the test setup. Note: the credentials used below are from the public [Minio playground](https://play.minio.io:9000/).

```bash
$ ./burry --endpoint etcd.mesos:1026 --isvc etcd --credentials play.minio.io:9000,AWS_ACCESS_KEY_ID=Q3AM3UQ867SPQQA43P2F,AWS_SECRET_ACCESS_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG --target s3
INFO[0000] Using existing burry manifest file /tmp/.burryfest  func=init
INFO[0000] My config: {InfraService:etcd Endpoint:etcd.mesos:1026 StorageTarget:s3 Credentials:}  func=init
INFO[0000] Adding /tmp/.burryfest to /tmp/1483173687     func=addbf
INFO[0000] Backup available in /tmp/1483173687.zip       func=arch
INFO[0000] Trying to back up to etcd-backup-1483173687/latest.zip in S3 compatible remote storage  func=remoteS3
INFO[0001] Successfully stored etcd-backup-1483173687/latest.zip (674 Bytes) in S3 compatible remote storage play.minio.io:9000  func=remoteS3
INFO[0001] Operation successfully completed.             func=main
```


### Restores

For restores you MUST set `--operation restore` as well as provide a `--snapshot` ID and note that you CAN NOT restore from screen, that is, `--target tty` is an invalid choice:

```bash
$ burry --operation restore --snapshot ID --target local|s3 (--isvc etcd|zk) (--overwrite) (--credentials STORAGE_TARGET_ENDPOINT,KEY1=VAL1,KEY2=VAL2,...KEYn=VALn)
```

#### Restore etcd from local storage 

See the [development and testing](dev.md#etcd) notes for the test setup.

```bash
# let's first back up etcd:
$ ./burry -e etcd.mesos:1026 -i etcd -t local
INFO[0000] Selected operation: BACKUP                    func=init
INFO[0000] My config: {InfraService:etcd Endpoint:10.0.1.139:1026 StorageTarget:local Creds:{StorageTargetEndpoint: Params:[]}}  func=init
INFO[0000] Added metadata to /tmp/1483383204  func=addmeta
INFO[0000] Backup available in /tmp/1483383204.zip  func=arch
INFO[0000] Created burry manifest file /tmp/.burryfest  func=writebf
INFO[0000] Operation successfully completed. The snapshot ID is: 1483383204  func=main
# now, let's destroy a key:
$ curl etcd.mesos:1026/v2/keys/foo -XDELETE
{"action":"delete","node":{"key":"/foo","modifiedIndex":16,"createdIndex":15},"prevNode":{"key":"/foo","value":"bar","modifiedIndex":15,"createdIndex":15}}
# ... and restore it from the backup:
$ ./burry -o restore -s 1483383204
INFO[0000] Using existing burry manifest file /tmp/.burryfest  func=init
INFO[0000] Selected operation: RESTORE                   func=init
INFO[0000] My config: {InfraService:etcd Endpoint:10.0.1.139:1026 StorageTarget:local Creds:{StorageTargetEndpoint: Params:[]}}  func=init
INFO[0000] Backup restored in /tmp  func=unarch
INFO[0000] Restored /foo                                 func=visitETCDReverse
INFO[0000] Operation successfully completed. Restored 1 items from snapshot 1483383204  func=main
# ... and we're back to normal:
$ curl 10.0.1.139:1026/v2/keys/foo
{"action":"get","node":{"key":"/foo","value":"bar","modifiedIndex":17,"createdIndex":17}}
```

## Architecture

`burry` assumes that the infra service it operates on is tree-like. The essence of `burry`'s backup algorithm is:

- Walk the tree from the root
- For every non-leaf node: process its children
- For every leaf node, store the content (that is, the node value) 
- Depending on the storage target selected, create archive incl. metadata

The restore algorithm: TBD.