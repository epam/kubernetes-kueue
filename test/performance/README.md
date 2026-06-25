# test/performance/

Scheduler performance benchmarks for Kueue.

## Purpose

Measures scheduling throughput and latency under load — how many workloads per second the Kueue scheduler can admit at various scales (queue depths, ClusterQueue counts, PodSet sizes).

## Structure

```
test/performance/
└── scheduler/
    ├── checker/      — Validation: verifies that benchmark results meet SLO thresholds
    ├── runner/       — The benchmark runner itself
    │   ├── controller/   — Benchmark controller watching workload admission events
    │   ├── generator/    — Synthetic workload generator (creates workloads at a set rate)
    │   ├── recorder/     — Records timing samples to CSV
    │   ├── scraper/      — Scrapes Prometheus metrics during the run
    │   └── stats/        — Computes p50/p90/p99 latency statistics
    └── README.md
```

## Running

```bash
make test-performance
```

The benchmark runs a kind cluster with a controlled workload generator, records scheduling latency for each workload from creation to admission, and outputs a CSV report.

## Output Metrics

- **Throughput**: workloads admitted per second
- **Latency p50/p90/p99**: time from workload creation to admission
- **Queue depth over time**: pending workloads during the run
