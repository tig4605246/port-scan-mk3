package scanapp

import (
	"github.com/xuxiping/port-scan-mk3/pkg/scanner"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

func recordFromScanTask(task scanTask, res scanner.Result) writer.Record {
	return writer.Record{
		IP:                res.IP,
		IPCidr:            task.ipCidr,
		Port:              res.Port,
		Status:            res.Status,
		ResponseMS:        res.ResponseTimeMS,
		FabName:           task.fabName,
		CIDRName:          task.cidrName,
		ServiceLabel:      task.serviceLabel,
		Decision:          task.decision,
		PolicyID:          task.policyID,
		Reason:            task.reason,
		ExecutionKey:      task.executionKey,
		SrcIP:             task.srcIP,
		SrcNetworkSegment: task.srcNetworkSegment,
	}
}
