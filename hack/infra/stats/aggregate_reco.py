#!/usr/bin/env python3

# Copyright The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
aggregate_reco.py — combine every <workdir>/*/recommendation.json (written by
plot.py) into a single recommendation.md table in the work directory.

The table shows, per job, the current vs recommended CPU/memory request and limit,
the CPU cores saved by the recommendation, and a TOTAL row summing the savings.

Example:
  ./aggregate_reco.py ./artifacts/infra-stats
"""
import argparse, glob, json, os

# A build node has ~7 usable cores (see hack/infra/stats/plot.py's CPU_MAX_CORES). Used both
# to snap a near-ceiling recommendation up to the full node and to bin-pack the CPU-only
# table's request list into node counts.
NODE_CAPACITY_CORES = 7.0
# Once a recommendation already claims more than half a node, round it up to the full node:
# above this point 2-per-node packing is lost either way (see plot.py's
# snap_to_node_packing), so the remaining slice is free headroom rather than a real saving.
NODE_ROUND_UP_THRESHOLD_CORES = 3.5


def num(v):
    return "?" if v is None else f"{v:g}"


def snap_near_ceiling(rec, capacity=NODE_CAPACITY_CORES, threshold=NODE_ROUND_UP_THRESHOLD_CORES):
    """Round rec up to the full node once it already exceeds `threshold` cores (default 3.5,
    half of a 7-core node). Below that, rec passes through unchanged."""
    if rec is None:
        return None
    return capacity if rec > threshold else rec


def cutoff_threshold_cores(args, capacity=NODE_CAPACITY_CORES):
    """Resolve --cutoff-frac / --cutoff-cores (mutually exclusive) to an absolute core
    threshold for snap_near_ceiling. Neither passed -> NODE_ROUND_UP_THRESHOLD_CORES."""
    if args.cutoff_frac is not None:
        return args.cutoff_frac * capacity
    if args.cutoff_cores is not None:
        return args.cutoff_cores
    return NODE_ROUND_UP_THRESHOLD_CORES


def first_fit_decreasing_bins(values, capacity=NODE_CAPACITY_CORES):
    """Number of `capacity`-sized bins (build nodes) needed to pack every value, via
    first-fit-decreasing: largest requests placed first, each into the first bin with enough
    remaining room, opening a new bin only when none fits. FFD is a simple, standard
    approximation (not necessarily optimal) for bin packing, good enough for an estimate of
    how many nodes a set of CPU requests would occupy if densely packed."""
    remaining = []
    for v in sorted((v for v in values if v is not None), reverse=True):
        for i, room in enumerate(remaining):
            if v <= room + 1e-9:
                remaining[i] -= v
                break
        else:
            remaining.append(capacity - v)
    return len(remaining)


def target_duration_line(rows, target_mins):
    """Format the 'Target duration: ...' line. plot.py's
    target-duration-longest-job-improved-recursive-with-cutoff auto-picks its target as the
    largest per-job p50 wall-clock build duration across the batch (see
    discover_longest_job_target_min) -- so that value IS some job's own p50 by construction.
    Find and name that job by matching the target back against every row's p50_duration_min
    (present regardless of which --cpu-algorithm produced the row), so a reader sees WHY this
    number was picked rather than a bare minute count. A target with no matching job (e.g. a
    manually-chosen --cpu-target-min that isn't any job's p50) is shown as a bare number."""
    if not target_mins:
        return "Target duration: ? min"
    parts = []
    for t in sorted(target_mins):
        source = next((r["job"] for r in rows
                       if r.get("p50_duration_min") is not None
                       and abs(r["p50_duration_min"] - t) < 0.01), None)
        if source:
            parts.append(f"{t:g} min ({source}, p50 duration)")
        else:
            parts.append(f"{t:g} min")
    return "Target duration: " + ", ".join(parts)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("workdir", help="work directory holding the per-job <job>_*/ folders")
    ap.add_argument("--out", default=None, help="output path (default <workdir>/recommendation.md)")
    cutoff = ap.add_mutually_exclusive_group()
    cutoff.add_argument("--cutoff-frac", type=float, default=None,
                        help="percentage-based node round-up cutoff: once rec cpu req exceeds "
                             "this fraction of a node (e.g. 0.85 = 85%%), round it up to the "
                             "full node. Mutually exclusive with --cutoff-cores.")
    cutoff.add_argument("--cutoff-cores", type=float, default=None,
                        help="absolute node round-up cutoff in cores (e.g. 3.5 = half a "
                             f"{NODE_CAPACITY_CORES:g}-core node): once rec cpu req exceeds "
                             "this many cores, round it up to the full node. Mutually "
                             "exclusive with --cutoff-frac. Default when neither is passed: "
                             f"{NODE_ROUND_UP_THRESHOLD_CORES:g} cores.")
    args = ap.parse_args()
    threshold = cutoff_threshold_cores(args)

    files = sorted(glob.glob(os.path.join(args.workdir, "*", "recommendation.json")))
    rows = []
    for f in files:
        with open(f) as fh:
            rows.append(json.load(fh))
    if not rows:
        raise SystemExit(f"no recommendation.json under {args.workdir}/*/ — run the plot scripts first")

    header = ("| job | builds | avg dur (min) | p50 dur (min) | p95 dur (min) | cur cpu req | cur cpu lim "
              "| cur mem req | cur mem lim | rec cpu req | rec cpu lim | rec mem req | rec mem lim |")
    sep = "|---|--:|--:|--:|--:|--:|--:|--:|--:|--:|--:|--:|--:|"
    lines = [f"# CPU/memory right-sizing recommendations ({len(rows)} jobs)",
             "", "CPU values are cores; memory values are GiB.", "",
             header, sep]

    # accumulate potential savings per metric: sum(current - recommended)
    saved = {"cpu_req": 0.0, "cpu_lim": 0.0, "mem_req": 0.0, "mem_lim": 0.0}

    # the CPU target build duration (minutes) all jobs were sized against
    target_mins = {r["cpu"].get("stats", {}).get("target_min") for r in rows}
    target_mins.discard(None)

    def add(key, cur, rec):
        if cur is not None and rec is not None:
            saved[key] += cur - rec

    for r in rows:
        c, m = r["cpu"], r["mem"]
        add("cpu_req", c["request_current"], c["request_recommended"])
        add("cpu_lim", c["limit_current"], c["limit_recommended"])
        add("mem_req", m["request_current"], m["request_recommended"])
        add("mem_lim", m["limit_current"], m["limit_recommended"])
        lines.append(
            f"| {r['job']} | {r['builds']} | {num(r.get('avg_duration_min'))} "
            f"| {num(r.get('p50_duration_min'))} | {num(r.get('p95_duration_min'))} "
            f"| {num(c['request_current'])} | {num(c['limit_current'])} "
            f"| {num(m['request_current'])} | {num(m['limit_current'])} "
            f"| {num(c['request_recommended'])} | {num(c['limit_recommended'])} "
            f"| {num(m['request_recommended'])} | {num(m['limit_recommended'])} |")

    lines += [
        "",
        f"Potentially saved cpu req: {saved['cpu_req']:g} cores",
        f"Potentially saved cpu lim: {saved['cpu_lim']:g} cores",
        f"Potentially saved mem req: {saved['mem_req']:g} GiB",
        f"Potentially saved mem lim: {saved['mem_lim']:g} GiB",
        "",
        target_duration_line(rows, target_mins),
    ]

    out = args.out or os.path.join(args.workdir, "recommendation.md")
    with open(out, "w") as f:
        f.write("\n".join(lines) + "\n")
    print(f"{len(rows)} jobs -> {out}")
    for k, unit in (("cpu_req", "cores"), ("cpu_lim", "cores"), ("mem_req", "GiB"), ("mem_lim", "GiB")):
        print(f"  saved {k}: {saved[k]:g} {unit}")

    # CPU-only view: request == limit is enforced (Guaranteed QoS), so the limit columns are
    # redundant here -- just request, and how much it changes. rec is snapped up to the full
    # node once it already exceeds `threshold` cores (see snap_near_ceiling /
    # cutoff_threshold_cores; --cutoff-frac or --cutoff-cores picks how that's expressed).
    cutoff_desc = (f"{args.cutoff_frac:.0%} of a {NODE_CAPACITY_CORES:g}-core node "
                   f"({threshold:g} cores)" if args.cutoff_frac is not None
                   else f"{threshold:g} cores")
    cpu_header = "| job | builds | p50 dur (min) | p95 dur (min) | now cpu req | rec cpu req | diff |"
    cpu_sep = "|---|--:|--:|--:|--:|--:|--:|"
    cpu_lines = [f"# CPU right-sizing recommendations ({len(rows)} jobs)",
                 "", "CPU values are cores. request == limit (Guaranteed QoS), so only the "
                 "request is shown. rec cpu req is rounded up to a full "
                 f"{NODE_CAPACITY_CORES:g}-core node once it already exceeds {cutoff_desc}.",
                 "", cpu_header, cpu_sep]
    cpu_saved = 0.0
    now_vals, rec_vals = [], []
    for r in rows:
        c = r["cpu"]
        cur, rec = c["request_current"], snap_near_ceiling(c["request_recommended"], threshold=threshold)
        diff = cur - rec if cur is not None and rec is not None else None
        if diff is not None:
            cpu_saved += diff
        if cur is not None:
            now_vals.append(cur)
        if rec is not None:
            rec_vals.append(rec)
        cpu_lines.append(
            f"| {r['job']} | {r['builds']} "
            f"| {num(r.get('p50_duration_min'))} | {num(r.get('p95_duration_min'))} "
            f"| {num(cur)} | {num(rec)} | {num(diff)} |")

    now_bins, rec_bins = first_fit_decreasing_bins(now_vals), first_fit_decreasing_bins(rec_vals)
    now_eff = 100 * sum(now_vals) / (now_bins * NODE_CAPACITY_CORES) if now_bins else 0.0
    rec_eff = 100 * sum(rec_vals) / (rec_bins * NODE_CAPACITY_CORES) if rec_bins else 0.0
    cpu_lines += [
        "", f"Potentially saved cpu req: {cpu_saved:g} cores", "",
        target_duration_line(rows, target_mins),
        "",
        f"Binpackability ({NODE_CAPACITY_CORES:g}-core nodes, first-fit-decreasing):",
        f"- now: {now_bins} nodes ({now_eff:.1f}% packed)",
        f"- recommended: {rec_bins} nodes ({rec_eff:.1f}% packed)",
    ]
    cpu_out = os.path.join(os.path.dirname(out) or ".", "recommendation_cpu.md")
    with open(cpu_out, "w") as f:
        f.write("\n".join(cpu_lines) + "\n")
    print(f"{len(rows)} jobs -> {cpu_out}")
    print(f"  binpack: now {now_bins} nodes ({now_eff:.1f}%), rec {rec_bins} nodes ({rec_eff:.1f}%)")


if __name__ == "__main__":
    main()
