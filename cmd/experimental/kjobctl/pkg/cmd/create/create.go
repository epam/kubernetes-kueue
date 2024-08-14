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

package create

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/attach"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/kueue/cmd/experimental/kjobctl/apis/v1alpha1"
	"sigs.k8s.io/kueue/cmd/experimental/kjobctl/pkg/builder"
	"sigs.k8s.io/kueue/cmd/experimental/kjobctl/pkg/cmd/completion"
	"sigs.k8s.io/kueue/cmd/experimental/kjobctl/pkg/cmd/util"
)

const (
	createJobExample = `  # Create job 
  kjobctl create job \ 
	--profile my-application-profile  \
	--cmd "sleep 5" \
	--parallelism 4 \
	--completions 4 \ 
	--request cpu=500m,ram=4Gi \
	--localqueue my-local-queue-name`
	createInteractiveExample = `  # Create interactive 
  kjobctl create interactive \ 
	--profile my-application-profile  \
	--pod-running-timeout 30s \
	--rm`
	createRayJobLong = `Create a rayjob.

KubeRay operator is required for RayJob.
How to install KubeRay operator you can find here https://ray-project.github.io/kuberay/deploy/installation/.`
	createRayJobExample = `  # Create rayjob 
  kjobctl create rayjob \ 
	--profile my-application-profile  \
	--cmd "python /home/ray/samples/sample_code.py" \
	--replicas small-group=1 \
	--min-replicas small-group=1 \ 
	--max-replicas small-group=5 \ 
	--localqueue my-local-queue-name`
	createRayClusterLong = `Create a raycluster.

KubeRay operator is required for RayCluster.
How to install KubeRay operator you can find here https://ray-project.github.io/kuberay/deploy/installation/.`
	createRayClusterExample = `  # Create raycluster 
  kjobctl create raycluster \ 
	--profile my-application-profile  \
	--replicas small-group=1 \
	--min-replicas small-group=1 \ 
	--max-replicas small-group=5 \ 
	--localqueue my-local-queue-name`
	createSlurmExample = `  # Create slurm 
  kjobctl create slurm ./script.sh \ 
	--profile my-application-profile  \
	--array 1-5`
	profileFlagName     = "profile"
	commandFlagName     = "cmd"
	parallelismFlagName = "parallelism"
	completionsFlagName = "completions"
	replicasFlagName    = string(v1alpha1.ReplicasFlag)
	minReplicasFlagName = string(v1alpha1.MinReplicasFlag)
	maxReplicasFlagName = string(v1alpha1.MaxReplicasFlag)
	requestFlagName     = "request"
	localQueueFlagName  = "localqueue"
	podRunningTimeout   = "pod-running-timeout"
	rayClusterFlagName  = string(v1alpha1.RayClusterFlag)
	arrayFlagName       = string(v1alpha1.ArrayFlag)
	cpusPerTaskFlagName = string(v1alpha1.CpusPerTaskFlag)
	gpusPerTaskFlagName = string(v1alpha1.GpusPerTaskFlag)
	memPerTaskFlagName  = string(v1alpha1.MemPerTaskFlag)
	memPerCPUFlagName   = string(v1alpha1.MemPerCPUFlag)
	memPerGPUFlagName   = string(v1alpha1.MemPerGPUFlag)
	nodesFlagName       = string(v1alpha1.NodesFlag)
	nTasksFlagName      = string(v1alpha1.NTasksFlag)
	stdoutFlagName      = string(v1alpha1.OutputFlag)
	stderrFlagName      = string(v1alpha1.ErrorFlag)
	inputFlagName       = string(v1alpha1.InputFlag)
	jobNameFlagName     = string(v1alpha1.JobNameFlag)
	partitionFlagName   = string(v1alpha1.PartitionFlag)
)

var (
	podRunningTimeoutDefault = 1 * time.Minute
)

