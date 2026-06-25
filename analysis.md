# Kueue — Complete Project Analysis

> Generated: 2026-06-23. Covers the full codebase at `sigs.k8s.io/kueue` (release v0.18.1).

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Directory Structure](#2-directory-structure)
3. [Go Module & Dependencies](#3-go-module--dependencies)
4. [Build System](#4-build-system)
5. [CI & GitHub Workflows](#5-ci--github-workflows)
6. [Helm Chart](#6-helm-chart)
7. [Kustomize / Deploy Config](#7-kustomize--deploy-config)
8. [Binaries & Entry Points](#8-binaries--entry-points)
9. [API Types & CRDs](#9-api-types--crds)
   - 9.1 Workload
   - 9.2 ClusterQueue
   - 9.3 LocalQueue
   - 9.4 ResourceFlavor
   - 9.5 AdmissionCheck
   - 9.6 Cohort
   - 9.7 WorkloadPriorityClass
   - 9.8 Topology (TAS)
   - 9.9 MultiKueue types
   - 9.10 ProvisioningRequestConfig
   - 9.11 Configuration types
10. [Feature Gates](#10-feature-gates)
11. [Scheduler & Admission System](#11-scheduler--admission-system)
    - 11.1 Scheduler loop
    - 11.2 Nomination phase
    - 11.3 Flavor assigner
    - 11.4 Preemption
    - 11.5 Fair sharing (DRS)
    - 11.6 Admission end-to-end flow
12. [In-Memory Cache & Queue](#12-in-memory-cache--queue)
    - 12.1 Scheduler cache
    - 12.2 Queue manager
    - 12.3 Snapshot mechanism
    - 12.4 Hierarchy manager
13. [Controllers & Reconcilers](#13-controllers--reconcilers)
    - 13.1 Core controllers
    - 13.2 Job framework & adapters
    - 13.3 Admission check controllers
    - 13.4 Specialized controllers
    - 13.5 Controller startup order
14. [Job Integrations](#14-job-integrations)
15. [MultiKueue](#15-multikueue)
    - 15.1 Architecture
    - 15.2 clustersReconciler
    - 15.3 wlReconciler
    - 15.4 MultiKueueAdapter interface
    - 15.5 Workload flow end-to-end
    - 15.6 Failure & recovery
16. [Topology-Aware Scheduling (TAS)](#16-topology-aware-scheduling-tas)
17. [Concurrent Admission](#17-concurrent-admission)
18. [Webhooks](#18-webhooks)
19. [Visibility API](#19-visibility-api)
20. [Metrics](#20-metrics)
21. [Test Infrastructure](#21-test-infrastructure)
    - 21.1 Test structure
    - 21.2 Unit tests
    - 21.3 Integration tests
    - 21.4 E2E tests
    - 21.5 Performance tests
22. [Code Generation](#22-code-generation)
23. [Linting & Verification](#23-linting--verification)
24. [Release Process](#24-release-process)
25. [Key Labels, Annotations & Constants](#25-key-labels-annotations--constants)
26. [RBAC Summary](#26-rbac-summary)
27. [File Path Quick Reference](#27-file-path-quick-reference)

---

## 1. Project Overview

Kueue is a Kubernetes-native job queueing system. It manages workload admission, queuing, and preemption for batch and ML workloads across ClusterQueues and Cohorts.

| Property | Value |
|---|---|
| Module | `sigs.k8s.io/kueue` |
| Current release | v0.18.1 |
| Stable API | v1beta2 |
| Min Kubernetes | 1.29 |
| Go version | 1.26 |
| License | Apache 2.0 |

**Core capabilities:**
- Priority-based queueing (StrictFIFO, BestEffortFIFO)
- Resource quota with cohort borrowing/lending and fair sharing
- Preemption (within ClusterQueue and within Cohort)
- Topology-Aware Scheduling (TAS)
- MultiKueue federation (multi-cluster dispatch)
- AdmissionChecks (provisioning, custom)
- All-or-nothing scheduling, partial admission, dynamic reclaim
- Elastic jobs via WorkloadSlices
- Concurrent admission
- 18+ job framework integrations

---

## 2. Directory Structure

```
kueue/
├── apis/                     # CRD and config API definitions
│   ├── config/v1beta1/       # Configuration types (older)
│   ├── config/v1beta2/       # Configuration types (current)
│   ├── kueue/v1alpha1/       # Alpha API types
│   ├── kueue/v1beta1/        # v1beta1 API types (deprecated, conversion target)
│   └── visibility/v1beta1/   # Visibility API types
├── bin/                      # Compiled binaries (gitignored)
├── charts/kueue/             # Helm chart
├── cmd/
│   ├── kueue/                # Main controller manager binary
│   ├── kueuectl/             # CLI tool (kubectl plugin)
│   ├── importer/             # Workload importer utility
│   ├── kueueviz/             # KueueViz dashboard (Go backend + Node.js frontend)
│   └── experimental/
│       ├── kueue-populator/  # Synthetic workload generator
│       ├── kueue-priority-booster/ # Priority boost example controller
│       └── skills/           # Agent skill runbooks (SKILL.md files)
├── config/
│   ├── components/           # Individual component manifests
│   │   ├── crd/              # 10 CRD manifests
│   │   ├── manager/          # Deployment, config
│   │   ├── rbac/             # Roles, bindings, service accounts
│   │   ├── webhook/          # Webhook service & config
│   │   ├── certmanager/      # Cert-Manager integration
│   │   ├── kueueviz/         # Dashboard manifests
│   │   ├── prometheus/       # ServiceMonitor, PrometheusRule
│   │   ├── visibility/       # APIService manifests
│   │   └── internalcert/     # Internal TLS secret
│   ├── default/              # Default Kustomize overlay
│   ├── dev/                  # Development overlay
│   ├── alpha-enabled/        # Alpha feature gates enabled
│   ├── prometheus/           # Prometheus monitoring setup
│   └── kueueviz/             # KueueViz overlay
├── dep-crds/                 # Dependent CRDs (JobSet, Kubeflow, Ray, etc.)
├── docs/                     # Documentation
├── hack/
│   ├── testing/              # E2E and integration test scripts
│   ├── releasing/            # Release automation scripts
│   ├── tools/                # Code generation tools
│   └── *.sh                  # Various utility scripts
├── internal/mocks/           # Generated mocks (mockgen)
├── keps/                     # Kubernetes Enhancement Proposals
├── pkg/
│   ├── cache/
│   │   ├── hierarchy/        # Cohort/ClusterQueue hierarchy manager
│   │   ├── queue/            # Queue cache (pending workloads)
│   │   └── scheduler/        # Scheduler cache (admitted workloads, snapshots)
│   ├── controller/
│   │   ├── admissionchecks/  # AdmissionCheck controllers (multikueue, provisioning)
│   │   ├── concurrentadmission/ # Concurrent admission controller
│   │   ├── core/             # Core controllers (workload, CQ, LQ, cohort, etc.)
│   │   ├── elasticjobs/      # Elastic job ungating
│   │   ├── failurerecovery/  # Pod termination recovery
│   │   ├── jobframework/     # Job interface, integration manager
│   │   ├── jobs/             # Per-framework job adapters
│   │   ├── tas/              # Topology-Aware Scheduling controllers
│   │   └── workloaddispatcher/ # MultiKueue workload dispatch
│   ├── features/             # Feature gate definitions
│   ├── metrics/              # Prometheus metrics
│   ├── scheduler/            # Scheduling algorithm, preemption, flavor assignment
│   ├── util/                 # Shared utilities
│   ├── visibility/           # Visibility API server
│   ├── webhooks/             # Validating/mutating webhooks
│   └── workload/             # Workload helper functions
├── site/                     # Hugo documentation site
└── test/
    ├── compatibility_lifecycle/ # API backwards-compat tests
    ├── e2e/                  # End-to-end tests
    ├── integration/          # Integration tests (envtest)
    ├── performance/          # Performance/benchmark tests
    └── util/                 # Shared test helpers
```

---

## 3. Go Module & Dependencies

**Module**: `sigs.k8s.io/kueue`
**Go version**: `1.26.0`
**File**: `go.mod`

### Key direct dependencies

| Dependency | Version | Purpose |
|---|---|---|
| `sigs.k8s.io/controller-runtime` | v0.24.0 | Controller framework |
| `k8s.io/api` | v0.36.1 | Kubernetes API types |
| `k8s.io/client-go` | v0.36.1 | Kubernetes client |
| `k8s.io/kubectl` | v0.36.1 | kubectl utilities |
| `k8s.io/apiextensions-apiserver` | v0.36.1 | CRD support |
| `k8s.io/component-base` | v0.36.1 | Feature gates, metrics |
| `k8s.io/dynamic-resource-allocation` | v0.36.1 | DRA support |
| `k8s.io/autoscaler/cluster-autoscaler/apis` | v0.0.0-20240830133931 | ProvisioningRequest |
| `sigs.k8s.io/jobset` | v0.12.0 | JobSet integration |
| `sigs.k8s.io/lws` | v0.8.0 | LeaderWorkerSet integration |
| `github.com/ray-project/kuberay/ray-operator` | v1.6.1 | Ray integration |
| `github.com/kubeflow/training-operator` | v1.9.3 | Kubeflow training |
| `github.com/kubeflow/trainer/v2` | v2.2.1 | TrainJob |
| `github.com/kubeflow/mpi-operator` | v0.8.0 | MPI jobs |
| `github.com/kubeflow/spark-operator/v2` | v2.5.0 | Spark integration |
| `github.com/project-codeflare/appwrapper` | v1.2.2 | CodeFlare AppWrapper |
| `sigs.k8s.io/cluster-inventory-api` | v0.1.3 | MultiKueue ClusterProfile |
| `github.com/prometheus/client_golang` | v1.23.2 | Metrics |
| `github.com/spf13/cobra` | v1.10.2 | CLI |
| `go.uber.org/zap` | v1.28.0 | Logging |
| `go.uber.org/mock` | v0.6.0 | Test mocking |
| `github.com/onsi/ginkgo/v2` | v2.30.0 | Test framework |
| `github.com/onsi/gomega` | v1.41.0 | Test assertions |
| `github.com/cert-manager/cert-manager` | v1.20.2 | TLS certs |

---

## 4. Build System

**Primary files**: `Makefile`, `Makefile-deps.mk`, `Makefile-test.mk`, `Makefile-verify.mk`

### Key variables

| Variable | Value / Purpose |
|---|---|
| `RELEASE_VERSION` | v0.18.1 |
| `RELEASE_BRANCH` | main |
| `GO_VERSION` | Extracted from go.mod (1.26) |
| `PLATFORMS` | linux/amd64, linux/arm64, linux/s390x, linux/ppc64le |
| `IMAGE_REGISTRY` | us-central1-docker.pkg.dev/k8s-staging-images/kueue |
| `BASE_IMAGE` | gcr.io/distroless/static:nonroot |
| `BIN_DIR` | bin/ |
| `ARTIFACTS` | artifacts/ |

### Major make targets

#### Development
| Target | Action |
|---|---|
| `make all` | generate + fmt + vet + build |
| `make build` | Compile `bin/manager` from `cmd/kueue/main.go` |
| `make generate` | All code/docs generation |
| `make generate-code` | DeepCopy + client-go |
| `make generate-mocks` | Interface mocks via mockgen |
| `make manifests` | CRDs, RBAC, webhooks via controller-gen |
| `make compile-crd-manifests` | Kustomize-built CRD bundle |
| `make fmt` | `go fmt ./...` |
| `make vet` | `go vet ./...` |

#### Testing
| Target | Action |
|---|---|
| `make test` | Unit tests with race detection + coverage |
| `make test-integration` | Integration tests (envtest) |
| `make test-e2e-baseline` | E2E baseline on Kind |
| `make test-e2e-extended` | E2E extended (sharded) |
| `make test-e2e-sequential-baseline` | Sequential E2E |
| `make test-multikueue-e2e-baseline` | MultiKueue E2E |
| `make test-performance` | Performance benchmarks |

#### Images & Deploy
| Target | Action |
|---|---|
| `make image-build` | Multi-platform container image |
| `make image-push` | Build + push |
| `make kind-image-build` | Local Kind image |
| `make deploy` | `kustomize build config/default \| kubectl apply` |
| `make install` | Apply CRDs |
| `make undeploy` | Remove controller |

#### Verification
| Target | Action |
|---|---|
| `make verify` | Full 3-phase verification (generation + lint + clean check) |
| `make golangci-lint` | Run golangci-lint |
| `make helm-lint` | Lint Helm chart |

#### Release
| Target | Action |
|---|---|
| `make artifacts` | All release artifacts (manifests + helm + CLI binaries) |
| `make release-artifacts` | Artifacts in `release-artifacts/` |
| `make prepare-release-branch` | Update versions for release |
| `make helm-chart-package` | Package Helm chart |

---

## 5. CI & GitHub Workflows

**Location**: `.github/workflows/`

| File | Trigger | Purpose |
|---|---|---|
| `krew-release.yml` | Release tag | Publish kubectl plugin to Krew |
| `openvex.yaml` | Release/schedule | OpenVEX security scan |
| `sbom.yaml` | Release | Software Bill of Materials |
| `sync-dependabot.yaml` | Schedule | Dependency updates |

**Note**: Main CI (test jobs) is handled by Kubernetes **Prow** (external), not GitHub Actions. Prow jobs run tests against Kubernetes versions 1.34.x, 1.35.x, and 1.36.x.

### hack/testing/ scripts

| Script | Purpose |
|---|---|
| `retry.sh` | Exponential backoff retry for flaky commands |
| `e2e-common.sh` | Kind cluster setup, image deployment helpers |
| `e2e-test.sh` | Single-cluster E2E runner |
| `e2e-multikueue-test.sh` | MultiKueue E2E runner |
| `e2e-kueueviz-backend.sh` | KueueViz backend E2E |
| `e2e-kueueviz-frontend.sh` | KueueViz frontend (npm/Cypress) E2E |
| `shard-integration-tests.sh` | Shard integration tests across parallel runs |
| `performance-test.sh` | Performance test runner |
| `compare-performance.sh` | Compare performance results |
| `get-build-logs.sh` | Fetch Prow build logs |

### hack/releasing/ scripts

| Script | Purpose |
|---|---|
| `prepare_pull.sh` | Prepare release branch (version bump) |
| `promote_pull.sh` | Promote release to production |
| `wait_for_images.sh` | Wait for image builds |
| `generate_krew.sh` | Generate Krew plugin manifest |
| `sync-notes.sh` | Synchronize release notes |

---

## 6. Helm Chart

**Location**: `charts/kueue/`
**Chart name**: kueue
**Chart version**: 0.18.1
**App version**: v0.18.1

### Templates

```
templates/
├── crd/           # 10 CRD templates
├── rbac/          # Roles, bindings, service account
├── manager/       # Deployment, config, PDB, auth proxy
├── webhook/       # Webhook service
├── visibility/    # APIService, RBAC
├── visibility-apf/# PriorityLevelConfiguration, FlowSchema
├── kueueviz/      # Frontend/backend deployments, services, ingresses
├── prometheus/    # ServiceMonitor
├── certmanager/   # Certificate, Issuer
└── internalcert/  # Internal TLS secret
```

### Helm unit tests

- `tests/kueue_test.yaml`
- `tests/manager_test.yaml`
- `tests/manager_config_test.yaml`
- `tests/certmanager_test.yaml`

---

## 7. Kustomize / Deploy Config

**Location**: `config/components/`

### 10 CRDs (config/components/crd/bases/)

| CRD | Kind |
|---|---|
| `kueue.x-k8s.io_admissionchecks.yaml` | AdmissionCheck |
| `kueue.x-k8s.io_clusterqueues.yaml` | ClusterQueue |
| `kueue.x-k8s.io_cohorts.yaml` | Cohort |
| `kueue.x-k8s.io_localqueues.yaml` | LocalQueue |
| `kueue.x-k8s.io_multikueueclusters.yaml` | MultiKueueCluster |
| `kueue.x-k8s.io_multikueueconfigs.yaml` | MultiKueueConfig |
| `kueue.x-k8s.io_provisioningrequestconfigs.yaml` | ProvisioningRequestConfig |
| `kueue.x-k8s.io_topologies.yaml` | Topology |
| `kueue.x-k8s.io_workloadpriorityclasses.yaml` | WorkloadPriorityClass |
| `kueue.x-k8s.io_workloads.yaml` | Workload |

---

## 8. Binaries & Entry Points

| Binary | Source | Output | Purpose |
|---|---|---|---|
| kueue (manager) | `cmd/kueue/main.go` | `bin/manager` | Main controller manager |
| kueuectl | `cmd/kueuectl/main.go` | `bin/kubectl-kueue` | kubectl plugin |
| importer | `cmd/importer/main.go` | `bin/importer` | Workload importer |
| kueueviz-backend | `cmd/kueueviz/backend/` | image | Dashboard API server |
| kueueviz-frontend | `cmd/kueueviz/frontend/` | image | Dashboard UI (Node.js) |
| kueue-populator | `cmd/experimental/kueue-populator/` | image | Synthetic workload generator |
| kueue-priority-booster | `cmd/experimental/kueue-priority-booster/` | image | Priority boost example |

**Supported platforms for release CLI binaries**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64

---

## 9. API Types & CRDs

All types defined in `apis/kueue/v1beta1/` (deprecated) and `apis/kueue/v1beta2/` (current).
Configuration types in `apis/config/v1beta2/`.

### 9.1 Workload

**File**: `apis/kueue/v1beta1/workload_types.go`
**Scope**: Namespaced
**ShortNames**: kwl, kueueworkload, kueueworkloads

**WorkloadSpec**:
| Field | Type | Notes |
|---|---|---|
| `podSets` | `[]PodSet` | 1-10 homogeneous pod sets, immutable after creation |
| `queueName` | `LocalQueueName` | Associated LocalQueue; mutable only if admission is null |
| `priorityClassName` | `string` | PriorityClass name (DNS pattern, max 253 chars) |
| `priority` | `*int32` | Populated from priorityClassName |
| `priorityClassSource` | `string` | `kueue.x-k8s.io/workloadpriorityclass` \| `scheduling.k8s.io/priorityclass` \| `""` |
| `active` | `*bool` | Default true; false = deactivated |
| `maximumExecutionTimeSeconds` | `*int32` | Auto-deactivation timeout (min 1) |

**PodSet**:
| Field | Type | Notes |
|---|---|---|
| `name` | `PodSetReference` | Default "main", max 63 chars |
| `template` | `corev1.PodTemplateSpec` | Only labels/annotations in metadata allowed |
| `count` | `int32` | Default 1, min 0 |
| `minCount` | `*int32` | Partial admission minimum (requires PartialAdmission gate) |
| `topologyRequest` | `*PodSetTopologyRequest` | TAS requirements |

**PodSetTopologyRequest**:
| Field | Type | Notes |
|---|---|---|
| `required` | `*string` | Required topology level |
| `preferred` | `*string` | Preferred topology level |
| `unconstrained` | `*bool` | No topology constraint |
| `podIndexLabel` | `*string` | e.g., `kubernetes.io/job-completion-index` |
| `subGroupIndexLabel` | `*string` | e.g., `jobset.sigs.k8s.io/job-index` |
| `subGroupCount` | `*int32` | Count of replicated job groups |
| `podSetSliceRequiredTopology` | `*string` | Topology level for pod set slices |
| `podSetSliceSize` | `*int32` | Size of pod set slice |

**WorkloadStatus**:
| Field | Type | Notes |
|---|---|---|
| `conditions` | `[]metav1.Condition` | Admitted, Finished, PodsReady, Evicted, Preempted, Requeued, DeactivationTarget |
| `admission` | `*Admission` | Immutable once set |
| `requeueState` | `*RequeueState` | Count + requeueAt time |
| `reclaimablePods` | `[]ReclaimablePod` | Pods no longer needing resources |
| `admissionChecks` | `[]AdmissionCheckState` | Max 8 |
| `resourceRequests` | `[]PodSetRequest` | Detailed requests for non-admitted workload |
| `accumulatedPastExecTimeSeconds` | `*int32` | Total execution time in previous cycles |
| `schedulingStats` | `*SchedulingStats` | Eviction counts etc. |
| `nominatedClusterNames` | `[]string` | MultiKueue nominated clusters (max 20) |
| `clusterName` | `*string` | Current cluster assignment (immutable once set) |
| `unhealthyNodes` | `[]UnhealthyNode` | TAS: failed nodes with pods |

**Admission**:
| Field | Type | Notes |
|---|---|---|
| `clusterQueue` | `ClusterQueueReference` | Admitting CQ |
| `podSetAssignments` | `[]PodSetAssignment` | Max 10 |

**PodSetAssignment**:
| Field | Type | Notes |
|---|---|---|
| `name` | `PodSetReference` | Default "main" |
| `flavors` | `map[ResourceName]ResourceFlavorReference` | Assigned flavors per resource |
| `resourceUsage` | `corev1.ResourceList` | Total resources (with LimitRanges, RuntimeClass) |
| `count` | `*int32` | Pod count at admission time |
| `topologyAssignment` | `*TopologyAssignment` | TAS domain assignments |
| `delayedTopologyRequest` | `*DelayedTopologyRequestState` | "Pending" \| "Ready" |

**TopologyAssignment**:
| Field | Type | Notes |
|---|---|---|
| `levels` | `[]string` | Ordered topology level keys, 1-16 items |
| `domains` | `[]TopologyDomainAssignment` | Domain + count assignments |

**AdmissionCheckState**:
| Field | Type | Notes |
|---|---|---|
| `name` | `AdmissionCheckReference` | Max 316 chars |
| `state` | `CheckState` | Pending \| Ready \| Retry \| Rejected |
| `lastTransitionTime` | `metav1.Time` | |
| `message` | `string` | Max 32768 chars |
| `requeueAfterSeconds` | `*int32` | Delay before retry |
| `retryCount` | `*int32` | Attempt counter |
| `podSetUpdates` | `[]PodSetUpdate` | Max 10 modifications from check |

**Condition reasons**:
- Admitted: `WorkloadAdmitted`, `WorkloadQuotaReserved`
- Evicted: `Preempted`, `PodsReadyTimeout`, `AdmissionCheck`, `ClusterQueueStopped`, `Deactivated`, `NodeFailures`
- Preempted: `InClusterQueue`, `InCohortReclamation`, `InCohortFairSharing`, `InCohortReclaimWhileBorrowing`
- PodsReady: `WaitForStart`, `WaitForRecovery`, `Started`, `Recovered`
- Finished: `Succeeded`, `Failed`, `OutOfSync`
- Requeued: `BackoffFinished`, `ClusterQueueRestarted`, `LocalQueueRestarted`, `RequeuingLimitExceeded`, `MaximumExecutionTimeExceeded`

---

### 9.2 ClusterQueue

**File**: `apis/kueue/v1beta1/clusterqueue_types.go`
**Scope**: Cluster-scoped
**ShortNames**: cq

**ClusterQueueSpec**:
| Field | Type | Notes |
|---|---|---|
| `resourceGroups` | `[]ResourceGroup` | Max 16 |
| `cohort` | `CohortReference` | Parent cohort for borrowing |
| `queueingStrategy` | `QueueingStrategy` | StrictFIFO \| BestEffortFIFO (default) |
| `namespaceSelector` | `*metav1.LabelSelector` | Namespaces allowed to use this CQ |
| `flavorFungibility` | `*FlavorFungibility` | Flavor fallback behavior |
| `preemption` | `*ClusterQueuePreemption` | Preemption policies |
| `admissionChecks` | `[]AdmissionCheckReference` | Required checks |
| `admissionChecksStrategy` | `*AdmissionChecksStrategy` | Per-flavor check strategy |
| `stopPolicy` | `*StopPolicy` | None \| Hold \| HoldAndDrain |
| `fairSharing` | `*FairSharing` | Weight config |
| `admissionScope` | `*AdmissionScope` | Admission fair sharing |

**ResourceGroup**:
| Field | Type | Notes |
|---|---|---|
| `coveredResources` | `[]ResourceName` | 1-64 items, max 256 total |
| `flavors` | `[]FlavorQuotas` | 1-64 items, max 256 total |

**ResourceQuota** (per flavor per resource):
| Field | Type | Notes |
|---|---|---|
| `nominalQuota` | `resource.Quantity` | Base quota |
| `borrowingLimit` | `*resource.Quantity` | Max borrowing from cohort |
| `lendingLimit` | `*resource.Quantity` | Max lending to cohort (null = all) |

**ClusterQueuePreemption**:
| Field | Type | Notes |
|---|---|---|
| `reclaimWithinCohort` | `PreemptionPolicy` | Never \| LowerPriority \| Any |
| `borrowWithinCohort` | `*BorrowWithinCohort` | Preemption when borrowing |
| `withinClusterQueue` | `PreemptionPolicy` | Never \| LowerPriority \| LowerOrNewerEqualPriority |

**FlavorFungibility**:
| Field | Type | Notes |
|---|---|---|
| `whenCanBorrow` | `FlavorFungibilityPolicy` | MayStopSearch (default) \| TryNextFlavor |
| `whenCanPreempt` | `FlavorFungibilityPolicy` | MayStopSearch \| TryNextFlavor (default) |
| `preference` | `*FlavorFungibilityPreference` | BorrowingOverPreemption (default) \| PreemptionOverBorrowing |

**ClusterQueueStatus**:
| Field | Type |
|---|---|
| `conditions` | `[]metav1.Condition` (Active) |
| `flavorsReservation` | `[]FlavorUsage` |
| `flavorsUsage` | `[]FlavorUsage` |
| `pendingWorkloads` | `int32` |
| `reservingWorkloads` | `int32` |
| `admittedWorkloads` | `int32` |
| `fairSharing` | `*FairSharingStatus` |

---

### 9.3 LocalQueue

**File**: `apis/kueue/v1beta1/localqueue_types.go`
**Scope**: Namespaced
**ShortNames**: queue, queues, lq

**LocalQueueSpec**:
| Field | Type | Notes |
|---|---|---|
| `clusterQueue` | `ClusterQueueReference` | Immutable |
| `stopPolicy` | `*StopPolicy` | None \| Hold \| HoldAndDrain |
| `fairSharing` | `*FairSharing` | Admission fair sharing weight |

**LocalQueueStatus**: `conditions`, `pendingWorkloads`, `reservingWorkloads`, `admittedWorkloads`, `flavorsReservation`, `flavorUsage`, `fairSharing`

---

### 9.4 ResourceFlavor

**File**: `apis/kueue/v1beta1/resourceflavor_types.go`
**Scope**: Cluster-scoped
**ShortNames**: flavor, flavors, rf

**ResourceFlavorSpec**:
| Field | Type | Notes |
|---|---|---|
| `nodeLabels` | `map[string]string` | Max 8; required + immutable if topologyName set |
| `nodeTaints` | `[]corev1.Taint` | Max 8; only NoSchedule/NoExecute evaluated |
| `tolerations` | `[]corev1.Toleration` | Max 8; added to admitted pods |
| `topologyName` | `*TopologyReference` | Associated Topology for TAS |

---

### 9.5 AdmissionCheck

**Scope**: Cluster-scoped
**ShortNames**: ac

**AdmissionCheckSpec**:
| Field | Type | Notes |
|---|---|---|
| `controllerName` | `string` | Immutable |
| `retryDelayMinutes` | `*int64` | Deprecated since v0.8; default 15 |
| `parameters` | `*AdmissionCheckParametersReference` | GVK + name of config object |

**AdmissionCheckStatus**: `conditions` (Active)
**CheckState enum**: Pending, Ready, Retry, Rejected

---

### 9.6 Cohort

**Scope**: Cluster-scoped
**ShortNames**: co

**CohortSpec**:
| Field | Type | Notes |
|---|---|---|
| `parentName` | `CohortReference` | Default None = root |
| `resourceGroups` | `[]ResourceGroup` | Max 16 |
| `fairSharing` | `*FairSharing` | Weight |

---

### 9.7 WorkloadPriorityClass

**Scope**: Cluster-scoped
**ShortNames**: wpc

**Fields**: `value int32`, `description string`

---

### 9.8 Topology (TAS)

**File**: `apis/kueue/v1beta1/topology_types.go`
**Scope**: Cluster-scoped
**ShortNames**: topo

**TopologySpec**:
- `levels []TopologyLevel` — 1-16 levels; unique, immutable; `kubernetes.io/hostname` only at lowest level

**TopologyLevel**: `nodeLabel string` (DNS+path pattern, max 316 chars)

**Key annotations on PodTemplate / Workload**:
| Annotation | Purpose |
|---|---|
| `kueue.x-k8s.io/podset-required-topology` | Required topology level |
| `kueue.x-k8s.io/podset-preferred-topology` | Preferred topology level |
| `kueue.x-k8s.io/podset-unconstrained-topology` | No constraint |
| `kueue.x-k8s.io/podset-slice-required-topology` | Topology for slices |
| `kueue.x-k8s.io/podset-slice-size` | Slice size |

---

### 9.9 MultiKueue Types

**MultiKueueCluster** (ShortNames: mkc):
| Field | Type | Notes |
|---|---|---|
| `spec.kubeConfig.location` | `string` | Secret name or file path |
| `spec.kubeConfig.locationType` | `LocationType` | Secret (default) \| Path |

**MultiKueueConfig** (ShortNames: mkconf):
| Field | Type | Notes |
|---|---|---|
| `spec.clusters` | `[]string` | 1-20 MultiKueueCluster names |

**MultiKueueClusterStatus**: `conditions` (Active = connected)

---

### 9.10 ProvisioningRequestConfig

**ShortNames**: prc

**ProvisioningRequestConfigSpec**:
| Field | Type | Notes |
|---|---|---|
| `provisioningClassName` | `string` | Autoscaling class (max 253 chars) |
| `parameters` | `map[string]Parameter` | Max 100 entries |
| `managedResources` | `[]ResourceName` | Max 100 |
| `retryStrategy` | `*ProvisioningRequestRetryStrategy` | backoffLimitCount=3, backoffBaseSeconds=60, backoffMaxSeconds=1800 |
| `podSetUpdates` | `*ProvisioningRequestPodSetUpdates` | Node selector updates |
| `podSetMergePolicy` | `*ProvisioningRequestConfigPodSetMergePolicy` | IdenticalPodTemplates \| IdenticalWorkloadSchedulingRequirements |

---

### 9.11 Configuration Types

**File**: `apis/config/v1beta2/configuration_types.go`

Full structure:

```
Configuration
├── Namespace: string
├── ControllerManager
│   ├── Webhook: {Port, Host, CertDir}
│   ├── LeaderElection: LeaderElectionConfiguration
│   ├── Metrics
│   │   ├── BindAddress: string
│   │   ├── EnableClusterQueueResources: bool
│   │   ├── CustomLabels: []ControllerMetricsCustomLabel (max 8)
│   │   └── LocalQueueMetrics: {Enable, LocalQueueSelector}
│   ├── Health: {HealthProbeBindAddress, ReadinessEndpointName, LivenessEndpointName}
│   ├── PprofBindAddress: string
│   ├── Controller: {GroupKindConcurrency, CacheSyncTimeout}
│   └── TLS: {MinVersion, CipherSuites}
├── ManageJobsWithoutQueueName: bool (default: false)
├── ManagedJobsNamespaceSelector: *LabelSelector
├── InternalCertManagement: {Enable, WebhookServiceName, WebhookSecretName}
├── WaitForPodsReady
│   ├── Timeout: duration (required)
│   ├── BlockAdmission: *bool (default: false)
│   ├── RequeuingStrategy: {Timestamp, BackoffLimitCount, BackoffBaseSeconds, BackoffMaxSeconds}
│   └── RecoveryTimeout: *duration
├── ClientConnection: {QPS, Burst}
├── Integrations
│   ├── Frameworks: []string  (batch/job, kubeflow.org/*, ray.io/*, pod, deployment, statefulset, etc.)
│   ├── ExternalFrameworks: []string
│   └── LabelKeysToCopy: []string
├── MultiKueue
│   ├── GCInterval: duration (default: 1min)
│   ├── Origin: string
│   ├── WorkerLostTimeout: duration (default: 15min)
│   ├── DispatcherName: string
│   └── ExternalFrameworks: []MultiKueueExternalFramework
├── FairSharing: {PreemptionStrategies}
├── AdmissionFairSharing: {UsageHalfLifeTime, UsageSamplingInterval, ResourceWeights}
├── Resources
│   ├── QuotaCheckStrategy: BlockUndeclared | IgnoreUndeclared
│   ├── ExcludeResourcePrefixes: []string
│   └── Transformations: []ResourceTransformation
├── FeatureGates: map[string]bool
├── ObjectRetentionPolicies: {Workloads: {AfterFinished, AfterDeactivatedByKueue}}
└── VisibilityServer: {BindAddress, BindPort}
```

---

## 10. Feature Gates

**File**: `pkg/features/kube_features.go`

### GA / Locked (removed in v0.19)
| Gate | Since | Notes |
|---|---|---|
| `LendingLimit` | v0.6 Alpha → v0.17 GA | Cohort lending limits |
| `MultiKueueBatchJobWithManagedBy` | v0.8 Alpha → v0.17 GA | |
| `LocalQueueDefaulting` | v0.10 Alpha → v0.17 GA | |
| `HierarchicalCohorts` | v0.11 Beta → v0.17 GA | |
| `ObjectRetentionPolicies` | v0.12 Alpha → v0.17 GA | |

### Beta
| Gate | Since | Purpose |
|---|---|---|
| `PartialAdmission` | v0.4 Alpha → v0.5 Beta | Partial pod count admission |
| `FlavorFungibility` | v0.5 Beta | Try next flavor before borrowing/preempting |
| `VisibilityOnDemand` | v0.6 Alpha → v0.9 Beta | Visibility API |
| `PrioritySortingWithinCohort` | v0.6 Beta | Priority sorting in cohorts |
| `MultiKueue` | v0.6 Alpha → v0.9 Beta | Multi-cluster queueing |
| `TopologyAwareScheduling` | v0.9 Alpha → v0.14 Beta | TAS |
| `LocalQueueMetrics` | v0.10 Alpha → v0.17 Beta | LocalQueue metrics |
| `TASProfileMixed` | v0.10 Alpha → v0.15 Beta | Mixed placement algorithm |
| `AdmissionFairSharing` | v0.12 Alpha → v0.15 Beta | Admission-time fair sharing |
| `TASFailedNodeReplacement` | v0.12 Alpha → v0.14 Beta | TAS node failure recovery |
| `ElasticJobsViaWorkloadSlices` | v0.13 Alpha → v0.18 Beta | Elastic jobs via slices |
| `TASFailedNodeReplacementFailFast` | v0.13 Alpha → v0.14 Beta | Fail fast on no replacement |
| `TASReplaceNodeOnPodTermination` | v0.13 Alpha → v0.14 Beta | Replace on pod termination |
| `ManagedJobsNamespaceSelectorAlwaysRespected` | v0.13 Alpha → v0.15 Beta | |
| `FairSharingPreemptWithinNominal` | v0.17 Beta | Preempt even when within quota |
| `FairSharingPrioritizeNonBorrowing` | v0.17 Beta | Prefer non-borrowing workloads |
| `RemoveFinalizersWithStrictPatch` | v0.17 Beta | |
| `TASReplaceNodeOnNodeTaints` | v0.17 Beta | Evict on node taints |
| `AssignQueueLabelsForPods` | v0.17 Beta | |
| `SchedulingEquivalenceHashing` | v0.17 Alpha → v0.18 Beta | Skip equivalent inadmissible workloads |
| `KueueDRAIntegration` | v0.18 Beta | DRA quota accounting |
| `KueueDRARejectWorkloadsWhenDRADisabled` | v0.18 Beta | |
| `ReclaimablePods` | v0.15 Beta | Count reclaimable pods toward quota |
| `MetricForWorkloadCreationLatency` | v0.18 Beta | WorkloadCreationLatency metric |
| `MetricsForCohorts` | v0.18 Beta | Cohort metrics |
| `TASHandleOverlappingFlavors` | v0.18 Beta | Overlapping TAS flavors |
| `WorkloadIdentifierAnnotations` | v0.18 Beta | Annotations as workload identifiers |
| `FinishOrphanedWorkloads` | v0.18 Beta | Finish workloads with missing controller |
| `TLSOptions` | v0.16 Beta | TLS MinVersion/CipherSuites |

### Alpha
| Gate | Since | Purpose |
|---|---|---|
| `TASBalancedPlacement` | v0.15 Alpha | Balanced pod placement |
| `KueueDRAIntegrationExtendedResource` | v0.18 Alpha | DRA extended resources |
| `KueueDRAIntegrationPartitionableDevices` | v0.18 Alpha | Partitionable DRA devices |
| `MultiKueueAdaptersForCustomJobs` | v0.14 Alpha → v0.15 Beta | Custom adapters |
| `FailureRecoveryPolicy` | v0.15 Alpha | Pod termination recovery |
| `SparkApplicationIntegration` | v0.17 Alpha | Spark integration |
| `MultiKueueOrchestratedPreemption` | v0.17 Alpha | MultiKueue preemption |
| `PriorityBoost` | v0.17 Alpha | Priority boost via annotation |
| `AdmissionGatedBy` | v0.17 Alpha → v0.19 Beta | Gate admission via annotation |
| `ShortWorkloadNames` | v0.17 Alpha | 63-char label limit |
| `FastQuotaReleaseInPodIntegration` | v0.17 Alpha | Immediate quota release on preemption |
| `ConcurrentAdmission` | v0.18 Alpha | Pursue multiple flavors concurrently |
| `QuotaCheckStrategy` | v0.18 Alpha | BlockUndeclared vs IgnoreUndeclared |
| `TASRespectNodeAffinityPreferred` | v0.18 Alpha | Node affinity in TAS |
| `MultiKueueManagerQuotaAutomation` | v0.18 Alpha | Auto-sync quota from workers |
| `WorkloadPriorityClassDefaulting` | v0.18 Alpha | Default WorkloadPriorityClass |
| `TASMultiLayerTopology` | v0.17 Alpha | Multi-layer topology |
| `ElasticJobsViaWorkloadSlicesWithTAS` | v0.17 Alpha | Elastic + TAS |

### Deprecated (remove in v0.19)
- `DRAExtendedResources` → use `KueueDRAIntegrationExtendedResource`
- `MultiKueueAllowInsecureKubeconfigs` → removed
- `LendingLimit`, `MultiKueueBatchJobWithManagedBy`, `LocalQueueDefaulting`, `HierarchicalCohorts`, `ObjectRetentionPolicies`, `SanitizePodSets`, `PropagateBatchJobLabelsToWorkload`, `LocalQueueDefaulting`

---

## 11. Scheduler & Admission System

### 11.1 Scheduler loop

**File**: `pkg/scheduler/scheduler.go`

**Struct**: `Scheduler`
- `queues *qcache.Manager` — queue manager
- `cache *schdcache.Cache` — in-memory cache
- `preemptor *preemption.Preemptor`
- `fairSharing *config.FairSharing`
- `schedulingCycle int64`

**`schedule(ctx) wait.SpeedSignal`** — 7-phase loop:

1. **Get head workloads** — `s.queues.Heads(ctx)` — one per LocalQueue
2. **Take snapshot** — `s.cache.Snapshot(ctx)` — immutable clone
3. **Nominate** — `s.nominate(ctx, heads, snapshot)` — assign flavors + find preemption targets
4. **Create iterator** — `makeIterator(entries, ordering, fairSharingEnabled)` — orders entries
5. **Process entries** — `s.processEntry(...)` — for each: TAS update, preempt check, fits check, issue preemptions, pods-ready gate, admit
6. **Requeue failed** — `requeueAndUpdate()` → back to ClusterQueue heap
7. **Report metrics**

**Entry struct**:
```go
type entry struct {
    workload.Info
    assignment        flavorassigner.Assignment
    status            entryStatus   // nominated | skipped | preemptionGated | evicted | assumed
    inadmissibleMsg   string
    requeueReason     qcache.RequeueReason
    preemptionTargets []*preemption.Target
    clusterQueueSnapshot *schdcache.ClusterQueueSnapshot
}
```

**processEntry** steps (in order):
1. Skip if more favorable ConcurrentAdmission variant already admitted
2. Evict if TAS node replacement failed
3. Skip if NoFit mode
4. Check preemption gate
5. Skip if overlapping preemption targets
6. `fits()` check (verify assignment still valid)
7. TAS usage update
8. Issue preemptions
9. Pods-ready gating (`cache.WaitForPodsReady()`)
10. Evict less-favorable sibling (ConcurrentAdmission migration)
11. `admit()` — `assumeWorkload()` + async `PatchAdmissionStatus()`

---

### 11.2 Nomination phase

**`nominate(ctx, workloads, snap)`** per workload:
1. Skip if already in cache
2. Check failed admission checks
3. Check ClusterQueue active
4. Validate namespace selector
5. Validate resources + LimitRange
6. `getAssignments()` → `flavorassigner.New().Assign()`
7. `preemptor.GetTargets()`

---

### 11.3 Flavor assigner

**File**: `pkg/scheduler/flavorassigner/flavorassigner.go`

**FlavorAssignmentMode**: `Fit` | `Preempt` | `NoFit`

**Assignment** struct:
- `PodSets []PodSetAssignment`
- `Borrowing int` — height of smallest cohort that fits
- `Usage workload.Usage`

**Assignment flow**:
1. Try `Fit` (no preemption needed)
2. Try `Preempt` (with preemption targets)
3. If `PartialAdmission`: reduce pod count via `podset_reducer.go` and retry
4. Return result with `RepresentativeMode()` = worst mode across all pod sets

**Related files**:
- `flavorassigner.go` — main algorithm
- `flavor_assigner_attempts.go` — attempt tracking
- `podset_reducer.go` — partial admission
- `tas_flavorassigner.go` — TAS-aware flavor assignment

---

### 11.4 Preemption

**File**: `pkg/scheduler/preemption/preemption.go`

**Preemptor struct**:
- `workloadOrdering workload.Ordering`
- `enableFairSharing bool`
- `fsStrategies []fairsharing.Strategy`
- `preemptionExpectations *expectations.Store`

**`GetTargets(log, wl, assignment, snapshot)`**:
- If FairSharing: `fairPreemptions()` — DRS-based
- If Classical: `classicalPreemptions()` — priority-based

**`IssuePreemptions(ctx, cache, preemptor, targets, cqSnap)`**:
- Runs up to 8 preemptions in parallel
- Calls `workload.Evict(wl, reason, message)` for each target
- Sets `Status.Conditions[Evicted]` and records event
- Writes event: `Preempted to accommodate a workload (UID: <uid>, JobUID: <juid>) due to <reason>; preemptor path: <path>; preemptee path: <path>`

**Preemption sub-packages**:
```
preemption/
├── preemption.go           — orchestrator
├── preempted_workloads.go  — track in-flight preemptions
├── preemption_oracle.go    — can-fit check
├── policy.go               — policy helpers
├── classical/
│   ├── candidate_generator.go     — classical candidacy
│   └── hierarchical_preemption.go — hierarchical candidacy
├── fairsharing/
│   ├── strategy.go         — FS preemption strategy
│   ├── ordering.go         — FS ordering
│   ├── target.go           — FS target selection
│   └── least_common_ancestor.go — LCA for cohort tree
└── common/
    ├── ordering.go          — shared ordering
    ├── preemption_policy.go — shared policy
    └── types.go
```

---

### 11.5 Fair sharing (DRS)

**File**: `pkg/cache/scheduler/fair_sharing.go`

**DRS (Dominant Resource Share) struct**:
- `fairWeight float64`
- `unweightedRatio float64`
- `dominantResource corev1.ResourceName`
- `borrowing bool`

**Formula**:
- Unweighted ratio = `(usage + request - quota) / lendable_resources`
- Weighted share = `unweighted_ratio / fairWeight`
- Dominant resource = resource with highest ratio

**Comparison**: Lower DRS → preferred for admission; higher DRS → preferred victim for preemption.

**Zero-weight handling**:
1. Both zero-weight and borrowing → compare unweighted ratio
2. Only one zero-weight and borrowing → zero-weight loses
3. Otherwise → compare precise weighted shares

---

### 11.6 Admission end-to-end flow

```
Workload created
  → queues.Manager.AddOrUpdateWorkload()
  → ClusterQueue heap (pending)

Scheduler picks up
  → Heads() → nominate() → flavorassigner → preemptor.GetTargets()
  → processEntry() → admit()
    → assumeWorkload() → cache.AddOrUpdateWorkload()
    → async PatchAdmissionStatus()
      → Status.Admission = {ClusterQueue, PodSetAssignments}
      → Status.Conditions[QuotaReserved] = True
      → AdmissionChecks begin (if any)
  → Status.Conditions[Admitted] = True  (after all checks pass)

Job controller unsuspends job
  → Kube-scheduler schedules pods onto matching nodes

Workload finishes
  → Job controller marks workload Finished
  → Status.Conditions[Finished] = True
  → Resources released from ClusterQueue
```

**RequeueReason enum**: `FailedAfterNomination`, `NamespaceMismatch`, `Generic`, `PreemptionGated`, `PendingPreemption`, `PendingMigration`, `PreemptionFailed`, `NoFit`, `PreemptionNoCandidates`

---

## 12. In-Memory Cache & Queue

### 12.1 Scheduler cache

**File**: `pkg/cache/scheduler/cache.go`

**Cache struct** (key fields):
- `resourceFlavors map[...]*kueue.ResourceFlavor`
- `admissionChecks map[...] AdmissionCheck`
- `hm hierarchy.Manager[*clusterQueue, *cohort]`
- `tasCache tasCache`
- `podsReadyTracking bool`

**Key methods**:
- `Snapshot(ctx, opts...)` — immutable copy for scheduling
- `AddOrUpdateWorkload(log, wl)` — add to usage tracking
- `DeleteWorkload(log, ref)` — remove, release quota
- `WaitForPodsReady(ctx)` — block until all admitted workloads ready
- `PodsReadyForAllAdmittedWorkloads(log)` — check all workloads

**TAS-related files**:
- `tas_cache.go`, `tas_flavor.go`, `tas_flavor_snapshot.go`
- `tas_balanced_placement.go`, `tas_nodes_cache.go`
- `tas_non_tas_pod_cache.go`, `tas_elastic_workloads.go`

---

### 12.2 Queue manager

**File**: `pkg/cache/queue/manager.go`

**ClusterQueue in queue**:
- `heap heap.Heap[workload.Info, workload.Reference]` — priority heap
- `inadmissibleWorkloads inadmissibleWorkloads` — failed once
- `noFitSchedulingHashes sets.Set[string]` — equivalence classes
- `inflight *workload.Info` — currently scheduling
- `afsEntryPenalties *queueafs.AfsEntryPenalties` — admission fair sharing

**Key methods**:
- `Heads(ctx)` — get head from each LocalQueue
- `RequeueWorkload(ctx, wl, reason)` — return to heap
- `QueueInadmissibleWorkloads()` — retry inadmissible set
- `QueueSecondPassIfNeeded()` — second-pass scheduling

**Heap complexity**: Push/Pop/Update/Delete all O(log N); Exists O(1)

**Admission fair sharing** (pkg/cache/queue/afs/):
- `entry_penalties.go` — tracks penalties per LocalQueue
- `consumed_resources.go` — tracks resource consumption with decay

---

### 12.3 Snapshot mechanism

**File**: `pkg/cache/scheduler/snapshot.go`

**Snapshot struct**:
```go
type Snapshot struct {
    hierarchy.Manager[*ClusterQueueSnapshot, *CohortSnapshot]
    ResourceFlavors          map[...]*kueue.ResourceFlavor
    InactiveClusterQueueSets sets.Set[...]
}
```

**Key methods**:
- `RemoveWorkload(wl)` / `AddWorkload(wl)` — simulate workload changes
- `SimulateWorkloadRemoval(workloads)` — returns revert closure
- `ClusterQueue(ref)` — get CQ snapshot

**Usage pattern**: Create → simulate preemption removals → check fits → revert

---

### 12.4 Hierarchy manager

**Files**: `pkg/cache/hierarchy/`
- `manager.go` — generic hierarchy over ClusterQueue + Cohort
- `cohort.go` — parent/child relationships
- `clusterqueue.go` — CQ in hierarchy
- `cycle.go` — cycle detection

---

## 13. Controllers & Reconcilers

### 13.1 Core controllers

**Location**: `pkg/controller/core/`

| Controller | File | Key responsibilities |
|---|---|---|
| WorkloadReconciler | `workload_controller.go` | Admission/eviction/quota mgmt; DRA; WaitForPodsReady; retention |
| ClusterQueueReconciler | `clusterqueue_controller.go` | CQ state; FairSharing metrics; notifies watchers |
| LocalQueueReconciler | `localqueue_controller.go` | LQ state; syncs with parent CQ; stop policy |
| CohortReconciler | `cohort_controller.go` | Cohort quota; parent-child tree |
| AdmissionCheckReconciler | `admissioncheck_controller.go` | AC state; Active condition |
| ResourceFlavorReconciler | `resourceflavor_controller.go` | RF lifecycle; in-use finalizer |
| WorkloadPriorityClassReconciler | `workloadpriorityclass_controller.go` | WPC validation |
| ResourceSliceReconciler | `resourceslice_controller.go` | DRA ResourceSlice (requires KueueDRAIntegrationPartitionableDevices) |

**WorkloadReconciler options**:
`WithWaitForPodsReady`, `WithWorkloadUpdateWatchers`, `WithWorkloadRetention`, `WithWorkloadRoleTracker`, `WithWorkloadCustomLabels`, `WithPreemptionExpectations`, `WithAdmissionFairSharing`, `WithDRAMapper`, `WithDRABackedResources`

**Watcher notification chain**:
```
ClusterQueueUpdateWatcher implemented by:
  LocalQueueReconciler, CohortReconciler, ResourceFlavorReconciler, AdmissionCheckReconciler

WorkloadUpdateWatcher implemented by:
  LocalQueueReconciler, ClusterQueueReconciler
```

---

### 13.2 Job framework & adapters

**Location**: `pkg/controller/jobframework/`

**GenericJob interface** (`interface.go`):
```go
type GenericJob interface {
    Object() client.Object
    IsSuspended() bool
    Suspend()
    RunWithPodSetsInfo(ctx, c, podSetsInfo) error
    RestorePodSetsInfo(podSetsInfo) bool
    Finished(ctx) (message, success, finished bool)
    PodSets(ctx, c) ([]kueue.PodSet, error)
    IsActive() bool
    PodsReady(ctx, c) bool
    GVK() schema.GroupVersionKind
}
```

**IntegrationCallbacks** (registered per framework):
```go
type IntegrationCallbacks struct {
    NewJob         func() GenericJob
    NewReconciler  ReconcilerFactory
    SetupWebhook   func(mgr, opts) error
    JobType        client.Object
    SetupIndexes   func(ctx, indexer) error
    AddToScheme    func(*runtime.Scheme) error
    MultiKueueAdapter MultiKueueAdapter  // optional
}
```

**Registration**: Each package calls `RegisterIntegration(name, callbacks)` in `init()`.

---

### 13.3 Admission check controllers

| Controller | Location | Feature gate |
|---|---|---|
| MultiKueue ACReconciler | `admissionchecks/multikueue/admissioncheck.go` | `MultiKueue` |
| Provisioning Controller | `admissionchecks/provisioning/controller.go` | `ProvisioningRequestsIntegration` |

**Provisioning controller** creates `ProvisioningRequest` resources for cluster autoscaler. Tracks admission check state transitions.

**MultiKueue ACReconciler** updates `AdmissionCheckActive` condition based on worker cluster health.

---

### 13.4 Specialized controllers

| Controller | File | Feature gate |
|---|---|---|
| ConcurrentAdmission variantReconciler | `concurrentadmission/controller.go` | `ConcurrentAdmission` |
| TerminatingPodReconciler | `failurerecovery/pod_termination_controller.go` | `FailureRecoveryPolicy` |
| TAS topologyReconciler | `tas/topology_controller.go` | `TopologyAwareScheduling` |
| TAS nodeReconciler | `tas/node_controller.go` | `TopologyAwareScheduling` |
| TAS nonTasUsageController | `tas/non_tas_usage_controller.go` | `TopologyAwareScheduling` |
| ElasticJobUngater | `elasticjobs/elastic_job_ungater.go` | `ElasticJobsViaWorkloadSlices` |
| IncrementalDispatcher | `workloaddispatcher/incrementaldispatcher.go` | MultiKueue |

---

### 13.5 Controller startup order

```
1. ResourceFlavorReconciler
2. AdmissionCheckReconciler
3. WorkloadPriorityClassReconciler
4. LocalQueueReconciler
5. CohortReconciler
6. ClusterQueueReconciler
7. WorkloadReconciler
8. ResourceSliceReconciler (conditional)
9. FailureRecovery controllers (conditional)
10. Provisioning controller (conditional)
11. MultiKueue controllers (conditional)
12. TAS controllers (conditional)
13. ElasticJobs controllers (conditional)
14. ConcurrentAdmission controllers (conditional)
15. All job framework controllers (per enabled integration)
```

---

## 14. Job Integrations

**Location**: `pkg/controller/jobs/`

| Integration | Framework name | Type | MultiKueueAdapter | Notes |
|---|---|---|---|---|
| batch/Job | `batch/job` | `batch.v1.Job` | Yes | managedBy; elastic slices |
| JobSet | `jobset.sigs.k8s.io/v1alpha2` | `JobSet` | Yes + Watcher | managedBy; prebuilt workload |
| Pod | `pod` | `core.v1.Pod` | Yes + Watcher | Pod groups via label |
| Deployment | `apps/deployment` | `apps.v1.Deployment` | No | Per-pod workloads |
| StatefulSet | `apps/statefulset` | `apps.v1.StatefulSet` | No | Per-pod workloads |
| RayJob | `ray.io/rayjob` | `RayJob` | Yes (generic factory) | managedBy |
| RayCluster | `ray.io/raycluster` | `RayCluster` | Yes (generic factory) | |
| RayService | `ray.io/rayservice` | `RayService` | Yes (generic factory) | active/pending status |
| PyTorchJob | `kubeflow.org/pytorchjob` | `PyTorchJob` | Yes | managedBy |
| TFJob | `kubeflow.org/tfjob` | `TFJob` | Yes | managedBy |
| MPIJob | `kubeflow.org/mpijob` | `MPIJob` | Yes | managedBy |
| PaddleJob | `kubeflow.org/paddlejob` | `PaddleJob` | Yes | managedBy |
| JAXJob | `kubeflow.org/jaxjob` | `JAXJob` | Yes | managedBy |
| XGBoostJob | `kubeflow.org/xgboostjob` | `XGBoostJob` | Yes | managedBy |
| TrainJob | `trainer.kubeflow.org/trainjob` | `TrainJob` | (via JobSet) | Creates JobSet internally |
| LeaderWorkerSet | `leaderworkerset.sigs.k8s.io` | `LeaderWorkerSet` | Yes + Watcher + **MultiWload** | No managedBy; replicas = workload count |
| AppWrapper | `workload.codeflare.dev/appwrapper` | `AppWrapper` | Yes | managedBy |
| SparkApplication | `sparkoperator.k8s.io/sparkapplication` | `SparkApplication` | No | |

**Pod group label**: `kueue.x-k8s.io/pod-group-name`

---

## 15. MultiKueue

### 15.1 Architecture

- **Manager cluster**: runs Kueue, scheduler, makes admission decisions
- **Worker clusters**: execute jobs; contacted via kubeconfig (Secret or file) or ClusterProfile
- **Origin label**: `kueue.x-k8s.io/multikueue-origin: <origin>` — isolates workloads per manager
- **GC interval**: default 1 minute — cleans up orphaned remote workloads

### 15.2 clustersReconciler

**File**: `pkg/controller/admissionchecks/multikueue/multikueuecluster.go`

**Key fields**: `remoteClients map[string]*remoteClient`, `gcInterval`, `origin`, `adapters`

**Reconcile flow**:
1. Load kubeconfig (Secret or ClusterProfile)
2. Validate: rejects TokenFile, InsecureSkipTLSVerify, CA file paths, exec plugins (unless ClusterProfile)
3. `setRemoteClientConfig()` → create `SelectivelyCachingClient`; start Workload watch; start job adapter watches; start queue watches (if quota automation)
4. Update `MultiKueueClusterActive` condition

**Reconnection backoff**: 5s → 10s → 20s → 40s → 80s → 160s → 300s (7 steps)
**Watch establishment timeout**: 1min → 10min (exponential)

**Garbage collection** (`runGC`):
1. List all remote Workloads with origin label
2. If local not found or pending deletion → delete remote
3. Call `adapter.DeleteRemoteObject()` to clean up job

### 15.3 wlReconciler

**File**: `pkg/controller/admissionchecks/multikueue/workload.go`

**Reconcile phases**:
1. Fetch local Workload; validate MultiKueue AC; find job adapter
2. `readGroup()` — assemble wlGroup (local + all remotes)
3. `reconcileGroup()`:
   - Detect elastic scaling
   - Delete remotes if local finished or no quota
   - Sync status if remote finished
   - Handle eviction (with workerLostTimeout grace)
   - Delete out-of-sync remotes or update (scale-down/priority)
   - Nominate worker clusters
   - Open preemption gates (if MultiKueueOrchestratedPreemption)
   - Sync workloads to workers

**Dispatcher modes**: AllAtOnce | Incremental | External
**Worker lost timeout**: default 15 min

**Out-of-sync detection**:
- Clone remote spec, strip PreemptionGates
- `equality.Semantic.DeepEqual()` comparison
- If elastic + scale-down: update; if priority changed: update; otherwise: delete

### 15.4 MultiKueueAdapter interface

```go
type MultiKueueAdapter interface {
    SyncJob(ctx, localClient, remoteClient, key, workloadName, origin string) error
    DeleteRemoteObject(ctx, localClient, remoteClient, key) error
    IsJobManagedByKueue(ctx, c, key) (bool, string, error)
    GVK() schema.GroupVersionKind
}

type MultiKueueWatcher interface {
    GetEmptyList() client.ObjectList
    WorkloadKeysFor(runtime.Object) ([]types.NamespacedName, error)
}

type MultiKueueMultiWorkloadAdapter interface {
    GetExpectedWorkloadCount(ctx, c, key) (int, error)
    GetWorkloadIndex(wl *kueue.Workload) int
}
```

### 15.5 Workload flow end-to-end

```
1. Job controller creates Workload with MultiKueue AdmissionCheck
2. Manager scheduler admits Workload; AC state = Pending
3. wlReconciler: dispatch to workers (AllAtOnce or Incremental)
   - cloneForCreate(): add MultiKueueOriginLabel, strip JobUIDLabel
4. Worker scheduler admits Workload on worker
5. adapter.SyncJob(): create Job on worker cluster
6. Worker Job controller runs the job
7. wlReconciler: detect remote AC = Ready; local AC → Ready
8. Status.clusterName = <worker>
9. Job completes; remote Workload = Finished
10. adapter.SyncJob(): pull completion status
11. workload.Finish() → local Workload = Finished
```

### 15.6 Failure & recovery

| Scenario | Recovery |
|---|---|
| Worker unreachable | Watch ends → reconnect backoff (5s–300s); workerLostTimeout (15 min) grace before requeue |
| Kubeconfig updated | Secret change → reconcile → reload; restart watches |
| Orphaned remote workload | GC loop (1 min) detects missing local → delete remote |
| Worker evicts workload | wlReconciler detects Evicted on remote → evict local → requeue |

---

## 16. Topology-Aware Scheduling (TAS)

**Location**: `pkg/controller/tas/`
**Feature gate**: `TopologyAwareScheduling` (Beta)

**TAS controllers**:
| File | Purpose |
|---|---|
| `topology_controller.go` | Reconciles Topology CRDs; watches ResourceFlavors |
| `node_controller.go` | Tracks node topology labels |
| `non_tas_usage_controller.go` | Tracks pod usage outside TAS |
| `pods.go` | Pod-level topology tracking |
| `topology_ungater.go` | Feature-gate gating logic |
| `resource_flavor.go` | Maps flavors to topology dimensions |

**Cache files** (`pkg/cache/scheduler/`):
- `tas_cache.go` — TAS usage cache
- `tas_flavor.go` — TAS flavor handling
- `tas_flavor_snapshot.go` — Snapshot for scheduling
- `tas_balanced_placement.go` — Balanced placement algorithm
- `tas_nodes_cache.go` — Node topology data
- `tas_elastic_workloads.go` — Elastic workload TAS support
- `tas_non_tas_pod_cache.go` — Non-TAS pod tracking

**Scheduler integration**: `pkg/scheduler/flavorassigner/tas_flavorassigner.go`

**TAS topology levels** (defined in Topology CRD):
- Examples: `cloud.provider.com/region`, `cloud.provider.com/zone`, `kubernetes.io/hostname`
- `kubernetes.io/hostname` must be at the lowest (innermost) level
- Max 16 levels

**TAS-specific feature gates**:
| Gate | Status | Purpose |
|---|---|---|
| `TASProfileMixed` | Beta | Switches BestFit/LeastFreeCapacity based on constraint type |
| `TASFailedNodeReplacement` | Beta | Replace failed nodes |
| `TASFailedNodeReplacementFailFast` | Beta | Fail-fast if no replacement found |
| `TASReplaceNodeOnPodTermination` | Beta | Replace on pod termination |
| `TASReplaceNodeOnNodeTaints` | Beta | Evict on node taints |
| `TASBalancedPlacement` | Alpha | Balanced placement |
| `TASMultiLayerTopology` | Alpha | Multi-layer slice topology |
| `TASRespectNodeAffinityPreferred` | Alpha | Evaluate preferred node affinity |
| `TASHandleOverlappingFlavors` | Beta | Handle flavors with overlapping nodes |
| `ElasticJobsViaWorkloadSlicesWithTAS` | Alpha | Elastic + TAS combined |

---

## 17. Concurrent Admission

**Location**: `pkg/controller/concurrentadmission/` and `pkg/workload/concurrentadmission/`
**Feature gate**: `ConcurrentAdmission` (Alpha, v0.18)

**Problem solved**: Allows a workload to simultaneously pursue multiple ResourceFlavors. Once one variant is admitted, the parent is running; better flavors continue to try to enable migration.

**Concepts**:
| Concept | Definition |
|---|---|
| Parent workload | Has label `kueue.x-k8s.io/concurrent-admission-parent: "true"` |
| Variant | Child workload (ownerRef → parent); pursues single allowed flavor |
| Flavor order | CQ's flavor list order (first = most preferred) |
| Migration | Switching to more preferable flavor while running |

**ConcurrentAdmissionMigrationMode**:
- `TryPreferredFlavors` — Keep higher-preference variants active after admission to enable migration
- `RetainFirstAdmission` — Lock to first admitted flavor; no migration

**variantReconciler flow** (`controller.go`):
1. Discover parent and all variants
2. Create variants for each flavor if missing
3. Sort by CQ flavor preference order
4. Mirror eviction between parent and variants
5. Deactivate variants on less-preferable flavors once one admitted
6. Reactivate higher-preference variants for migration
7. Sync parent admission status from variant status

**Key functions** (`concurrentadmission.go`):
- `IsParent(wl)`, `IsVariant(wl)`, `GetParentWorkloadName(wl)`
- `IsFlavorAllowedForVariant(wl, flavor)`

---

## 18. Webhooks

**Location**: `pkg/webhooks/`
**Setup**: `webhooks.go` — `Setup()` registers all webhooks

| Resource | Type | Key validation/defaulting |
|---|---|---|
| Workload | Validating + Mutating | PodSet validation; PartialAdmission minCount; TAS/elastic incompatibility; PriorityBoost annotation |
| ClusterQueue | Validating + Mutating | Finalizer injection; resource group validation; preemption/fungibility checks; ConcurrentAdmissionPolicy immutability |
| ResourceFlavor | Validating + Mutating | Flavor validation |
| Cohort | Validating + Mutating | Hierarchy validation |
| LocalQueue | Validating + Mutating | Defaults and validation |

**Key validation limits**:
- ResourceGroups: max 16 per CQ; max 256 total flavors
- Flavors: max 64 per group; max 64 resources per flavor
- CoveredResources: must match across all flavors in a group
- BorrowingLimit/LendingLimit: only valid when CQ has cohort
- AdmissionChecks: max 64 per CQ
- `ConcurrentAdmissionPolicy`: immutable on update

---

## 19. Visibility API

**Location**: `pkg/visibility/`
**Feature gate**: `VisibilityOnDemand` (Beta)
**Default port**: 8082 (configurable via `VisibilityServer.BindPort`)
**Default bind**: `0.0.0.0` (configurable)

- Standalone API server (not a webhook)
- TLS via internal cert management
- Exposes pending workloads for debugging/monitoring
- Supports filtering by LocalQueue, namespace
- Integrates with queue cache manager for live queries
- APIService manifests: `config/components/visibility/`
- Flow control: `config/components/visibility-apf/`

**Visibility API versions**: v1beta1, v1beta2 (both registered as APIService)

---

## 20. Metrics

**File**: `pkg/metrics/metrics.go`
**Registry**: `metrics.Registry.MustRegister()`

### ClusterQueue metrics

| Metric | Type | Labels | Gate |
|---|---|---|---|
| `kueue_pending_workloads` | Gauge | cluster_queue, status(active/inadmissible), replica_role | |
| `kueue_quota_reserved_workloads_total` | Counter | cluster_queue, priority_class, replica_role | |
| `kueue_quota_reserved_wait_time_seconds` | Histogram | cluster_queue, priority_class, replica_role | |
| `kueue_admitted_workloads_total` | Counter | cluster_queue, priority_class, replica_role | |
| `kueue_admission_wait_time_seconds` | Histogram | cluster_queue, priority_class, replica_role | |
| `kueue_admission_checks_wait_time_seconds` | Histogram | cluster_queue, priority_class, replica_role | |
| `kueue_queued_until_ready_wait_time_seconds` | Histogram | cluster_queue, priority_class, replica_role | |
| `kueue_admitted_until_ready_wait_time_seconds` | Histogram | cluster_queue, priority_class, replica_role | |
| `kueue_evicted_workloads_total` | Counter | cluster_queue, reason, underlying_cause, priority_class, replica_role | |
| `kueue_preempted_workloads_total` | Counter | preempting_cluster_queue, reason, replica_role | |
| `kueue_workload_eviction_latency_seconds` | Histogram | cluster_queue, reason, replica_role | |
| `kueue_reserving_active_workloads` | Gauge | cluster_queue, replica_role | |
| `kueue_admitted_active_workloads` | Gauge | cluster_queue, replica_role | |
| `kueue_finished_workloads` | Gauge | cluster_queue, replica_role | |
| `kueue_cluster_queue_by_status` | Gauge | cluster_queue, status(pending/active/terminating), replica_role | |
| `kueue_admission_cycle_preemption_skips` | Gauge | cluster_queue, replica_role | |
| `kueue_replaced_workload_slices_total` | Counter | cluster_queue, replica_role | |

### Resource metrics (gated: EnableClusterQueueResources)

| Metric | Type | Labels |
|---|---|---|
| `kueue_cluster_queue_resource_reservation` | Gauge | cluster_queue, flavor, resource, replica_role |
| `kueue_cluster_queue_resource_usage` | Gauge | cluster_queue, flavor, resource, replica_role |
| `kueue_cluster_queue_resource_pending_workloads` | Gauge | cluster_queue, flavor, resource, replica_role |
| `kueue_cluster_queue_nominal_quota` | Gauge | cohort, cluster_queue, flavor, resource, replica_role |
| `kueue_cluster_queue_borrowing_limit` | Gauge | cohort, cluster_queue, flavor, resource, replica_role |
| `kueue_cluster_queue_lending_limit` | Gauge | cohort, cluster_queue, flavor, resource, replica_role |

### LocalQueue metrics (gated: `LocalQueueMetrics`)

| Metric | Type | Labels |
|---|---|---|
| `kueue_local_queue_pending_workloads` | Gauge | name, namespace, replica_role |
| `kueue_local_queue_quota_reserved_workloads_total` | Counter | name, namespace, priority_class, replica_role |
| `kueue_local_queue_admitted_workloads_total` | Counter | name, namespace, priority_class, replica_role |
| `kueue_local_queue_evicted_workloads_total` | Counter | name, namespace, reason, underlying_cause, replica_role |
| `kueue_local_queue_by_status` | Gauge | name, namespace, active, replica_role |
| `kueue_local_queue_resource_reservation` | Gauge | name, namespace, flavor, resource, replica_role |
| `kueue_local_queue_resource_usage` | Gauge | name, namespace, flavor, resource, replica_role |

### Cohort metrics (gated: `MetricsForCohorts`)

| Metric | Type | Labels |
|---|---|---|
| `kueue_cohort_weighted_share` | Gauge | cohort, replica_role |
| `kueue_cohort_subtree_quota` | Gauge | cohort, flavor, resource, replica_role |
| `kueue_cohort_subtree_resource_reservations` | Gauge | cohort, flavor, resource, replica_role |
| `kueue_cohort_subtree_admitted_active_workloads` | Gauge | cohort, replica_role |
| `kueue_cohort_subtree_admitted_workloads_total` | Counter | cohort, replica_role |
| `kueue_cohort_info` | Gauge | cohort, parent_cohort, root_cohort, replica_role |
| `kueue_cluster_queue_info` | Gauge | cluster_queue, cohort, parent_cohort, root_cohort, replica_role |

### Fair sharing

| Metric | Type | Labels |
|---|---|---|
| `kueue_cluster_queue_weighted_share` | Gauge | cluster_queue, cohort, replica_role |

### Health / build info

| Metric | Type | Labels |
|---|---|---|
| `kueue_admission_attempts_total` | Counter | result, replica_role |
| `kueue_admission_attempt_duration_seconds` | Histogram | result, replica_role |
| `kueue_workload_creation_latency_seconds` | Histogram (gated: MetricForWorkloadCreationLatency) | job_kind, replica_role |
| `kueue_build_info` | Gauge | git_version, git_commit, build_date, go_version, compiler, platform |

**Custom labels** (gated: `CustomMetricLabels`): max 8 labels from Kubernetes labels/annotations; prefix `custom_<name>`

---

## 21. Test Infrastructure

### 21.1 Test structure

```
test/
├── util/                     # Shared helpers
│   ├── util.go               # 71KB; main helper functions
│   ├── e2e.go                # 35KB; E2E runner + cluster setup
│   ├── constants.go          # Timeouts: ShortTimeout, MediumTimeout, LongTimeout
│   ├── metrics.go            # 18KB; metric validation helpers
│   ├── multikueue.go         # 10KB; MultiKueue setup
│   ├── util_scheduling.go    # Scheduling-specific helpers
│   ├── job.go                # Job creation factories
│   ├── events.go             # Event helpers
│   └── factory.go            # Object factories
├── integration/
│   ├── singlecluster/
│   │   ├── controller/       # Core + job controller tests
│   │   ├── scheduler/        # Scheduler + preemption + fair sharing tests
│   │   ├── webhook/          # Webhook validation tests
│   │   ├── tas/              # TAS tests
│   │   ├── kueuectl/         # CLI tests
│   │   ├── importer/         # Importer tests
│   │   └── conversion/       # CRD conversion tests
│   └── multikueue/           # MultiKueue integration tests
├── e2e/
│   ├── singlecluster/
│   │   ├── baseline/         # Core: job, pod, deployment, metrics, visibility, kueuectl, fair sharing, TAS
│   │   └── extended/         # JobSet, Ray, kubeflow, LWS, AppWrapper
│   ├── sequential/
│   │   ├── baseline/         # HA, metrics, reconcile, WaitForPodsReady, retention
│   │   └── extended/         # Spark, managed-without-queue
│   ├── multikueue/           # MultiKueue E2E
│   ├── tas/                  # TAS E2E
│   ├── upgrade/              # Version upgrade tests
│   ├── dra/                  # DRA tests
│   ├── certmanager/          # Cert-Manager tests
│   └── kueueviz/             # Dashboard tests (Cypress)
└── performance/
    ├── scheduler/            # Scheduler load testing
    └── e2e/                  # E2E performance scenarios
```

### 21.2 Unit tests

- Live alongside code: `pkg/foo/bar_test.go` alongside `pkg/foo/bar.go`
- Use `testing.T` (standard Go)
- Mocks via `go.uber.org/mock`; stored in `internal/mocks/`

### 21.3 Integration tests

**Framework**: `test/integration/framework/framework.go`
Uses `envtest` (kubebuilder assets) — downloads real Kubernetes binaries.

**Key variables**:
- `ENVTEST_K8S_VERSION`: default 1.36
- `KUBEBUILDER_ASSETS`: path to downloaded binaries
- Retry: 4 attempts, 5s delay

**Ginkgo labels**:
- Controllers: `controller:workload`, `controller:localqueue`, `controller:clusterqueue`
- Jobs: `job:batch`, `job:pod`, `job:jobset`, `job:pytorch`, `job:ray`, etc.
- Features: `feature:tas`, `feature:multikueue`, `feature:fairsharing`
- Performance: `slow`, `redundant` (excluded from baseline)

**Run targets**:
```bash
make test-integration              # All (4 parallel processes)
make test-integration-baseline     # Exclude slow+redundant
make test-integration-extended     # Only slow+redundant
make test-multikueue-integration   # MultiKueue (3 parallel)
```

### 21.4 E2E tests

**E2E environment variables**:
| Variable | Default | Purpose |
|---|---|---|
| `E2E_MODE` | ci | `ci` (create/delete clusters) or `dev` (reuse) |
| `E2E_KIND_VERSION` | kindest/node:v1.36.1 | Kind node image |
| `E2E_SKIP_REINSTALL` | false | Skip Kueue reinstall (dev mode) |
| `E2E_SKIP_IMAGE_RELOAD` | false | Skip docker pull if exists |
| `E2E_ENFORCE_OPERATOR_UPDATE` | false | Force reinstall external operators |

**E2E test suites**:
| Suite | Target | K8s versions | Parallelism |
|---|---|---|---|
| baseline | singlecluster/baseline | 1.34, 1.35, 1.36 | 4 |
| extended | singlecluster/extended | 1.34, 1.35, 1.36 | 4 (3 shards) |
| sequential-baseline | sequential/baseline | 1.36 | 1 |
| sequential-extended | sequential/extended | 1.36 | 1 (2 shards) |
| tas-baseline | tas/baseline | 1.36 | — |
| tas-extended | tas/extended | 1.36 | — |
| multikueue-baseline | multikueue/baseline | 1.36 | 5 |
| multikueue-extended | multikueue/extended | 1.36 | 5 (2 shards) |
| multikueue-sequential | multikueue/sequential | 1.36 | 1 |
| upgrade | upgrade | 1.36 | — |
| dra | dra | 1.36 | — |
| certmanager | certmanager | 1.36 | — |

**Extended shard breakdown**:
- Shard 0: KubeRay tests
- Shard 1: JobSet, LWS, AppWrapper, Kubeflow (PyTorch, MPI, Trainer)
- Shard 2: KubeRay variant

**External operators in E2E**: JobSet, Kubeflow Training, Trainer v2, MPI Operator, KubeRay, AppWrapper, LeaderWorkerSet, Spark Operator, Cert-Manager, Prometheus Operator (v0.89.0), DRA Example Driver

### 21.5 Performance tests

**Location**: `test/performance/`
**Scheduler load tests**: `test/performance/scheduler/` — `configs/`, `minimalkueue/`, `runner/`, `checker/`
**E2E performance**: `test/performance/e2e/` — job and pod group scenarios
**Analysis tool**: `ginkgo-top` extracts slowest tests from JSON reports → `artifacts/*-top.yaml`

---

## 22. Code Generation

**Main target**: `make generate`

| Target | Tool | Output |
|---|---|---|
| `make generate-code` | controller-gen + code-generator | DeepCopy methods, client-go libraries |
| `make generate-mocks` | mockgen (go.uber.org/mock) | `internal/mocks/` |
| `make manifests` | controller-gen | `config/components/crd/bases/`, `config/components/rbac/`, `config/components/webhook/` |
| `make compile-crd-manifests` | kustomize | Compiled CRD bundle |
| `make generate-apiref` | genref | API reference docs |
| `make generate-kueuectl-docs` | cobra | kueuectl CLI docs |
| `make generate-metrics-tables` | custom | Metrics docs |
| `make generate-featuregates` | custom | Feature gate docs |
| `make generate-helm-docs` | helm-docs | Helm chart README |
| `make update-helm` | kustomize | Helm chart from manifests |

**Scripts**:
- `hack/tools/code-generator/generate.sh` — generates Kubernetes client libraries
- `hack/tools/compatibility-lifecycle/generate.sh` — API compatibility info

---

## 23. Linting & Verification

### Linter config: `.golangci.yaml`

**Enabled linters** (27): `copyloopvar`, `dupword`, `durationcheck`, `exptostd`, `fatcontext`, `ginkgolinter`, `gocritic`, `goheader`, `intrange`, `loggercheck`, `makezero`, `misspell`, `modernize`, `nilerr`, `nilnesserr`, `nolintlint`, `perfsprint`, `revive`, `unconvert`, `usetesting`, `forbidigo`, `staticcheck` (via exclusions), `gci`, `golines`

**Key settings**:
- `forbidigo`: bans `sort.Slice/Sort/Stable/String` → use `slices` package
- `goheader`: template `hack/boilerplate.txt` (Apache 2.0)
- `golines`: max line length 200
- `gci` import order: stdlib → third-party → `sigs.k8s.io/kueue` → blank → dot
- `revive` rules: `context-as-argument`, `deep-exit`, `empty-lines`, `increment-decrement`, `var-naming`, `use-any`, `use-slices-sort`, `use-waitgroup-go`
- `nolintlint`: requires explanation + specific linter name

### Verification pipeline (`make verify`)

**Phase 1: Code generation** — regenerates all artifacts (code, docs, CRDs, mocks, Helm)
**Phase 2: Verification checks** (read-only, 8 parallel):
- `verify-ci-lint` — golangci-lint all modules
- `verify-lint-api` — KAL (kubernetes API linter)
- `verify-fmt-verify` — gofmt
- `verify-e2e-common-test` — e2e-common_test.sh
- `verify-shell-lint` — shellcheck
- `verify-helm-verify` — Helm lint + template
- `verify-helm-unit-test` — Helm unit tests
- `verify-npm-depcheck` — npm deps
- `verify-kustomize-build` — kustomize build
- `verify-skills-lint` — Skills docs linting

**Phase 3: Clean check** — verifies `config/components apis charts/kueue client-go keps site/` are unchanged

---

## 24. Release Process

### Versioning

- Semantic versioning: `vMAJOR.MINOR.PATCH`
- Current: `v0.18.1`
- Release cycle: 2-3 months
- Release branches: `release-0.X` (main stays at HEAD)

### CHANGELOG structure

**Location**: `CHANGELOG/CHANGELOG-0.{minor}.md`

Sections:
1. Actions Required Before Upgrading
2. Changes by Kind: Feature | Bug or Regression | API Changes | Deprecation | Design Pattern / Internal Implementation

Entry format: `<Component>: <description>. (#<PR>, @<author>)`

### Cherry-pick process

`hack/cherry_pick_pull.sh` — automates backports to release branches with `--skip-version-updates` flag.

### Release artifacts

**CLI binaries** (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64):
- `kubectl-kueue` (kueuectl)

**Container images** (linux/amd64, linux/arm64, linux/s390x, linux/ppc64le):
- `kueue-controller-manager`
- `kueueviz-backend`, `kueueviz-frontend`

**Package artifacts**:
- Helm chart: `kueue-<version>.tgz`
- Krew plugin manifest
- SBOM, OpenVEX
- Compiled manifests (Kustomize bundle)

### OWNERS file

```
approvers: kueue-approvers
reviewers: kueue-reviewers
emeritus_approvers: ahg-g, alculquicondor, denkensk, gabesaba, kerthcet
labels:
  dependency-approvers: go.mod, go.sum, Makefile-deps.mk, vendor/, package.json
  test-approvers: _test.go, Makefile-test.mk, .golangci.yml
  hugo-approvers: site/hugo.toml, netlify.toml
  infra-approvers: Dockerfile, .github/dependabot.yml, .krew.yaml
  agent-approvers: AGENTS.md
```

---

## 25. Key Labels, Annotations & Constants

### Labels

| Label | Purpose |
|---|---|
| `kueue.x-k8s.io/queue-name` | LocalQueue assignment on job |
| `kueue.x-k8s.io/job-uid` | Parent job UID on Workload |
| `kueue.x-k8s.io/pod-group-name` | Pod group identifier |
| `kueue.x-k8s.io/managed` | Kueue-managed pod marker |
| `kueue.x-k8s.io/workload-uid` | Workload UID on pod |
| `kueue.x-k8s.io/multikueue-origin` | MultiKueue manager identifier |
| `kueue.x-k8s.io/concurrent-admission-parent` | Marks parent workload |

### Annotations

| Annotation | Purpose |
|---|---|
| `kueue.x-k8s.io/prebuilt-workload-name` | Prebuilt workload name on job |
| `kueue.x-k8s.io/podset-required-topology` | TAS required topology |
| `kueue.x-k8s.io/podset-preferred-topology` | TAS preferred topology |
| `kueue.x-k8s.io/podset-unconstrained-topology` | No TAS constraint |
| `kueue.x-k8s.io/podset-slice-required-topology` | TAS slice topology |
| `kueue.x-k8s.io/podset-slice-size` | TAS slice size |
| `kueue.x-k8s.io/priority-boost` | Priority boost value (alpha) |

### Finalizers

| Finalizer | Owner | Purpose |
|---|---|---|
| `kueue.x-k8s.io/resource-in-use` | ResourceFlavorReconciler | Prevent deletion while in use |

### WorkloadStatus enum values

**Status string**: `StatusPending` | `StatusQuotaReserved` | `StatusAdmitted` | `StatusFinished`

---

## 26. RBAC Summary

### Core Kueue APIs
```
kueue.x-k8s.io/workloads: get;list;watch;create;update;patch;delete
kueue.x-k8s.io/workloads/status: get;update;patch
kueue.x-k8s.io/workloads/finalizers: update
kueue.x-k8s.io/clusterqueues: get;list;watch;update
kueue.x-k8s.io/clusterqueues/status: get;update;patch
kueue.x-k8s.io/localqueues: get;list;watch
kueue.x-k8s.io/localqueues/status: get;update;patch
kueue.x-k8s.io/cohorts: get;list;watch
kueue.x-k8s.io/cohorts/status: get;update;patch
kueue.x-k8s.io/admissionchecks: get;list;watch;update
kueue.x-k8s.io/admissionchecks/status: get;update;patch
kueue.x-k8s.io/resourceflavors: get;list;watch;update
kueue.x-k8s.io/resourceflavors/finalizers: update
kueue.x-k8s.io/workloadpriorityclasses: get;list;watch
kueue.x-k8s.io/topologies: get;list;watch;update  (TAS)
```

### Kubernetes built-in
```
batch/jobs: get;list;watch;update;patch;delete
apps/deployments, apps/statefulsets: get;list;watch;update;patch;delete
core/pods: get;list;watch;delete
core/namespaces: get;list;watch
core/limitranges: get;list;watch
node.k8s.io/runtimeclasses: get;list;watch
events.k8s.io/events: create;watch;update;patch
```

### Job framework
```
jobset.sigs.k8s.io/jobsets: get;list;watch;update;patch;delete
ray.io/rayjobs, ray.io/rayclusters, ray.io/rayservices: get;list;watch;update;patch;delete
kubeflow.org/<job-type>: get;list;watch;update;patch;delete
trainer.kubeflow.org/trainjobs: get;list;watch;update;patch;delete
leaderworkerset.sigs.k8s.io/leaderworkersets: get;list;watch;update;patch;delete
workload.codeflare.dev/appwrappers: get;list;watch;update;patch;delete
sparkoperator.k8s.io/sparkapplications: get;list;watch;update;patch;delete
```

### DRA & autoscaler
```
resource.k8s.io/resourceclaims: get;list;watch
resource.k8s.io/resourceclaimtemplates: get;list;watch
resource.k8s.io/deviceclasses: get;list;watch
resource.k8s.io/resourceslices: get;list;watch
autoscaling.x-k8s.io/provisioningrequests: get;list;watch;create;update;patch;delete
core/podtemplates: get;list;watch;create;delete;update
```

---

## 27. File Path Quick Reference

### APIs
```
apis/kueue/v1beta1/workload_types.go
apis/kueue/v1beta1/clusterqueue_types.go
apis/kueue/v1beta1/localqueue_types.go
apis/kueue/v1beta1/resourceflavor_types.go
apis/kueue/v1beta1/admissioncheck_types.go  (in types.go)
apis/kueue/v1beta1/cohort_types.go
apis/kueue/v1beta1/workloadpriorityclass_types.go
apis/kueue/v1beta1/topology_types.go
apis/kueue/v1beta1/multikueue_types.go
apis/kueue/v1beta1/provisioningrequestconfig_types.go
apis/kueue/v1beta1/fairsharing_types.go
apis/config/v1beta2/configuration_types.go
apis/config/v1beta2/defaults.go
```

### Feature gates
```
pkg/features/kube_features.go
```

### Scheduler
```
pkg/scheduler/scheduler.go                          — main loop
pkg/scheduler/flavorassigner/flavorassigner.go      — flavor assignment
pkg/scheduler/flavorassigner/podset_reducer.go      — partial admission
pkg/scheduler/flavorassigner/tas_flavorassigner.go  — TAS assignment
pkg/scheduler/preemption/preemption.go              — main preemption
pkg/scheduler/preemption/classical/                 — classical preemption
pkg/scheduler/preemption/fairsharing/               — FS preemption
pkg/scheduler/fair_sharing_iterator.go              — FS admission iterator
```

### Cache
```
pkg/cache/scheduler/cache.go           — scheduler cache
pkg/cache/scheduler/snapshot.go        — snapshot
pkg/cache/scheduler/fair_sharing.go    — DRS calculation
pkg/cache/queue/manager.go             — queue manager
pkg/cache/queue/cluster_queue.go       — CQ in queue
pkg/cache/queue/inadmissible_workloads.go
pkg/cache/hierarchy/manager.go         — hierarchy
```

### Workload utilities
```
pkg/workload/workload.go
pkg/workload/resources.go
pkg/workload/usage.go
pkg/workload/concurrentadmission/concurrentadmission.go
```

### Controllers
```
pkg/controller/core/workload_controller.go
pkg/controller/core/clusterqueue_controller.go
pkg/controller/core/localqueue_controller.go
pkg/controller/core/cohort_controller.go
pkg/controller/core/admissioncheck_controller.go
pkg/controller/core/resourceflavor_controller.go
pkg/controller/jobframework/interface.go
pkg/controller/jobframework/multikueue.go
pkg/controller/jobframework/integrationmanager.go
pkg/controller/admissionchecks/multikueue/multikueuecluster.go
pkg/controller/admissionchecks/multikueue/workload.go
pkg/controller/admissionchecks/multikueue/admissioncheck.go
pkg/controller/admissionchecks/multikueue/controllers.go
pkg/controller/admissionchecks/provisioning/controller.go
pkg/controller/concurrentadmission/controller.go
pkg/controller/failurerecovery/pod_termination_controller.go
pkg/controller/tas/topology_controller.go
pkg/controller/elasticjobs/elastic_job_ungater.go
```

### Job adapters (MultiKueue)
```
pkg/controller/jobs/job/job_multikueue_adapter.go
pkg/controller/jobs/jobset/jobset_multikueue_adapter.go
pkg/controller/jobs/rayjob/rayjob_multikueue_adapter.go
pkg/controller/jobs/ray/ray_multikueue_adapter.go
pkg/controller/jobs/kubeflow/jobs/pytorchjob/pytorch_multikueue_adapter.go
pkg/controller/jobs/mpijob/mpijob_multikueue_adapter.go
pkg/controller/jobs/leaderworkerset/leaderworkerset_multikueue_adapter.go
pkg/controller/jobs/pod/pod_multikueue_adapter.go
pkg/controller/jobs/appwrapper/appwrapper_multikueue_adapter.go
```

### Webhooks & metrics
```
pkg/webhooks/webhooks.go
pkg/webhooks/workload_webhook.go
pkg/webhooks/clusterqueue_webhook.go
pkg/metrics/metrics.go
pkg/visibility/server.go
```

### Entry point
```
cmd/kueue/main.go
```

### Tests
```
test/util/util.go               — 71KB main helpers
test/util/e2e.go                — E2E runner
test/util/constants.go          — timeout constants
test/integration/framework/framework.go
test/e2e/singlecluster/baseline/job_test.go  (example)
```

### Build & CI
```
Makefile
Makefile-deps.mk
Makefile-test.mk
Makefile-verify.mk
.golangci.yaml
hack/testing/e2e-common.sh
hack/testing/retry.sh
hack/releasing/prepare_pull.sh
CHANGELOG/CHANGELOG-0.18.md
.github/PULL_REQUEST_TEMPLATE.md
```
