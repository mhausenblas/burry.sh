# burry

This is the `burry`, the BackUp & RecoveRY tool for cloud native infrastructure services. Use `burry` to back up and restore
critical infrastructure base services such as ZooKeeper and etcd.

`burry` support back up/restore the following infrastructure services with the respective storage targets:

|to/from         |ZooKeeper    |etcd        |
| --------------:| ----------- | ---------- |
| Amazon S3      | backlog 1   | -          |
| Azure Storage  | -           | -          |
| Google Storage | -           | -          |
| Local          | WIP         | backlog 3  |
| Minio*         | backlog 2   | -          |

*) [Minio](https://www.minio.io/) either on-premises or in the cloud, self-hosted.

## Architecture

The essence of burry's algorithm is:

- Until user cancels
  - Either on changes or every `AT_LEAST_SEC`
  - Walk the tree from root
  - Retrieve data and metadata from each non-ephemeral node
  - Write all data and metadata to storage target

## Install

## Use