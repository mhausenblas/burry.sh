/*
The burry command line tool enables to backup and restore
infra services such as ZooKeeper and etcd to and from local storage,
Amazon S3, Azure Storage, Google Storage, Minio as well as screen
dumps via the --target tty parameter.

Use:

 $ burry [args]

 With [args]:

  -b, --burryfest
        Create a burry manifest file .burryfest in the current directory.
        The manifest file captures the current command line parameters for re-use in subsequent operations.
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
  -s, --snapshot string
        The ID of the snapshot.
        Example: 1483193387
  -t, --target string
        The storage target to use.
        Supported values are [local minio s3 tty] (default "tty")
  -v, --version
        Display version information and exit.

Note that for a backup operation (-o backup) the endpoint (-e EEE) is mandatory and
for a restore operation (-o restore) in addition the storage target (-t TTT) and the
snapshot ID (-s SSS) are mandatory.

Examples:

 # dump DC/OS system ZooKeeper content to screen:
 $ burry --endpoint leader.mesos:2181

 # back up a DC/OS etcd service to Minio playground:
 $ burry --endpoint etcd.mesos:1026 --isvc etcd --target s3 --credentials play.minio.io:9000,AWS_ACCESS_KEY_ID=Q3AM3UQ867SPQQA43P2F,AWS_SECRET_ACCESS_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

 # restore etcd from snapshot ID 1483383204
 $ burry -o restore -e etcd.mesos:1026 -i etcd -t local -s 1483383204

To enable debug info, set an environment variable DEBUG, for example DEBUG=true burry ...

For a full guide visit https://github.com/mhausenblas/burry.sh
*/
package main