type CreateOptions struct {
	exec.StreamOptions

	PrintFlags *genericclioptions.PrintFlags
	Config     *restclient.Config
	Attach     attach.RemoteAttach
	AttachFunc func(*CreateOptions, *corev1.Container, remotecommand.TerminalSizeQueue, *corev1.Pod) func() error

	DryRunStrategy util.DryRunStrategy

	Namespace   string
	ProfileName string
	ModeName    v1alpha1.ApplicationProfileMode

	Command              []string
	Parallelism          *int32
	Completions          *int32
	Replicas             map[string]int
	MinReplicas          map[string]int
	MaxReplicas          map[string]int
	Requests             corev1.ResourceList
	LocalQueue           string
	PodRunningTimeout    time.Duration
	RemoveInteractivePod bool
	RayCluster           string

	Script      string
	Array       string
	CpusPerTask *apiresource.Quantity
	GpusPerTask *apiresource.Quantity
	MemPerTask  *apiresource.Quantity
	MemPerCPU   *apiresource.Quantity
	MemPerGPU   *apiresource.Quantity
	Nodes       *int32
	NTasks      *int32
	Output      string
	Error       string
	Input       string
	JobName     string
	Partition   string

	UserSpecifiedCommand     string
	UserSpecifiedParallelism int32
	UserSpecifiedCompletions int32
	UserSpecifiedRequest     map[string]string
	UserSpecifiedCpusPerTask string
	UserSpecifiedGpusPerTask string
	UserSpecifiedMemPerTask  string
	UserSpecifiedMemPerCPU   string
	UserSpecifiedMemPerGPU   string
	UserSpecifiedNodes       int32
	UserSpecifiedNTasks      int32

	PrintObj printers.ResourcePrinterFunc

	genericiooptions.IOStreams
}

func NewCreateOptions(streams genericiooptions.IOStreams) *CreateOptions {
	return &CreateOptions{
		PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:  streams,
		StreamOptions: exec.StreamOptions{
			IOStreams: streams,
		},
		Attach:     &attach.DefaultRemoteAttach{},
		AttachFunc: defaultAttachFunc,
	}
}

type modeSubcommand struct {
	ModeName v1alpha1.ApplicationProfileMode
	Setup    func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions)
}

