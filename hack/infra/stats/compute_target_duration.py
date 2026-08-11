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
compute_target_duration.py — pick the shared CPU target duration for
target-duration-longest-job-improved-recursive-with-cutoff (see plot.py).

That recommender sizes every job's CPU against one GLOBAL wall-clock target (minutes)
instead of a fixed constant, so it must be computed once, up front, across every job in
the batch -- unlike the other recommenders, which only ever see one job's raw_series.json
and cannot see across the batch themselves.

Definition: for each fetched job, take the MEDIAN (p50) wall-clock duration of its own
builds (wall_durations, from prow:job phase transitions -- see fetch_prow_metrics.py's
fetch_prow_wall_durations); the target is the LARGEST of those per-job medians. That is
the typical runtime of whichever single job in the batch normally takes the longest, used
as the wall-clock budget every job's recursive CPU fit is then stretched/compressed
toward.

Medians are restricted to build ids that also appear in that job's cpu_used_cores series,
matching the usable-build set the recursive recommender itself scores (a build with wall
time but no CPU samples contributes nothing to the CPU fit either).

Usage:
  ./compute_target_duration.py ./ci-stats
  # then feed the printed value into plot.py for every job folder:
  ./plot.py <job-dir> \\
      --cpu-algorithm target-duration-longest-job-improved-recursive-with-cutoff \\
      --cpu-target-min <printed value>
"""
import argparse, glob, json, os
import statistics


def per_job_median_wall_minutes(raw):
    """Median wall-clock duration (minutes) of `raw`'s builds, restricted to ids also
    present in cpu_used_cores. Returns None if fewer than 2 such builds."""
    cpu_ids = set(raw["series"].get("cpu_used_cores", {}))
    vals = [v for bid, v in raw.get("wall_durations", {}).items() if bid in cpu_ids]
    if len(vals) < 2:
        return None
    return statistics.median(vals)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("workdir", help="work directory holding the per-job <job>_*/ folders "
                                     "(each with a raw_series.json from fetch_prow_metrics.py)")
    args = ap.parse_args()

    files = sorted(glob.glob(os.path.join(args.workdir, "*", "raw_series.json")))
    if not files:
        raise SystemExit(f"no raw_series.json under {args.workdir}/*/ — run fetch_prow_metrics.py first")

    rows = []
    for f in files:
        with open(f) as fh:
            raw = json.load(fh)
        median_min = per_job_median_wall_minutes(raw)
        rows.append((raw["job"], median_min, len(raw["series"].get("cpu_used_cores", {}))))

    rows.sort(key=lambda r: (r[1] is None, -(r[1] or 0)))
    print(f"{'job':<55} {'builds':>7} {'p50 wall (min)':>15}")
    for job, median_min, n in rows:
        print(f"{job:<55} {n:>7} {'?' if median_min is None else f'{median_min:>15.2f}'}")

    scored = [r for r in rows if r[1] is not None]
    if not scored:
        raise SystemExit("no job had >=2 usable builds; cannot compute a target duration")
    target_job, target_min, _ = scored[0]
    print(f"\nlongest-job target (max of per-job p50 wall duration): {target_min:.2f} min "
          f"(from {target_job})")
    print(f"\n--cpu-target-min {target_min:.2f}")


if __name__ == "__main__":
    main()
