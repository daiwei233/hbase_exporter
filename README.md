# HBase Exporter

Prometheus exporter for HBase which fetch data from hbase jmx, written in Go.

You can even see region-level metrics.



## Installation and Usage

For pre-built binaries please take a look at the releases.



### Configuration

Below is the command line options summary:

`hbase_exporter --help`

| Argument               | Introduced in Version | Description                                           | Default                    |
| ---------------------- | --------------------- | ----------------------------------------------------- | -------------------------- |
| web.listen-address     | 1.2.0-cdh5.12.1       | Address to listen on for web interface and telemetry. | :9114                      |
| web.telemetry-path     | 1.2.0-cdh5.12.1       | Path under which to expose metrics.                   | /metrics                   |
| hbase.master.uri       | 1.2.0-cdh5.12.1       | HTTP jmx address of an HBase master node.             | http://localhost:60010/jmx |
| hbase.regionserver.uri | 1.2.0-cdh5.12.1       | HTTP jmx address of an HBase regionserver node.       | http://localhost:60030/jmx |
| hbase.master           | 1.2.0-cdh5.12.1       | Is hbase master.                                      | false                      |



#### Master

Start in master:

```
./hbase_exporter --web.listen-address=":9003" --hbase.master.uri="http://localhost:60010/jmx" --hbase.master
```

#### Regionserver

Start in regionserver:

```
./hbase_exporter --web.listen-address=":9003" --hbase.regionserver.uri="http://localhost:60010/jmx"
```



### Metrics

#### common

> Common jvm metrics, both hmaster and regionservers.

> From(both hmaster and regionservers):
>
>  http://localhost:60030/jmx?qry=Hadoop:service=HBase,name=JvmMetrics and http://localhost:60010/jmx?qry=Hadoop:service=HBase,name=JvmMetrics

| Name                      | Type  | Origin in jmx   |
| ------------------------- | ----- | --------------- |
| hbase_mem_non_head_used_m | gauge | MemNonHeapUsedM |
| hbase_mem_heap_userd_m    | gauge | MemHeapUsedM    |
| hbase_heap_max_m          | gauge | MemHeapMaxM     |
| hbase_mem_max_m           | gauge | MemMaxM         |
| hbase_gc_time_millis      | gauge | GcTimeMillis    |
| hbaes_gc_count            | gauge | GcCount         |
| hbase_thread_blocked      | gauge | ThreadsBlocked  |



#### HMaster

> HMaster server metrics, only for hmaster.
>
> From: http://localhost:60030/jmx?qry=Hadoop:service=HBase,name=Master,sub=Server

| Name                          | Type  | Origin in jmx        |
| ----------------------------- | ----- | -------------------- |
| hbase_num_region_servers      | gauge | NumRegionServers     |
| hbase_num_dead_region_servers | gauge | NumDeadRegionServers |
| hbase_is_active_master        | gauge | IsActiveMaster       |
| hbase_average_load            | gauge | AverageLoad          |



#### Regionserver

>Regionserver server metrics, only for regionserver.
>
>From: http://localhost:60030/jmx?qry=qry=Hadoop:service=HBase,name=RegionServer,sub=Server

| Name                          | Type  | Origin in jmx         |
| ----------------------------- | ----- | --------------------- |
| hbase_mem_store_size          | gauge | MemStoreSize          |
| hbase_region_count            | gauge | RegionCount           |
| hbase_store_count             | gauge | StoreCount            |
| hbase_store_file_count        | gauge | StoreFileCount        |
| hbase_store_file_size         | gauge | StoreFileSize         |
| hbase_total_request_count     | gauge | TotalRequestCount     |
| hbase_split_queue_length      | gauge | SplitQueueLength      |
| hbase_compaction_queue_length | gauge | CompactionQueueLength |
| hbase_flush_queue_length      | gauge | FlushQueueLength      |
| hbase_block_count_hit_percent | gauge | BlockCountHitPercent  |
| hbase_slow_append_count       | gauge | SlowAppendCount       |
| hbase_slow_delete_count       | gauge | SlowDeleteCount       |
| hbase_slow_get_count          | gauge | SlowGetCount          |
| hbase_slow_put_count          | gauge | SlowPutCount          |
| hbase_slow_increment_count    | gauge | SlowIncrementCount    |



> Regionserver region metrics, only for regionserver.
>
> From: http://localhost:60030/jmx?qry=Hadoop:service=HBase,name=RegionServer,sub=Regions
>
> Example:  hbase_store_count{host="localhost",hregion="4fcaf7b9d1fedc1b62c15cbb1c9a10dc",htable="t1",namespace="n1",role="ddn013018.heracles.sohuno.com"} 1

| Name                        | Type  | Origin in jmx             |
| --------------------------- | ----- | ------------------------- |
| hbase_store_count           | gauge | storeCount                |
| hbase_store_file_count      | gauge | storeFileCount            |
| hbase_mem_store_size        | gauge | memStoreSize              |
| hbase_store_file_size       | gauge | storeFileSize             |
| compactions_completed_count | gauge | compactionsCompletedCount |
| read_request_count          | gauge | readRequestCount          |
| write_request_count         | gauge | writeRequestCount         |
| num_files_compacted_count   | gauge | numFilesCompactedCount    |
| num_bytes_compacted_count   | gauge | numBytesCompactedCount    |