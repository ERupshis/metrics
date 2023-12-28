package utils

import (
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/pb"
)

func ConvertMetricToGrpcFormat(metric *networkmsg.Metric) *pb.Metric {
	if metric.MType == "gauge" {
		return &pb.Metric{
			Id:    metric.ID,
			Type:  pb.Metric_GAUGE,
			Value: *metric.Value,
		}
	} else {
		return &pb.Metric{
			Id:    metric.ID,
			Type:  pb.Metric_COUNTER,
			Delta: *metric.Delta,
		}
	}
}

func ConvertGrpcFormatToMetric(metric *pb.Metric) *networkmsg.Metric {
	if metric.Type == pb.Metric_GAUGE {
		value := metric.Value
		return &networkmsg.Metric{
			ID:    metric.Id,
			MType: "gauge",
			Value: &value,
		}
	} else if metric.Type == pb.Metric_COUNTER {
		delta := metric.Delta
		return &networkmsg.Metric{
			ID:    metric.Id,
			MType: "counter",
			Delta: &delta,
		}
	}

	return nil
}
