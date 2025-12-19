package handlers

import (
	"sort"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/health"
)

type BannerSummary struct {
	DownCount    int
	StaleCount   int
	UnknownCount int
	UpCount      int

	TopReasonLabel string
	TopReasonCount int
}

func SummarizeServices(views []ServiceView) BannerSummary {
	var s BannerSummary

	reasonCounts := make(map[string]int)

	for _, v := range views {
		switch {
		case v.Status == string(health.StatusDown):
			s.DownCount++
		case v.IsStale:
			s.StaleCount++
		case v.Status == string(health.StatusUnknown):
			s.UnknownCount++
		default:
			s.UpCount++
		}

		if v.Status == string(health.StatusDown) && v.ReasonLabel != "" {
			reasonCounts[v.ReasonLabel]++
		}
	}

	// pick top reason (stable tie-break: alphabetical)
	type kv struct {
		k string
		v int
	}
	var items []kv
	for k, v := range reasonCounts {
		items = append(items, kv{k: k, v: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].v != items[j].v {
			return items[i].v > items[j].v
		}
		return items[i].k < items[j].k
	})

	if len(items) > 0 {
		s.TopReasonLabel = items[0].k
		s.TopReasonCount = items[0].v
	}

	return s
}
