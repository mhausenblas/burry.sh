/*
The burry command line tool enables to backup and restore
infra services such as ZooKeeper and etcd to and from local storage,
Amazon S3, Azure Storage, Google Storage, Minio as well as screen
dumps via the --target tty parameter.

Use:

 $ burry --endpoint IP:PORT (--isvc etcd|zk) (--target tty|local|s3) (--overwrite)

Examples:

 $ burry --endpoint leader.mesos:2181
 $ burry --endpoint etcd.mesos:1026 --isvc etcd --target s3

To enable debug info, set an environment variable DEBUG.

For a full guide visit https://github.com/mhausenblas/burry.sh
*/
package main
