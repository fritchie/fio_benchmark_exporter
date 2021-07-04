package main

// Prometheus exporter for fio benchmarks

// By default a chosen benchmark job will run periodically with the results
// being exported in Prometheus format

// inspired by
// https://github.com/neoaggelos/fio-exporter

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var labels = []string{"benchmark"}

var (
	promRegistry = prometheus.NewRegistry()
	// START METRICS
	fioReadBW    = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_bandwidth_kbps",
			Help: "Read bandwidth (KiB/s)",
		},
		labels,
	)
	fioReadIOPS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_iops",
			Help: "Read IOPS",
		},
		labels,
	)
	fioReadLat90 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_pct90",
			Help: "Read total latency 90th percentile (usec)",
		},
		labels,
	)
	fioReadLat95 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_pct95",
			Help: "Read total latency 95th percentile (usec)",
		},
		labels,
	)
	fioReadLat99 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_pct99",
			Help: "Read total latency 99th percentile (usec)",
		},
		labels,
	)
	fioReadLatMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_min",
			Help: "Read total latency minimum (usec)",
		},
		labels,
	)
	fioReadLatMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_max",
			Help: "Read total latency maximum (usec)",
		},
		labels,
	)
	fioReadLatMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_lat_mean",
			Help: "Read total latency mean (usec)",
		},
		labels,
	)
	fioReadBWMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_bw_min_kb",
			Help: "Read bandwidth minimum (KiB/s)",
		},
		labels,
	)
	fioReadBWMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_bw_max_kb",
			Help: "Read bandwidth maximum (KiB/s)",
		},
		labels,
	)
	fioReadBWMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_bw_mean_kb",
			Help: "Read bandwidth mean (KiB/s)",
		},
		labels,
	)
	fioReadIOPSMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_iops_min",
			Help: "Read IOPS minimum",
		},
		labels,
	)
	fioReadIOPSMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_iops_max",
			Help: "Read IOPS maximum",
		},
		labels,
	)
	fioReadIOPSMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_read_iops_mean",
			Help: "Read IOPS mean",
		},
		labels,
	)
	fioWriteBW = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_bandwidth_kbps",
			Help: "Write bandwidth (KiB/s)",
		},
		labels,
	)
	fioWriteIOPS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_iops",
			Help: "Write IOPS",
		},
		labels,
	)
	fioWriteLat90 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_pct90",
			Help: "Write total latency 90th percentile (usec)",
		},
		labels,
	)
	fioWriteLat95 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_pct95",
			Help: "Write total latency 95th percentile (usec)",
		},
		labels,
	)
	fioWriteLat99 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_pct99",
			Help: "Write total latency 99th percentile (usec)",
		},
		labels,
	)
	fioWriteLatMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_min",
			Help: "Write total latency minimum (usec)",
		},
		labels,
	)
	fioWriteLatMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_max",
			Help: "Write total latency maximum (usec)",
		},
		labels,
	)
	fioWriteLatMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_lat_mean",
			Help: "Read total latency mean (usec)",
		},
		labels,
	)
	fioWriteBWMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_bw_min_kb",
			Help: "Write bandwidth minimum (KiB/s)",
		},
		labels,
	)
	fioWriteBWMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_bw_max_kb",
			Help: "Write bandwidth maximum (KiB/s)",
		},
		labels,
	)
	fioWriteBWMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_bw_mean_kb",
			Help: "Write bandwidth mean (KiB/s)",
		},
		labels,
	)
	fioWriteIOPSMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_iops_min",
			Help: "Write IOPS minimum",
		},
		labels,
	)
	fioWriteIOPSMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_iops_max",
			Help: "Write IOPS maximum",
		},
		labels,
	)
	fioWriteIOPSMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_write_iops_mean",
			Help: "Write IOPS mean",
		},
		labels,
	)
	fioCpuUser = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_cpu_user",
			Help: "User CPU utilization (%)",
		},
		labels,
	)
	fioCpuSys = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_cpu_sys",
			Help: "System CPU utilization (%)",
		},
		labels,
	)
	fioIODepth1 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_1",
			Help: "Queue depth <=1 (%)",
		},
		labels,
	)
	fioIODepth2 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_2",
			Help: "Queue depth 2 (%)",
		},
		labels,
	)
	fioIODepth4 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_4",
			Help: "Queue depth 4 (%)",
		},
		labels,
	)
	fioIODepth8 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_8",
			Help: "Queue depth 8 (%)",
		},
		labels,
	)
	fioIODepth16 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_16",
			Help: "Queue depth 16 (%)",
		},
		labels,
	)
	fioIODepth32 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_32",
			Help: "Queue depth 32 (%)",
		},
		labels,
	)
	fioIODepth64 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_iodepth_64",
			Help: "Queue depth 64+ (%)",
		},
		labels,
	)
	fioBenchmarkSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fio_benchmark_success",
			Help: "1 if last benchmark was successful, 0 otherwise",
		},
		labels,
	)
	// END METRICS
)