var createModeSubcommands = map[string]modeSubcommand{
	"job": {
		ModeName: v1alpha1.JobMode,
		Setup: func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions) {
			subcmd.Use += " --profile APPLICATION_PROFILE_NAME" +
				" [--localqueue LOCAL_QUEUE_NAME]" +
				" [--cmd COMMAND]" +
				" [--request RESOURCE_NAME=QUANTITY]" +
				" [--parallelism PARALLELISM]" +
				" [--completions COMPLETIONS]"
			subcmd.Short = "Create a job"
			subcmd.Example = createJobExample
			subcmd.Flags().StringVar(&o.UserSpecifiedCommand, commandFlagName, "",
				"Command which is associated with the resource.")
			subcmd.Flags().StringToStringVar(&o.UserSpecifiedRequest, requestFlagName, nil,
				"Request is a set of (resource name, quantity) pairs.")
			subcmd.Flags().Int32Var(&o.UserSpecifiedParallelism, parallelismFlagName, 0,
				"Parallelism specifies the maximum desired number of pods the job should run at any given time.")
			subcmd.Flags().Int32Var(&o.UserSpecifiedCompletions, completionsFlagName, 0,
				"Completions specifies the desired number of successfully finished pods.")

			cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(localQueueFlagName, completion.LocalQueueNameFunc(clientGetter)))
		},
	},
	"interactive": {
		ModeName: v1alpha1.InteractiveMode,
		Setup: func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions) {
			subcmd.Use += " --profile APPLICATION_PROFILE_NAME" +
				" [--localqueue LOCAL_QUEUE_NAME]" +
				" [--cmd COMMAND]" +
				" [--request RESOURCE_NAME=QUANTITY]" +
				" [--pod-running-timeout DURATION]" +
				" [--rm]"
			subcmd.Short = "Create an interactive shell"
			subcmd.Example = createInteractiveExample
			subcmd.Flags().StringVar(&o.UserSpecifiedCommand, commandFlagName, "",
				"Command which is associated with the resource.")
			subcmd.Flags().StringToStringVar(&o.UserSpecifiedRequest, requestFlagName, nil,
				"Request is a set of (resource name, quantity) pairs.")
			subcmd.Flags().DurationVar(&o.PodRunningTimeout, podRunningTimeout, podRunningTimeoutDefault,
				"The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one pod is running.")
			subcmd.Flags().BoolVar(&o.RemoveInteractivePod, "rm", false,
				"Remove pod when interactive session exits.")

			cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(localQueueFlagName, completion.LocalQueueNameFunc(clientGetter)))
		},
	},
	"rayjob": {
		ModeName: v1alpha1.RayJobMode,
		Setup: func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions) {
			subcmd.Use += " --profile APPLICATION_PROFILE_NAME" +
				" [--localqueue LOCAL_QUEUE_NAME]" +
				" [--cmd COMMAND]" +
				" [--replicas [WORKER_GROUP]=REPLICAS]" +
				" [--min-replicas [WORKER_GROUP]=MIN_REPLICAS]" +
				" [--max-replicas [WORKER_GROUP]=MAX_REPLICAS]"
			subcmd.Short = "Create a rayjob"
			subcmd.Long = createRayJobLong
			subcmd.Example = createRayJobExample
			subcmd.Flags().StringVar(&o.UserSpecifiedCommand, commandFlagName, "",
				"Command which is associated with the resource.")
			subcmd.Flags().StringToIntVar(&o.Replicas, replicasFlagName, nil,
				"Replicas is the number of desired Pods for this worker group.")
			subcmd.Flags().StringToIntVar(&o.MinReplicas, minReplicasFlagName, nil,
				"MinReplicas denotes the minimum number of desired Pods for this worker group.")
			subcmd.Flags().StringToIntVar(&o.MaxReplicas, maxReplicasFlagName, nil,
				"MaxReplicas denotes the maximum number of desired Pods for this worker group, and the default value is maxInt32.")
			subcmd.Flags().StringVar(&o.RayCluster, rayClusterFlagName, "",
				"Existing ray cluster on which the job will be created.")

			subcmd.MarkFlagsMutuallyExclusive(rayClusterFlagName, replicasFlagName)
			subcmd.MarkFlagsMutuallyExclusive(rayClusterFlagName, minReplicasFlagName)
			subcmd.MarkFlagsMutuallyExclusive(rayClusterFlagName, maxReplicasFlagName)
			subcmd.MarkFlagsMutuallyExclusive(rayClusterFlagName, localQueueFlagName)

			cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(localQueueFlagName, completion.LocalQueueNameFunc(clientGetter)))
		},
	},
	"raycluster": {
		ModeName: v1alpha1.RayClusterMode,
		Setup: func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions) {
			subcmd.Use += " --profile APPLICATION_PROFILE_NAME" +
				" [--localqueue LOCAL_QUEUE_NAME]" +
				" [--replicas [WORKER_GROUP]=REPLICAS]" +
				" [--min-replicas [WORKER_GROUP]=MIN_REPLICAS]" +
				" [--max-replicas [WORKER_GROUP]=MAX_REPLICAS]"
			subcmd.Short = "Create a raycluster"
			subcmd.Long = createRayClusterLong
			subcmd.Example = createRayClusterExample
			subcmd.Flags().StringToIntVar(&o.Replicas, replicasFlagName, nil,
				"Replicas is the number of desired Pods for this worker group.")
			subcmd.Flags().StringToIntVar(&o.MinReplicas, minReplicasFlagName, nil,
				"MinReplicas denotes the minimum number of desired Pods for this worker group.")
			subcmd.Flags().StringToIntVar(&o.MaxReplicas, maxReplicasFlagName, nil,
				"MaxReplicas denotes the maximum number of desired Pods for this worker group, and the default value is maxInt32.")

			cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(localQueueFlagName, completion.LocalQueueNameFunc(clientGetter)))
		},
	},
	"slurm": {
		ModeName: v1alpha1.SlurmMode,
		Setup: func(clientGetter util.ClientGetter, subcmd *cobra.Command, o *CreateOptions) {
			subcmd.Use += " SCRIPT --profile APPLICATION_PROFILE_NAME" +
				" --array ARRAY" +
				" [--cpus-per-task QUANTITY]" +
				" [--gpus-per-task QUANTITY]" +
				" [--mem-per-task QUANTITY]" +
				" [--mem-per-cpu QUANTITY]" +
				" [--mem-per-gpu QUANTITY]" +
				" [--nodes COUNT]" +
				" [--ntasks COUNT]" +
				" [--stdout FILENAME_PATTERN]" +
				" [--stderr FILENAME_PATTERN]" +
				" [--input FILENAME_PATTERN]" +
				" [--job-name NAME]" +
				" [--partition NAME]"
			subcmd.Short = "Create a raycluster"
			subcmd.Example = createSlurmExample
			subcmd.Args = cobra.ExactArgs(1)
			subcmd.Flags().StringVarP(&o.Array, arrayFlagName, "a", "",
				`Submit a job array, multiple jobs to be executed with identical parameters. 
The indexes specification identifies what array index values should be used. 
Multiple values may be specified using a comma separated list and/or a range of values with a "-" separator. For example, "--array=0-15" or "--array=0,6,16-32". 
A maximum number of simultaneously running tasks from the job array may be specified using a "%" separator. For example "--array=0-15%4" will limit the number of simultaneously running tasks from this job array to 4. 
The minimum index value is 0. The maximum index value is 2147483647.`)
			subcmd.Flags().StringVarP(&o.UserSpecifiedCpusPerTask, cpusPerTaskFlagName, "c", "",
				"How much cpus a container inside a pod requires.")
			subcmd.Flags().StringVar(&o.UserSpecifiedGpusPerTask, gpusPerTaskFlagName, "",
				"How much gpus a container inside a pod requires.")
			subcmd.Flags().StringVar(&o.UserSpecifiedMemPerTask, memPerTaskFlagName, "",
				"How much memory a container requires.")
			subcmd.Flags().StringVar(&o.UserSpecifiedMemPerCPU, memPerCPUFlagName, "",
				"How much memory a container requires, it multiplies the number of requested cpus per task by mem-per-cpu.")
			subcmd.Flags().StringVar(&o.UserSpecifiedMemPerGPU, memPerGPUFlagName, "",
				"How much memory a container requires, it multiplies the number of requested gpus per task by mem-per-gpu.")
			subcmd.Flags().Int32Var(&o.UserSpecifiedNodes, nodesFlagName, 0,
				"Number of pods to be used at a time.")
			subcmd.Flags().Int32Var(&o.UserSpecifiedNTasks, nTasksFlagName, 0,
				"Number of identical containers inside of a pod, usually 1.")
			subcmd.Flags().StringVar(&o.Output, stdoutFlagName, "",
				"Where to redirect the standard output stream of a task. If not passed it proceeds to stdout, and is available via kubectl logs.")
			subcmd.Flags().StringVar(&o.Error, stderrFlagName, "",
				"Where to redirect std error stream of a task.  If not passed it proceeds to stdout, and is available via kubectl logs.")
			subcmd.Flags().StringVar(&o.Input, inputFlagName, "",
				"What to pipe into the script.")
			subcmd.Flags().StringVar(&o.JobName, jobNameFlagName, "",
				"What is the job name.")
			subcmd.Flags().StringVar(&o.Partition, partitionFlagName, "",
				"Local queue name.")

			_ = subcmd.MarkFlagRequired(arrayFlagName)

			cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(partitionFlagName, completion.LocalQueueNameFunc(clientGetter)))
		},
	},
}

