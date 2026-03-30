package scanapp

import (
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type runPlan struct {
	chunks         []task.Chunk
	runtimes       []*chunkRuntime
	outputPaths    batchOutputPaths
	scanOutputPath string
	openOnlyPath   string
}

type runDependencies struct {
	loadCIDRRecords          func(path, ipCol, ipCidrCol string) ([]input.CIDRRecord, error)
	loadPortSpecs            func(path string) ([]input.PortSpec, error)
	loadOrBuildRuntimeChunks func(cfg config.Config, cidrRecords []input.CIDRRecord, portSpecs []input.PortSpec) ([]task.Chunk, error)
	buildChunkRuntime        func(chunks []task.Chunk, cidrRecords []input.CIDRRecord, defaultPorts []input.PortSpec, policy runtimePolicy) ([]*chunkRuntime, error)
	resolveOutputPaths       func(output string, now time.Time) (batchOutputPaths, error)
}

func defaultRunDependencies() runDependencies {
	return runDependencies{
		loadCIDRRecords:          readCIDRFile,
		loadPortSpecs:            readPortFile,
		loadOrBuildRuntimeChunks: loadOrBuildChunks,
		buildChunkRuntime:        buildRuntime,
		resolveOutputPaths:       resolveBatchOutputPaths,
	}
}

func prepareRunPlan(cfg config.Config, inputs runInputs, deps runDependencies, now time.Time) (runPlan, error) {
	chunks, err := deps.loadOrBuildRuntimeChunks(cfg, inputs.cidrRecords, inputs.portSpecs)
	if err != nil {
		return runPlan{}, err
	}
	runtimes, err := deps.buildChunkRuntime(chunks, inputs.cidrRecords, inputs.portSpecs, runtimePolicyFromConfig(cfg))
	if err != nil {
		return runPlan{}, err
	}
	outputPaths, err := deps.resolveOutputPaths(cfg.Output, now)
	if err != nil {
		return runPlan{}, err
	}
	return runPlan{
		chunks:         chunks,
		runtimes:       runtimes,
		outputPaths:    outputPaths,
		scanOutputPath: outputPaths.scanPath,
		openOnlyPath:   outputPaths.openPath,
	}, nil
}