func init() {
	promRegistry.MustRegister(
		fioReadBW,
		fioReadIOPS,
		fioReadLat90,
		fioReadLat95,
		fioReadLat99,
		fioReadLatMin,
		fioReadLatMax,
		fioReadLatMean,
		fioReadBWMin,
		fioReadBWMax,
		fioReadBWMean,
		fioReadIOPSMin,
		fioReadIOPSMax,
		fioReadIOPSMean,
		fioWriteBW,
		fioWriteIOPS,
		fioWriteLat90,
		fioWriteLat95,
		fioWriteLat99,
		fioWriteLatMin,
		fioWriteLatMax,
		fioWriteLatMean,
		fioWriteBWMin,
		fioWriteBWMax,
		fioWriteBWMean,
		fioWriteIOPSMin,
		fioWriteIOPSMax,
		fioWriteIOPSMean,
		fioCpuUser,
		fioCpuSys,
		fioIODepth1,
		fioIODepth2,
		fioIODepth4,
		fioIODepth8,
		fioIODepth16,
		fioIODepth32,
		fioIODepth64,
		fioBenchmarkSuccess,
	)
}

func main() {
	// START FLAGS
	benchmark := flag.String("benchmark", "latency", "iops, latency or throughput")
	customFioBenchmarkFlags := flag.String("customFioBenchmarkFlags", "", "experts only")
	directory := flag.String("directory", "/tmp", "absolute path to directory to use for benchmark files")
	duration := flag.Duration("interval", 6 * time.Hour, "interval for consecutive benchmark runs")
	fileSize := flag.String("fileSize", "1G", "size of file to use for benchmark")
	port := flag.String("port", "9996", "tcp listen port")
	runOnce := flag.Bool("runOnce", false, "exit after benchmark complete and runOnceWait has expired")
	runOnceWait := flag.Duration("runOnceWait", 1 * time.Hour, "wait this duration before exiting a runOnce benchmark")
	benchmarkRuntime := flag.String("benchmarkRuntime", "60", "runtime for benchmark in seconds")
	statusUpdates := flag.Bool("statusUpdates", false, "update metrics every statusUpdateTime seconds during benchmark")
	statusUpdateInterval := flag.String("statusUpdateInterval", "30", "metric update interval in seconds when statusUpdates enabled")
	flag.Parse()
	// END FLAGS

	var fioBenchmarkFlags string
	switch *benchmark {
	case "iops":
		fioBenchmarkFlags = "--name=iops --numjobs=4 --ioengine=libaio --direct=1 --bs=4k --iodepth=128 --readwrite=randrw"
	case "latency":
		fioBenchmarkFlags = "--name=latency --numjobs=1 --ioengine=libaio --direct=1 --bs=4k --iodepth=1 --readwrite=randrw"
	case "throughput":
		fioBenchmarkFlags = "--name=throughput --numjobs=4 --ioengine=libaio --direct=1 --bs=128k --iodepth=64 --readwrite=rw"
	case "custom":
		break
	default:
		fioBenchmarkFlags = "--name=latency --numjobs=1 --ioengine=libaio --direct=1 --bs=4k --iodepth=1 --readwrite=randrw"
	}

	// make sure custom fio flags supplied for custom benchmark
	if *benchmark == "custom" && *customFioBenchmarkFlags == "" {
		log.Fatal("customFioBenchmarkFlags must be used when benchmark is custom")
	}

	// fio terse version 5 output used for all benchmarks
	// custom benchmarks cannot use the --output-format or --output flags
	if *benchmark == "custom" && strings.Contains(*customFioBenchmarkFlags, "output") {
		log.Fatal("customFioBenchmarkFlags cannot contain the flag --output-format or --output")
	}

	// make sure custom benchmark does not include any percentile related flags
	if *benchmark == "custom" && strings.Contains(*customFioBenchmarkFlags, "percentile") {
		log.Fatal("customFioBenchmarkFlags cannot contain any percentile related flags")
	}

	if !*runOnce {
		log.Printf("Configured interval: %v\n", *duration)
	}

	go func() {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		for {
			<-ch
			time.AfterFunc(*duration, func() { ch <- struct{}{} })

			var cmd string
			if *benchmark != "custom" {
				if !*statusUpdates {
					cmd = fmt.Sprintf("fio %s --directory=%s --size=%s --runtime=%s --time_based --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting", fioBenchmarkFlags, *directory, *fileSize, *benchmarkRuntime)
				} else {
					cmd = fmt.Sprintf("fio %s --status-interval=%s --directory=%s --size=%s --runtime=%s --time_based --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting", fioBenchmarkFlags, *statusUpdateInterval, *directory, *fileSize, *benchmarkRuntime)
				}
			} else {
				cmd = fmt.Sprintf("fio --output-format=terse --terse-version=5 --lat_percentiles=1 --clat_percentiles=0 --group_reporting %s", *customFioBenchmarkFlags)
			}

			log.Printf("Running fio: %s", cmd)
			cmdParts := strings.Split(cmd, " ")
			fioCommand := exec.Command(cmdParts[0], cmdParts[1:]...)
			fioStdout, err := fioCommand.StdoutPipe()
			if err != nil {
				log.Fatalf("Error creating StdoutPipe: %s", err)
			}
			fioStderr, err := fioCommand.StderrPipe()
			if err != nil {
				log.Fatalf("Error creating StderrPipe: %s", err)
			}
			var fioStderrBytes []byte
			go func() {
				fioStderrBytes, _ = io.ReadAll(fioStderr)
			}()
			if err := fioCommand.Start(); err != nil {
				log.Fatalf("Error starting fioCommand: %s", err)
			}
			scanner := bufio.NewScanner(fioStdout)
			// fio terse output format provides all stats on a single line
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				s := scanner.Text()
				// check if line matches fio terse v5 signature
				if s[0:6] != "5;fio-" {
					log.Printf("Line does not have the fio terse v5 signature, skipping: %s\n", s[0:6])
					continue
				}
				parts := strings.Split(s, ";")
				log.Printf("Fio update: %s\n", parts)
				fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(1)

				// START PARSE
				readBW, err := strconv.ParseFloat(parts[6], 64)
				if err != nil {
					log.Printf("Error parsing readBW (parts[6]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadBW.WithLabelValues(*benchmark).Set(readBW)
				}

				readIOPS, err := strconv.ParseFloat(parts[7], 64)
				if err != nil {
					log.Printf("Error parsing readIOPS (parts[7]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadIOPS.WithLabelValues(*benchmark).Set(readIOPS)
				}

				readLat90, err := strconv.ParseFloat(strings.Split(parts[27], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing readLat90 (parts[27]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLat90.WithLabelValues(*benchmark).Set(readLat90)
				}

				readLat95, err := strconv.ParseFloat(strings.Split(parts[28], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing readLat95 (parts[28]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLat95.WithLabelValues(*benchmark).Set(readLat95)
				}

				readLat99, err := strconv.ParseFloat(strings.Split(parts[29], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing readLat99 (parts[29]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLat99.WithLabelValues(*benchmark).Set(readLat99)
				}

				readLatMin, err := strconv.ParseFloat(parts[37], 64)
				if err != nil {
					log.Printf("Error parsing readLatMin (parts[37]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLatMin.WithLabelValues(*benchmark).Set(readLatMin)
				}

				readLatMax, err := strconv.ParseFloat(parts[38], 64)
				if err != nil {
					log.Printf("Error parsing readLatMax (parts[38]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLatMax.WithLabelValues(*benchmark).Set(readLatMax)
				}

				readLatMean, err := strconv.ParseFloat(parts[39], 64)
				if err != nil {
					log.Printf("Error parsing readLatMean (parts[39]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadLatMean.WithLabelValues(*benchmark).Set(readLatMean)
				}
				readBWMin, err := strconv.ParseFloat(parts[41], 64)
				if err != nil {
					log.Printf("Error parsing readBWMin (parts[41]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadBWMin.WithLabelValues(*benchmark).Set(readBWMin)
				}

				readBWMax, err := strconv.ParseFloat(parts[42], 64)
				if err != nil {
					log.Printf("Error parsing readBWMax (parts[42]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadBWMax.WithLabelValues(*benchmark).Set(readBWMax)
				}

				readBWMean, err := strconv.ParseFloat(parts[44], 64)
				if err != nil {
					log.Printf("Error parsing readBWMean (parts[44]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadBWMean.WithLabelValues(*benchmark).Set(readBWMean)
				}

				readIOPSMin, err := strconv.ParseFloat(parts[47], 64)
				if err != nil {
					log.Printf("Error parsing readIOPSMin (parts[47]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadIOPSMin.WithLabelValues(*benchmark).Set(readIOPSMin)
				}

				readIOPSMax, err := strconv.ParseFloat(parts[48], 64)
				if err != nil {
					log.Printf("Error parsing readIOPSMax (parts[48]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadIOPSMax.WithLabelValues(*benchmark).Set(readIOPSMax)
				}

				readIOPSMean, err := strconv.ParseFloat(parts[49], 64)
				if err != nil {
					log.Printf("Error parsing readIOPSMean (parts[49]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioReadIOPSMean.WithLabelValues(*benchmark).Set(readIOPSMean)
				}

				writeBW, err := strconv.ParseFloat(parts[53], 64)
				if err != nil {
					log.Printf("Error parsing writeBW (parts[53]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteBW.WithLabelValues(*benchmark).Set(writeBW)
				}

				writeIOPS, err := strconv.ParseFloat(parts[54], 64)
				if err != nil {
					log.Printf("Error parsing writeIOPS (parts[54]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteIOPS.WithLabelValues(*benchmark).Set(writeIOPS)
				}

				writeLat90, err := strconv.ParseFloat(strings.Split(parts[74], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing writeLat90 (parts[74]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLat90.WithLabelValues(*benchmark).Set(writeLat90)
				}

				writeLat95, err := strconv.ParseFloat(strings.Split(parts[75], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing writeLat95 (parts[75]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLat95.WithLabelValues(*benchmark).Set(writeLat95)
				}

				writeLat99, err := strconv.ParseFloat(strings.Split(parts[76], "=")[1], 64)
				if err != nil {
					log.Printf("Error parsing writeLat99 (parts[76]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLat99.WithLabelValues(*benchmark).Set(writeLat99)
				}

				writeLatMin, err := strconv.ParseFloat(parts[84], 64)
				if err != nil {
					log.Printf("Error parsing writeLatMin (parts[84]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLatMin.WithLabelValues(*benchmark).Set(writeLatMin)
				}

				writeLatMax, err := strconv.ParseFloat(parts[85], 64)
				if err != nil {
					log.Printf("Error parsing writeLatMax (parts[85]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLatMax.WithLabelValues(*benchmark).Set(writeLatMax)
				}

				writeLatMean, err := strconv.ParseFloat(parts[86], 64)
				if err != nil {
					log.Printf("Error parsing writeLatMean (parts[86]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteLatMean.WithLabelValues(*benchmark).Set(writeLatMean)
				}

				writeBWMin, err := strconv.ParseFloat(parts[88], 64)
				if err != nil {
					log.Printf("Error parsing writeBWMin (parts[88]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteBWMin.WithLabelValues(*benchmark).Set(writeBWMin)
				}

				writeBWMax, err := strconv.ParseFloat(parts[89], 64)
				if err != nil {
					log.Printf("Error parsing writeBWMax (parts[89]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteBWMax.WithLabelValues(*benchmark).Set(writeBWMax)
				}

				writeBWMean, err := strconv.ParseFloat(parts[91], 64)
				if err != nil {
					log.Printf("Error parsing writeBWMean (parts[91]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteBWMean.WithLabelValues(*benchmark).Set(writeBWMean)
				}
				writeIOPSMin, err := strconv.ParseFloat(parts[94], 64)
				if err != nil {
					log.Printf("Error parsing writeIOPSMin (parts[94]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteIOPSMin.WithLabelValues(*benchmark).Set(writeIOPSMin)
				}

				writeIOPSMax, err := strconv.ParseFloat(parts[95], 64)
				if err != nil {
					log.Printf("Error parsing writeIOPSMax (parts[95]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteIOPSMax.WithLabelValues(*benchmark).Set(writeIOPSMax)
				}

				writeIOPSMean, err := strconv.ParseFloat(parts[96], 64)
				if err != nil {
					log.Printf("Error parsing writeIOPSMean (parts[96]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioWriteIOPSMean.WithLabelValues(*benchmark).Set(writeIOPSMean)
				}

				cpuUser, err := strconv.ParseFloat(strings.Trim(parts[146], "%"), 64)
				if err != nil {
					log.Printf("Error parsing cpuUser (parts[146]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioCpuUser.WithLabelValues(*benchmark).Set(cpuUser)
				}

				cpuSys, err := strconv.ParseFloat(strings.Trim(parts[147], "%"), 64)
				if err != nil {
					log.Printf("Error parsing cpuSys (parts[147]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioCpuSys.WithLabelValues(*benchmark).Set(cpuSys)
				}

				ioDepth1, err := strconv.ParseFloat(strings.Trim(parts[151], "%"), 64)
				if err != nil {
					log.Printf("Error: parsing ioDepth1 (parts[151]) %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth1.WithLabelValues(*benchmark).Set(ioDepth1)
				}

				ioDepth2, err := strconv.ParseFloat(strings.Trim(parts[152], "%"), 64)
				if err != nil {
					log.Printf("Error: parsing ioDepth2 (parts[152]) %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth2.WithLabelValues(*benchmark).Set(ioDepth2)
				}

				ioDepth4, err := strconv.ParseFloat(strings.Trim(parts[153], "%"), 64)
				if err != nil {
					log.Printf("Error parsing ioDepth4 (parts[153]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth4.WithLabelValues(*benchmark).Set(ioDepth4)
				}

				ioDepth8, err := strconv.ParseFloat(strings.Trim(parts[154], "%"), 64)
				if err != nil {
					log.Printf("Error parsing ioDepth8 (parts[154]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth8.WithLabelValues(*benchmark).Set(ioDepth8)
				}

				ioDepth16, err := strconv.ParseFloat(strings.Trim(parts[155], "%"), 64)
				if err != nil {
					log.Printf("Error: parsing ioDepth16 (parts[155]) %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth16.WithLabelValues(*benchmark).Set(ioDepth16)
				}

				ioDepth32, err := strconv.ParseFloat(strings.Trim(parts[156], "%"), 64)
				if err != nil {
					log.Printf("Error parsing ioDepth32 (parts[156]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth32.WithLabelValues(*benchmark).Set(ioDepth32)
				}

				ioDepth64, err := strconv.ParseFloat(strings.Trim(parts[157], "%"), 64)
				if err != nil {
					log.Printf("Error parsing ioDepth64 (parts[157]): %s\n", err)
					fioBenchmarkSuccess.WithLabelValues(*benchmark).Set(0)
				} else {
					fioIODepth64.WithLabelValues(*benchmark).Set(ioDepth64)
				}
				// END PARSE
			}
			if err := fioCommand.Wait(); err != nil {
				log.Fatalf("Fio command error: %s\n%s\n", err, fioStderrBytes)
			}
			log.Println("Benchmark complete")
			if *runOnce {
				log.Printf("Waiting for runOnceWait of %s to expire", runOnceWait)
				time.Sleep(*runOnceWait)
				os.Exit(0)
			}
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(
		promRegistry,
		promhttp.HandlerOpts{},
	))

	log.Printf("Listening on :%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