func NewCreateCmd(clientGetter util.ClientGetter, streams genericiooptions.IOStreams, clock clock.Clock) *cobra.Command {
	o := NewCreateOptions(streams)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		Example: fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			createJobExample,
			createInteractiveExample,
			createRayJobExample,
			createRayClusterExample,
			createSlurmExample,
		),
	}

	for modeName, modeSubcommand := range createModeSubcommands {
		subcmd := &cobra.Command{
			Use:                   modeName,
			DisableFlagsInUseLine: true,
			Args:                  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true

				err := o.Complete(clientGetter, cmd, args)
				if err != nil {
					return err
				}

				return o.Run(cmd.Context(), clientGetter, clock.Now())
			},
		}

		o.PrintFlags.AddFlags(subcmd)

		subcmd.Flags().StringVarP(&o.ProfileName, profileFlagName, "p", "",
			"Application profile contains a template (with defaults set) for running a specific type of application.")
		subcmd.Flags().StringVar(&o.LocalQueue, localQueueFlagName, "",
			"Kueue localqueue name which is associated with the resource.")
		modeSubcommand.Setup(clientGetter, subcmd, o)

		util.AddDryRunFlag(subcmd)

		_ = subcmd.MarkFlagRequired(profileFlagName)

		cobra.CheckErr(subcmd.RegisterFlagCompletionFunc(profileFlagName, completion.ApplicationProfileNameFunc(clientGetter)))

		cmd.AddCommand(subcmd)
	}

	return cmd
}

