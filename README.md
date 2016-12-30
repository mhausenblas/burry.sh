# burry

This is the `burry`, the BackUp & RecoveRY tool for cloud native infrastructure services. Use `burry` to back up and restore
critical infrastructure base services such as ZooKeeper and etcd.

`burry` support back up/restore the following infrastructure services with the respective storage targets:

|to/from         |ZooKeeper   |etcd        |
| --------------:| ---------- | ---------- |
| Amazon S3      | WIP        | backlog    |
| Azure Storage  | backlog    | backlog    |
| Google Storage | backlog    | backlog    |
| Local          | yes        | WIP        |
| Minio*         | backlog    | backlog    |
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
1. If a manifest `.burryfest` exists in the current directory it will be used.
1. For every storage target other than `tty` a new manifest in the timestamped ZIP file will be created.