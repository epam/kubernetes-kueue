/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builder

import (
	"context"
	"fmt"
	"slices"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

type slurmBuilder struct {
	*Builder

	arrayIndexes arrayIndexes
}

var _ builder = (*slurmBuilder)(nil)

func (b *slurmBuilder) build(ctx context.Context) ([]runtime.Object, error) {
	// TODO: Implement method...

	return []runtime.Object{&batchv1.Job{}}, nil
}

func (b *slurmBuilder) buildIndexesMap() map[int32][]int32 {
	indexMap := make(map[int32][]int32)
	nTasks := ptr.Deref(b.nTasks, 1)
	var (
		completionIndex int32
		containerIndex  int32
	)
	for _, index := range b.arrayIndexes.Indexes {
		indexMap[completionIndex] = append(indexMap[completionIndex], index)
		containerIndex++
		if containerIndex >= nTasks {
			containerIndex = 0
			completionIndex++
		}
	}
	return indexMap
}

func (b *slurmBuilder) buildEntrypointScript() string {
	nTasks := ptr.Deref(b.nTasks, 1)

	indexesMap := b.buildIndexesMap()
	keyValues := make([]string, 0, len(indexesMap))
	for key, value := range indexesMap {
		strIndexes := make([]string, 0, len(value))
		for _, index := range value {
			strIndexes = append(strIndexes, fmt.Sprintf("%d", index))
		}
		keyValues = append(keyValues, fmt.Sprintf(`["%d"]="%s"`, key, strings.Join(strIndexes, ",")))
	}

	slices.Sort(keyValues)

	return fmt.Sprintf(`
#!/usr/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# External
# JOB_COMPLETION_INDEX  - completion index of the job.
# JOB_CONTAINER_INDEX   - container index in the container template.

# COMPLETION_INDEX=CONTAINER_INDEX1,CONTAINER_INDEX2
declare -A array_indexes=(%[1]s)

container_indexes=${array_indexes[${JOB_COMPLETION_INDEX}]}
container_indexes=(${container_indexes//,/ })

# Generated on the builder
export SLURM_ARRAY_JOB_ID=1       			# Job array’s master job ID number.
export SLURM_ARRAY_TASK_COUNT=%[2]d  		# Total number of tasks in a job array.
export SLURM_ARRAY_TASK_MAX=%[3]d    		# Job array’s maximum ID (index) number.
export SLURM_ARRAY_TASK_MIN=%[4]d    		# Job array’s minimum ID (index) number.
export SLURM_TASKS_PER_NODE=%[5]d    		# Job array’s master job ID number.
export SLURM_CPUS_PER_TASK=       			# Number of CPUs per task.
export SLURM_CPUS_ON_NODE=        			# Number of CPUs on the allocated node (actually pod).
export SLURM_JOB_CPUS_PER_NODE=   			# Count of processors available to the job on this node.
export SLURM_CPUS_PER_GPU=        			# Number of CPUs requested per allocated GPU.
export SLURM_MEM_PER_CPU=         			# Memory per CPU. Same as --mem-per-cpu .
export SLURM_MEM_PER_GPU=         			# Memory per GPU.
export SLURM_MEM_PER_NODE=        			# Memory per node. Same as --mem.
export SLURM_GPUS=                			# Number of GPUs requested (in total).
export SLURM_NTASKS=%[6]d              		# Same as -n, –ntasks. The number of tasks.
export SLURM_NTASKS_PER_NODE=$SLURM_NTASKS  # Number of tasks requested per node.
export SLURM_NPROCS=$SLURM_NTASKS       	# Same as -n, --ntasks. See $SLURM_NTASKS.
export SLURM_NNODES=%[7]d            		# Total number of nodes (actually pods) in the job’s resource allocation.
# export SLURM_SUBMIT_DIR=        			# The path of the job submission directory.
# export SLURM_SUBMIT_HOST=       			# The hostname of the node used for job submission.

# To be supported later
# export SLURM_JOB_NODELIST=        # Contains the definition (list) of the nodes (actually pods) that is assigned to the job. To be supported later.
# export SLURM_NODELIST=            # Deprecated. Same as SLURM_JOB_NODELIST. To be supported later.
# export SLURM_NTASKS_PER_SOCKET    # Number of tasks requested per socket. To be supported later.
# export SLURM_NTASKS_PER_CORE      # Number of tasks requested per core. To be supported later.
# export SLURM_NTASKS_PER_GPU       # Number of tasks requested per GPU. To be supported later.

# Calculated variables in runtime
export SLURM_JOB_ID=$(( JOB_COMPLETION_INDEX * SLURM_TASKS_PER_NODE + JOB_CONTAINER_INDEX + SLURM_ARRAY_JOB_ID ))   # The Job ID.
export SLURM_JOBID=$SLURM_JOB_ID                                                                                    # Deprecated. Same as $SLURM_JOB_ID
export SLURM_ARRAY_TASK_ID=${container_indexes[${JOB_CONTAINER_INDEX}]}												# Task ID.

bash ./script.sh
`,
		strings.Join(keyValues, " "),
		b.arrayIndexes.Count(),
		b.arrayIndexes.Max(),
		b.arrayIndexes.Min(),
		nTasks,
		nTasks,
		ptr.Deref(b.nodes, 1),
	)
}

func newSlurmBuilder(b *Builder) *slurmBuilder {
	return &slurmBuilder{Builder: b}
}