func (o *CreateOptions) Complete(clientGetter util.ClientGetter, cmd *cobra.Command, args []string) error {
	currentSubcommand := createModeSubcommands[cmd.Name()]
	o.ModeName = currentSubcommand.ModeName

	var err error

	o.Namespace, _, err = clientGetter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if o.UserSpecifiedCommand != "" {
		o.Command = strings.Fields(o.UserSpecifiedCommand)
	}

	if cmd.Flags().Changed(parallelismFlagName) {
		o.Parallelism = ptr.To(o.UserSpecifiedParallelism)
	}

	if cmd.Flags().Changed(completionsFlagName) {
		o.Completions = ptr.To(o.UserSpecifiedCompletions)
	}

	if len(o.UserSpecifiedRequest) > 0 {
		o.Requests = make(corev1.ResourceList)
		for key, value := range o.UserSpecifiedRequest {
			quantity, err := apiresource.ParseQuantity(value)
			if err != nil {
				return err
			}
			o.Requests[corev1.ResourceName(key)] = quantity
		}
	}

	if cmd.Flags().Changed(podRunningTimeout) && o.PodRunningTimeout <= 0 {
		return errors.New("--pod-running-timeout must be higher than zero")
	}

	if len(args) > 0 {
		o.Script = args[0]
	}

	if cmd.Flags().Changed(cpusPerTaskFlagName) {
		quantity, err := apiresource.ParseQuantity(o.UserSpecifiedCpusPerTask)
		if err != nil {
			return err
		}
		o.CpusPerTask = &quantity
	}

	if cmd.Flags().Changed(gpusPerTaskFlagName) {
		quantity, err := apiresource.ParseQuantity(o.UserSpecifiedGpusPerTask)
		if err != nil {
			return err
		}
		o.GpusPerTask = &quantity
	}

	if cmd.Flags().Changed(memPerTaskFlagName) {
		quantity, err := apiresource.ParseQuantity(o.UserSpecifiedMemPerTask)
		if err != nil {
			return err
		}
		o.MemPerTask = &quantity
	}

	if cmd.Flags().Changed(memPerCPUFlagName) {
		quantity, err := apiresource.ParseQuantity(o.UserSpecifiedMemPerCPU)
		if err != nil {
			return err
		}
		o.MemPerCPU = &quantity
	}

	if cmd.Flags().Changed(memPerGPUFlagName) {
		quantity, err := apiresource.ParseQuantity(o.UserSpecifiedMemPerGPU)
		if err != nil {
			return err
		}
		o.MemPerGPU = &quantity
	}

	if cmd.Flags().Changed(nodesFlagName) {
		o.Nodes = &o.UserSpecifiedNodes
	}

	if cmd.Flags().Changed(nTasksFlagName) {
		o.NTasks = &o.UserSpecifiedNTasks
	}

	o.DryRunStrategy, err = util.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}

	err = util.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)
	if err != nil {
		return err
	}

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	o.PrintObj = printer.PrintObj

	o.Config, err = clientGetter.ToRESTConfig()
	if err != nil {
		return err
	}

	err = setKubernetesDefaults(o.Config)
	if err != nil {
		return err
	}

	return nil
}

