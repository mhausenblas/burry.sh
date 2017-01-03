# Development and testing notes

## S3

Using Minio [Go Client SDK](https://docs.minio.io/docs/golang-client-quickstart-guide) for maximum coverage. Testing against [play.minio.io](https://play.minio.io:9000/) as the default S3 backend.

## Consul

Using HashiCorp's [Consul API client](https://github.com/hashicorp/consul/tree/master/api). Testing with following local environment:

```bash
$ docker run -d --name=dev-consul -p 8500:8500 consul:v0.6.4 agent -server -client=0.0.0.0 -node=node0 -bootstrap-expect=1

```

Once you have either a local environment or a remote Consul cluster up and running you can interact with it as follows (see also the [Consul K/V store API](https://www.consul.io/docs/agent/http/kv.html)):

```bash
# add keys:
$ curl -d @- localhost:8500/v1/kv/foo -XPUT <<< bar
$ curl -d @- localhost:8500/v1/kv/hi/ho -XPUT <<< test
# get value at key:
$ curl 127.0.0.1:8500/v1/kv/hi/ho?raw
# remove key:
$ curl localhost:8500/v1/kv/foo -XDELETE
```

## etcd

Using CoreOS' [etcd client](https://github.com/coreos/etcd/tree/master/client). Testing with following local environment:

```bash
$ docker run -d -p 2379:2379 -p 2380:2380 -p 4001:4001 -p 7001:7001 -v /data/backup/dir:/data --name test-etcd elcolio/etcd:2.0.10 -name test-etcd
```

Note: the IANA assigned ports for etcd are 2379 for client communication and 2380 for server-to-server communication.

Once you have either a local environment or a remote etcd cluster up and running you can interact with it as follows (see also the [etcd v2 API](https://coreos.com/etcd/docs/latest/v2/api.html)):

```bash
# add keys:
$ curl localhost:2379/v2/keys/foo -XPUT -d value="bar"
$ curl localhost:2379/v2/keys/baz -XPUT -d value="some"
$ curl localhost:2379/v2/keys/meh/hu -XPUT -d value="moar"
# get value at key:
$ curl localhost:2379/v2/keys/foo
{
  "action": "get",
  "node": {
    "key": "/foo",
    "value": "bar",
    "modifiedIndex": 8,
    "createdIndex": 8
  }
}
# list all top-level keys:
$ curl localhost:2379/v2/keys/
{
  "action": "get",
  "node": {
    "dir": true,
    "nodes": [
      {
        "key": "/foo",
        "value": "bar",
        "modifiedIndex": 3,
        "createdIndex": 3
      },
      {
        "key": "/baz",
        "value": "some",
        "modifiedIndex": 4,
        "createdIndex": 4
      },
      {
        "key": "/meh",
        "dir": true,
        "modifiedIndex": 5,
        "createdIndex": 5
      }
    ]
  }
}
# remove key:
$ curl localhost:2379/v2/keys/meh/hu -XDELETE
```

## ZooKeeper

For a local ZK test environment you can use [mbabineau/zookeeper-exhibitor](https://hub.docker.com/r/mbabineau/zookeeper-exhibitor/).
Note to replace `HOSTNAME=mh9` with the value for your host:

```bash
$ docker run -p 8181:8181 -p 2181:2181 -p 2888:2888 -p 3888:3888 -e HOSTNAME=mh9 mbabineau/zookeeper-exhibitor:latest
```

Once above container is running (make sure with `docker ps | grep mbabineau`), confirm Exhibitor is running:

```bash
$ http localhost:8181/exhibitor/v1/cluster/status
```

Also, check if ZK is playing along:

```bash
$ telnet localhost 2181
Trying ::1...
telnet: connect to address ::1: Connection refused
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
ruok
imokConnection closed by foreign host.
```

Note: Exhibitor's UI is at [localhost:8181/exhibitor/v1/ui/index.html](http://localhost:8181/exhibitor/v1/ui/index.html) available.

With the following command we can read out the config:

```bash
$ http localhost:8181/exhibitor/v1/config/get-state
...
{
    "backupActive": true,
    "config": {
        "autoManageInstances": 1,
        "autoManageInstancesApplyAllAtOnce": 1,
        "autoManageInstancesFixedEnsembleSize": 0,
        "autoManageInstancesSettlingPeriodMs": 0,
        "backupExtra": {
            "directory": ""
        },
        "backupMaxStoreMs": 21600000,
        "backupPeriodMs": 600000,
        "checkMs": 30000,
        "cleanupMaxFiles": 20,
        "cleanupPeriodMs": 300000,
        "clientPort": 2181,
        "connectPort": 2888,
        "controlPanel": {},
        "electionPort": 3888,
        "hostname": "mh9",
        "javaEnvironment": "",
        "log4jProperties": "",
        "logIndexDirectory": "/opt/zookeeper/transactions",
        "observerThreshold": 0,
        "rollInProgress": false,
        "rollPercentDone": 0,
        "rollStatus": "n/a",
        "serverId": 1,
        "serversSpec": "1:mh9",
        "zooCfgExtra": {
            "initLimit": "10",
            "quorumListenOnAllIPs": "true",
            "syncLimit": "5",
            "tickTime": "2000"
        },
        "zookeeperDataDirectory": "/opt/zookeeper/snapshots",
        "zookeeperInstallDirectory": "/opt/zookeeper",
        "zookeeperLogDirectory": "/opt/zookeeper/transactions"
    },
    "extraHeadingText": null,
    "nodeMutationsAllowed": true,
    "running": true,
    "standaloneMode": false,
    "version": "v1.5.5"
}
```