# Background

## Use cases

- debugging
- simple standby
- x-cluster backups (DC1 ZK -> DC2 minio, DC1 minio -> DC2 ZK)

## Design considerations

### Design goals

Safe and usable. Only non-existing nodes or keys will be restored, that is, no existing data in ZK or etcd will be overwritten when attempting to restore data.

### Assumptions

`burry` assumes that the infra service it operates on is tree-like. 

### Backup algorithm

The essence of `burry`'s backup algorithm is:

- Walk the tree from the root
- For every non-leaf node: process its children
- For every leaf node, store the content (that is, the node value) 
- Depending on the storage target selected, create archive incl. metadata

### Restore algorithm

TBD.