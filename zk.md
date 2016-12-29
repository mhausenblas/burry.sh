# ZooKeeper

The local test environment for ZooKeeper uses [mbabineau/zookeeper-exhibitor](https://hub.docker.com/r/mbabineau/zookeeper-exhibitor/).
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