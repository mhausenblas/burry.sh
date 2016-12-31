/*
The burry command line tool enables to backup and restore
infra services such as ZooKeeper and etcd to and from local storage,
Amazon S3, Azure Storage, Google Storage, Minio as well as screen
dumps via the --target tty parameter.

Use:

 $ burry --endpoint IP:PORT (--isvc etcd|zk) (--target tty|local|s3) (--overwrite) (--credentials STORAGE_TARGET_ENDPOINT,KEY1=VAL1,KEY2=VAL2,...KEYn=VALn)

Examples:

 # dump DC/OS system ZooKeeper content to screen:
 $ burry --endpoint leader.mesos:2181

 # back up a DC/OS etcd service to Minio playground:
 $ burry --endpoint etcd.mesos:1026 --isvc etcd --target s3 --credentials play.minio.io:9000,AWS_ACCESS_KEY_ID=Q3AM3UQ867SPQQA43P2F,AWS_SECRET_ACCESS_KEY=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

To enable debug info, set an environment variable DEBUG, for example DEBUG=true burry ...

For a full guide visit https://github.com/mhausenblas/burry.sh
*/
package main
