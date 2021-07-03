# Fio Benchmark Exporter

Prometheus exporter for [fio](https://github.com/axboe/fio) benchmarks. 

By default a chosen benchmark job will run periodically with the results being exported in Prometheus format.

## Building and running

### Build

```
go build .
```

A sample Dockerfile, docker-compose.yaml, kustomization.yaml and kubernetes manifests are also provided.

### Running

Running the exporter requires fio and the libaio development packages to be installed on the host.

```
./fio_benchmark_exporter <flags>
```

For a kubernetes deployment edit kustomization.yaml as needed (you will probably need to change the storageClass in resources/pvc.yaml) and apply the resources:

```
kustomize build | kubectl apply -f -
```

#### Usage

```
./fio_benchmark_exporter -h
```

#### Flags

| Name | Description |
|-------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| benchmark                     | Name for a predefined set of fio job flags. Type: String. Default: latency. |
| customFioBenchmarkFlags       | Fio flags for a custom benchmark. Type: String. Experts Only. Fio can be destructive if used improperly. |
| directory                     | Absolute path to directory for fio benchmark files. Type: String. Default: /tmp. |
| fileSize                      | Size of file to use for fio benchmark. Fio --size flag. Type: String. Default: 1G. |
| interval                      | Time to wait in between consecutive benchmark runs. Type: Duration. Default: 6 hours. |
| port                          | Listen port number. Type: String. Default: 9996. |
| runOnce                       | Run benchmark once and exit. |
| runOnceWait                   | Wait this duration before exiting after runOnce benchmark completes. Type: Duration. Default: 1 hour. |
| benchmarkRuntime              | Benchmark runtime in seconds. Fio --runtime flag. Type: String. Default: 60. |
| statusUpdateInterval          | Seconds to wait in between metric updates when the statusUpdates flag is used. Fio --status-interval flag. Type: String. Default: 30. |
| statusUpdates                 | Update metrics periodically while benchmark is running. |

For Duration flag syntax see: [Golang Duration](https://golang.org/pkg/time/#ParseDuration)

#### Predefined Benchmarks

| Name             | Equivalent fio command when used with all defaults |
|------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| iops             | fio --name=iops --numjobs=4 --ioengine=libaio --direct=1 --bs=4k --iodepth=128 --readwrite=randrw --directory=/tmp --size=1G --runtime=60 --time_based --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting |
| latency          | fio --name=latency --numjobs=1 --ioengine=libaio --direct=1 --bs=4k --iodepth=1 --readwrite=randrw --directory=/tmp --size=1G --runtime=60 --time_based --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting |
| throughput       | fio --name=throughput --numjobs=4 --ioengine=libaio --direct=1 --bs=128k --iodepth=64 --readwrite=rw --directory=/tmp --size=1G --runtime=60 --time_based --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting |
| custom           | User defined. Experts only. Fio can be destructive if used improperly.|

#### Custom benchmark usage

For a custom benchmark supply all fio flags as a string.

```
./fio_benchmark_exporter -benchmark=custom -customFioBenchmarkFlags="--name=latency --status-interval=30 --numjobs=1 --ioengine=libaio --direct=1 --bs=4k --iodepth=1 --readwrite=randrw --directory=/tmp --size=1G --runtime=60 --time_based --lat_percentiles=1 --clat_percentiles=0 --group_reporting"
```

The flags

```
--output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting
```

will be used with custom benchmarks.

**Don't use the --output-format flag or any percentile related flags in customFioBenchmarkFlags. Additionally, don't specify a job file. Any flag that produces additional fio output may lead to metric parsing errors and incorrect reporting.**

## Collected Metrics

```
# HELP fio_benchmark_success 1 if last benchmark was successful, 0 otherwise
# TYPE fio_benchmark_success gauge
# HELP fio_cpu_sys System CPU utilization (%)
# TYPE fio_cpu_sys gauge
# HELP fio_cpu_user User CPU utilization (%)
# TYPE fio_cpu_user gauge
# HELP fio_iodepth_1 Queue depth 1 up to but not including 2 (%)
# TYPE fio_iodepth_1 gauge
# HELP fio_iodepth_16 Queue depth 16 up to but not including 32 (%)
# TYPE fio_iodepth_16 gauge
# HELP fio_iodepth_2 Queue depth 2 up to but not including 4 (%)
# TYPE fio_iodepth_2 gauge
# HELP fio_iodepth_32 Queue depth 32 up to but not including 64 (%)
# TYPE fio_iodepth_32 gauge
# HELP fio_iodepth_4 Queue depth 4 up to but not including 8 (%)
# TYPE fio_iodepth_4 gauge
# HELP fio_iodepth_64 Queue depth 64+ (%)
# TYPE fio_iodepth_64 gauge
# HELP fio_iodepth_8 Queue depth 8 up to but not including 16 (%)
# TYPE fio_iodepth_8 gaug
# HELP fio_read_bandwidth_kbps Read bandwidth (KiB/s)
# TYPE fio_read_bandwidth_kbps gauge
# HELP fio_read_bw_max_kb Read bandwidth maximum (KiB/s)
# TYPE fio_read_bw_max_kb gauge
# HELP fio_read_bw_mean_kb Read bandwidth mean (KiB/s)
# TYPE fio_read_bw_mean_kb gauge
# HELP fio_read_bw_min_kb Read bandwidth minimum (KiB/s)
# TYPE fio_read_bw_min_kb gauge
# HELP fio_read_iops Read IOPS
# TYPE fio_read_iops gauge
# HELP fio_read_iops_max Read IOPS maximum
# TYPE fio_read_iops_max gauge
# HELP fio_read_iops_mean Read IOPS mean
# TYPE fio_read_iops_mean gauge
# HELP fio_read_iops_min Read IOPS minimum
# TYPE fio_read_iops_min gauge
# HELP fio_read_lat_max Read total latency maximum (usec)
# TYPE fio_read_lat_max gauge
# HELP fio_read_lat_mean Read total latency mean (usec)
# TYPE fio_read_lat_mean gauge
# HELP fio_read_lat_min Read total latency minimum (usec)
# TYPE fio_read_lat_min gauge
# HELP fio_read_lat_pct90 Read total latency 90th percentile (usec)
# TYPE fio_read_lat_pct90 gauge
# HELP fio_read_lat_pct95 Read total latency 95th percentile (usec)
# TYPE fio_read_lat_pct95 gauge
# HELP fio_read_lat_pct99 Read total latency 99th percentile (usec)
# TYPE fio_read_lat_pct99 gauge
# HELP fio_write_bandwidth_kbps Write bandwidth (KiB/s)
# TYPE fio_write_bandwidth_kbps gauge
# HELP fio_write_bw_max_kb Write bandwidth maximum (KiB/s)
# TYPE fio_write_bw_max_kb gauge
# HELP fio_write_bw_mean_kb Write bandwidth mean (KiB/s)
# TYPE fio_write_bw_mean_kb gauge
# HELP fio_write_bw_min_kb Write bandwidth minimum (KiB/s)
# TYPE fio_write_bw_min_kb gauge
# HELP fio_write_iops Write IOPS
# TYPE fio_write_iops gauge
# HELP fio_write_iops_max Write IOPS maximum
# TYPE fio_write_iops_max gauge
# HELP fio_write_iops_mean Write IOPS mean
# TYPE fio_write_iops_mean gauge
# HELP fio_write_iops_min Write IOPS minimum
# TYPE fio_write_iops_min gauge
# HELP fio_write_lat_max Write total latency maximum (usec)
# TYPE fio_write_lat_max gauge
# HELP fio_write_lat_mean Read total latency mean (usec)
# TYPE fio_write_lat_mean gauge
# HELP fio_write_lat_min Write total latency minimum (usec)
# TYPE fio_write_lat_min gauge
# HELP fio_write_lat_pct90 Write total latency 90th percentile (usec)
# TYPE fio_write_lat_pct90 gauge
# HELP fio_write_lat_pct95 Write total latency 95th percentile (usec)
# TYPE fio_write_lat_pct95 gauge
# HELP fio_write_lat_pct99 Write total latency 99th percentile (usec)
# TYPE fio_write_lat_pct99 gauge
```
