package factory

import (
	"DecisionTreeDemoNew/vmselection/domain"
	"DecisionTreeDemoNew/vmselection/tree"
	"encoding/json"
)

func CreateDefaultConfigs() map[string]domain.NodeConfig {
	return map[string]domain.NodeConfig{
		"mds": {
			MinCPU:           4,
			MinMemory:        16,
			MinNic:           3,
			WriteIopsDensity: 16,
			SetattrRatio:     0.3,
			MdsNodeCount:     2,
			MdsPerformanceTier: map[string]int{
				"4":  25000,
				"8":  60000,
				"16": 90000,
				"24": 100000,
				"32": 110000,
				"48": 120000,
				"64": 130000,
			},
		},
		"nas": {
			MinCPU:                4,
			MinMemory:             8,
			MinNic:                3,
			ReadIopsDensity:       48,
			UnevennessLimit:       0.3,
			BandwidthReserved:     0.9,
			NASPerformanceTierX86: map[string]int{"4": 50000, "8": 100000, "16": 160000, "32": 200000},
			NASPerformanceTierARM: map[string]int{"4": 50000, "8": 100000, "16": 160000, "32": 200000},
		},
		"space": {
			MinCPU:                  4,
			MinMemory:               8,
			MinNic:                  3,
			ReadIopsDensity:         48,
			UnevennessLimit:         0.3,
			BandwidthReserved:       0.9,
			SPACEPerformanceTierX86: map[string]int{"4": 60000, "8": 120000, "16": 200000, "32": 400000},
			SPACEPerformanceTierARM: map[string]int{"4": 60000, "8": 120000, "16": 200000, "32": 400000},
		},
		"ns": {
			MinCPU:               4,
			MinMemory:            16,
			MinNic:               3,
			ReadIopsDensity:      48,
			UnevennessLimit:      0.3,
			BandwidthReserved:    0.9,
			NSPerformanceTierX86: map[string]int{"4": 30000, "8": 80000, "16": 150000, "32": 240000},
			NSPerformanceTierARM: map[string]int{"4": 30000, "8": 80000, "16": 150000, "32": 240000},
		},
	}
}

func BuildDecisionTree() *tree.RootNode {
	configs := CreateDefaultConfigs()

	mdsConfigJSON, _ := json.Marshal(configs["mds"])
	nasConfigJSON, _ := json.Marshal(configs["nas"])
	spaceConfigJSON, _ := json.Marshal(configs["space"])
	nsConfigJSON, _ := json.Marshal(configs["ns"])

	mdsNode := tree.NewMDSNode(mdsConfigJSON)
	nasNode := tree.NewNASNode(nasConfigJSON)
	spaceNode := tree.NewSPACENode(spaceConfigJSON)
	nsNode := tree.NewNASSPACENode(nsConfigJSON)

	separateAll := tree.NewSeparateAllBranch()
	separateAll.MDSNode = mdsNode
	separateAll.NASNode = nasNode
	separateAll.SPACENode = spaceNode

	separateNS := tree.NewSeparateNSBranch()
	separateNS.MDSNode = mdsNode
	separateNS.NASSPACENode = nsNode

	root := tree.NewRootNode()
	root.SeparateAll = separateAll
	root.SeparateNS = separateNS
	return root
}