func (o *CreateOptions) Run(ctx context.Context, clientGetter util.ClientGetter, runTime time.Time) error {
	objs, err := builder.NewBuilder(clientGetter, runTime).
		WithNamespace(o.Namespace).
		WithProfileName(o.ProfileName).
		WithModeName(o.ModeName).
		WithCommand(o.Command).
		WithParallelism(o.Parallelism).
		WithCompletions(o.Completions).
		WithReplicas(o.Replicas).
		WithMinReplicas(o.MinReplicas).
		WithMaxReplicas(o.MaxReplicas).
		WithRequests(o.Requests).
		WithLocalQueue(o.LocalQueue).
		WithRayCluster(o.RayCluster).
		WithScript(o.Script).
		WithArray(o.Array).
		WithCpusPerTask(o.CpusPerTask).
		WithGpusPerTask(o.GpusPerTask).
		WithMemPerTask(o.MemPerTask).
		WithMemPerCPU(o.MemPerCPU).
		WithMemPerGPU(o.MemPerGPU).
		WithNodes(o.Nodes).
		WithNTasks(o.NTasks).
		WithStdOut(o.Output).
		WithStdErr(o.Error).
		WithInput(o.Input).
		WithJobName(o.JobName).
		WithPartition(o.Partition).
		Do(ctx)
	if err != nil {
		return err
	}
	obj := objs[0]

	if o.DryRunStrategy != util.DryRunClient {
		obj, err = o.createObject(ctx, clientGetter, obj)
		if err != nil {
			return err
		}
	}

	err = o.PrintObj(obj, o.Out)
	if err != nil {
		return err
	}

	if o.ModeName == v1alpha1.InteractiveMode {
		pod := obj.(*corev1.Pod)
		return o.RunInteractivePod(ctx, clientGetter, pod.Name)
	}

	return nil
}

func (o *CreateOptions) createObject(ctx context.Context, clientGetter util.ClientGetter, obj runtime.Object) (runtime.Object, error) {
	options := metav1.CreateOptions{}
	if o.DryRunStrategy == util.DryRunServer {
		options.DryRun = []string{metav1.DryRunAll}
	}

	dynamicClient, err := clientGetter.DynamicClient()
	if err != nil {
		return nil, err
	}

	restMapper, err := clientGetter.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	gvr := mapping.Resource

	unstructured := &unstructured.Unstructured{}
	unstructured.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	unstructured, err = dynamicClient.Resource(gvr).Namespace(o.Namespace).Create(ctx, unstructured, options)
	if err != nil {
		return nil, err
	}

	createdObj := obj.DeepCopyObject()

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.UnstructuredContent(), createdObj)
	if err != nil {
		return nil, err
	}

	return createdObj, nil
}

func (o *CreateOptions) RunInteractivePod(ctx context.Context, clientGetter util.ClientGetter, podName string) error {
	k8sclient, err := clientGetter.K8sClientset()
	if err != nil {
		return err
	}

	if o.RemoveInteractivePod {
		defer func() {
			err = k8sclient.CoreV1().Pods(o.Namespace).Delete(ctx, podName, metav1.DeleteOptions{})
			if err != nil {
				fmt.Fprintln(o.ErrOut, err.Error())
			}
			fmt.Fprintf(o.Out, "pod \"%s\" deleted\n", podName)
		}()
	}

	fmt.Fprintf(o.Out, "waiting for pod \"%s\" to be running...\n", podName)
	err = waitForPodRunning(ctx, k8sclient, o.Namespace, podName, o.PodRunningTimeout)
	if err != nil {
		return err
	}

	pod, err := k8sclient.CoreV1().Pods(o.Namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = attachTTY(o, pod)
	if err != nil {
		return err
	}

	return nil
}

func attachTTY(o *CreateOptions, pod *corev1.Pod) error {
	o.Stdin = true
	o.TTY = true
	containerToAttach := &pod.Spec.Containers[0]
	if !containerToAttach.TTY {
		return fmt.Errorf("error: Unable to use a TTY - container %s did not allocate one", containerToAttach.Name)
	}

	tty := o.SetupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue
	if tty.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = tty.MonitorSize(tty.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is true
		o.ErrOut = nil
	}

	return tty.Safe(o.AttachFunc(o, containerToAttach, sizeQueue, pod))
}

func defaultAttachFunc(o *CreateOptions, containerToAttach *corev1.Container, sizeQueue remotecommand.TerminalSizeQueue, pod *corev1.Pod) func() error {
	return func() error {
		restClient, err := restclient.RESTClientFor(o.Config)
		if err != nil {
			return err
		}

		req := restClient.Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("attach")
		req.VersionedParams(&corev1.PodAttachOptions{
			Container: containerToAttach.Name,
			Stdin:     o.Stdin,
			Stdout:    o.Out != nil,
			Stderr:    o.ErrOut != nil,
			TTY:       o.TTY,
		}, scheme.ParameterCodec)

		return o.Attach.Attach(req.URL(), o.Config, o.In, o.Out, o.ErrOut, o.TTY, sizeQueue)
	}
}
