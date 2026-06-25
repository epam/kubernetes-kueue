# test/e2e/sequential/

End-to-end tests that must run sequentially because they modify cluster-wide configuration or test scenarios that are sensitive to concurrent cluster activity.

## Why Sequential

These tests change `kueue-manager-config` ConfigMap settings, restart the controller manager, or test HA failover — all operations that would interfere with concurrently running tests in other namespaces.

## Sub-packages

### `baseline/`

Tests that change cluster-wide Kueue config or test singular global behaviors:
- `ha_test.go` — leader election and failover
- `metrics_test.go` — controller manager metrics endpoint
- `reconcile_test.go` — controller restart and re-admission
- `visibility_server_test.go` — Visibility aggregated API server
- `waitforpodsready_test.go` — WaitForPodsReady global configuration
- `default_config_test.go` — default configuration validation
- `admission_fair_sharing_test.go` — Admission Fair Sharing global config
- `quota_check_strategy_test.go` — QuotaCheckStrategy global config
- `failure_recovery_policy_test.go` — failure recovery policy config
- `managejobswithoutqueuename_test.go` — ManageJobsWithoutQueueName config
- `podintegrationautoenablement_test.go` — Pod integration auto-enablement
- `objectretentionpolicies_test.go` — Object retention policies
- `workloadpriorityclassdefaulting_test.go` — WorkloadPriorityClass defaulting

### `extended/`

- `sparkapplication_test.go` — SparkApplication integration (sequential due to CRD install)
- `managejobswithoutqueuename_test.go` — extended config scenarios
- `workloadidentifierannotations_test.go` — workload identifier annotation behavior
