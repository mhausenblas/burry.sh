# burry

This is the `burry`, the Cloud Native Infrastructure BackUp & RecoveRY tool. Use `burry` to back up and restore
critical infrastructure services such as ZooKeeper and etcd.


|to/from         |ZooKeeper    |etcd        |
| --------------:| ----------- | ---------- |
| Amazon S3      | backlog 1   | -          |
| Azure Storage  | -           | -          |
| Google Storage | -           | -          |
| Local          | WIP         | backlog 3  |
| Minio*         | backlog 2   | -          |

*) [Minio](https://www.minio.io/) either on-premises or in the cloud, self-hosted.

The essence of burry's algorithm is:

- Until user cancels
  - Either on changes or every `AT_LEAST_SEC`
  - Walk the tree from root
  - Retrieve data and metadata from each non-ephemeral node
  - Write all data and metadata to storage target
