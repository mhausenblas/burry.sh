# burry

This is the `burry`, the Cloud Native Infrastructure BackUp & RecoveRY tool. Use `burry` to back up and restore
critical infrastructure metadata services such as ZooKeeper and etcd.


|to/from         |ZooKeeper    |etcd        |
| --------------:| ----------- | ---------- |
| Amazon S3      | backlog 1   | -          |
| Azure Storage  | -           | -          |
| Google Storage | -           | -          |
| Local          | WIP         | backlog 3  |
| Minio*         | backlog 2   | -          |

*) [Minio](https://www.minio.io/) either on-premises or in the cloud, self-hosted.

The essence of burry's algorithm is:

1. On startup, discover data directory location
1. Watch data directory location
1. Until end: on changes, but at latest every `AT_LEAST_SEC` zip and upload to target storage
