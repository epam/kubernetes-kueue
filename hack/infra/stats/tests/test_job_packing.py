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

import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


STATS_DIR = Path(__file__).resolve().parents[1]


def load_module(name, filename):
    spec = importlib.util.spec_from_file_location(name, STATS_DIR / filename)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


class JobPackingTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.fetch = load_module("fetch_prow_metrics", "fetch_prow_metrics.py")
        cls.plot = load_module("stats_plot", "plot.py")

    def test_query_counts_fleet_pods_and_preserves_target_build_id(self):
        expr = self.fetch.job_packing_expr("kubernetes-sigs", "kueue", "pull-example-main")

        self.assertIn('max by (node,namespace,pod)(prow:job{phase="Running",node!=""} == 1)', expr)
        self.assertIn("count by (node)", expr)
        self.assertIn("* on(node) group_right", expr)
        self.assertIn("max by (node,id)", expr)
        self.assertIn('org="kubernetes-sigs",repo="kueue",name="pull-example-main"', expr)

    def test_free_resource_query_uses_allocatable_and_effective_pod_requests(self):
        for resource, unit in (("cpu", "core"), ("memory", "byte")):
            with self.subTest(resource=resource):
                expr = self.fetch.node_free_resource_expr(
                    "kubernetes-sigs", "kueue", "pull-example-main", resource, unit)

                selector = f'{{resource="{resource}",unit="{unit}"}}'
                self.assertIn(f'kube_node_status_allocatable{selector}', expr)
                self.assertIn(f'kube_pod_container_resource_requests{selector}', expr)
                self.assertIn(f'kube_pod_init_container_resource_requests{selector}', expr)
                self.assertIn(
                    'kube_pod_status_phase{phase=~"Pending|Running|Unknown"} == 1', expr)
                self.assertIn("max without (request_kind)", expr)
                self.assertIn("clamp_min", expr)
                self.assertIn("* on(node) group_right", expr)
                self.assertIn("max by (node,id)", expr)

    def test_plot_uses_each_builds_peak_node_occupancy(self):
        data = {
            "job": "pull-example-main",
            "start": 0,
            "end": 3600,
            "step": 30,
            "series": {
                "job_packing_pods": {
                    "build-1": [[0, 1], [30, 1]],
                    "build-2": [[0, 1], [30, 2], [60, 1]],
                    "build-3": [[0, 2], [30, 3], [60, 2]],
                }
            },
        }

        self.assertEqual([1, 2, 3], self.plot.per_build_job_packing(data).tolist())
        self.assertEqual(
            [1, 1, 1, 2, 1, 2, 3, 2],
            self.plot.job_packing_samples(data).tolist(),
        )

        with tempfile.TemporaryDirectory() as out:
            self.plot.plot_job_packing_distribution(data, out)
            image = Path(out) / "dist_job_packing.png"
            self.assertTrue(image.is_file())
            self.assertGreater(image.stat().st_size, 0)

    def test_percentiles_stay_within_observed_packing_range(self):
        values = [1] * 7 + [2]

        self.assertAlmostEqual(1.3, self.fetch.percentile(values, 90))
        self.assertAlmostEqual(1.65, self.fetch.percentile(values, 95))

    def test_free_resource_plot_uses_each_builds_minimum_cpu_and_memory(self):
        gib = 1024 ** 3
        data = {
            "job": "pull-example-main",
            "start": 0,
            "end": 3600,
            "step": 30,
            "series": {
                "node_free_cpu_cores": {
                    "build-1": [[0, 4], [30, 2]],
                    "build-2": [[0, 5], [30, 3]],
                    "build-3": [[0, 6], [30, 4]],
                },
                "node_free_memory_bytes": {
                    "build-1": [[0, 10 * gib], [30, 8 * gib]],
                    "build-2": [[0, 12 * gib], [30, 9 * gib]],
                    "build-3": [[0, 14 * gib], [30, 11 * gib]],
                },
            },
        }

        cpu, memory = self.plot.per_build_node_free_resources(data)
        self.assertEqual([2, 3, 4], cpu.tolist())
        self.assertEqual([8, 9, 11], memory.tolist())

        with tempfile.TemporaryDirectory() as out:
            legacy = Path(out) / "dist_node_leftover_cpu.png"
            legacy.write_bytes(b"obsolete")
            self.plot.plot_node_free_resource_distribution(data, out)
            image = Path(out) / "dist_node_free_resource.png"
            self.assertTrue(image.is_file())
            self.assertGreater(image.stat().st_size, 0)
            self.assertFalse(legacy.exists())


if __name__ == "__main__":
    unittest.main()
