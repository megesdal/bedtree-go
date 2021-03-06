package bedtree

type RangeResult struct {
	Key      string
	Values   []interface{}
	Distance float64
}

func (tree *BPlusTree) RangeQuery(q string, distanceThreshold float64) []*RangeResult {

	results := make([]*RangeResult, 0)
	//recRangeQuery(q, tree.root, distanceThreshold, "", "", results)

	// TODO: is this the best way to parallelize this?
	resultChan := make(chan *RangeResult, tree.branchFactor)
	njobs := recParallelRangeQuery(q, tree.root, distanceThreshold, "", "", resultChan)

	for i := 0; i < njobs; i++ {
		jobResult := <-resultChan
		if jobResult != nil {
			results = append(results, jobResult)
		}
	}
	return results
}

func recRangeQuery(q string, node *bPlusTreeNode, distanceThreshold float64, smin string, smax string, results []*RangeResult) []*RangeResult {
	if node.isLeafNode() {
		for j, leaf := range node.splits {
			if success, editDistance := VerifyEditDistance(q, leaf, denormalizeDistance(distanceThreshold, q, leaf)); success {
				results = append(results, &RangeResult{leaf, node.data[j], normalizeDistance(editDistance, q, leaf)})
			}
		}
	} else {
		if len(node.splits) > 0 {
			if VerifyLowerBound(q, smin, node.splits[0], denormalizeDistance(distanceThreshold, q, smin)) {
				results = recRangeQuery(q, node.children[0], distanceThreshold, smin, node.splits[0], results)
			}

			for j, m := 1, len(node.splits); j < m; j++ {
				if VerifyLowerBound(q, node.splits[j-1], node.splits[j], denormalizeDistance(distanceThreshold, q, node.splits[j-1])) {
					results = recRangeQuery(q, node.children[j], distanceThreshold, node.splits[j-1], node.splits[j], results)
				}
			}

			// I want smax == "" to be interpretted like the last possible word in the alphabet?
			// which would pretty much guarantee an lcp of 0, which the empty string achieves
			if VerifyLowerBound(q, node.splits[len(node.splits)-1], smax, denormalizeDistance(distanceThreshold, q, node.splits[len(node.splits)-1])) {
				results = recRangeQuery(q, node.children[len(node.splits)], distanceThreshold, node.splits[len(node.splits)-1], smax, results)
			}
		} else {
			if len(node.children) > 0 { // should only ever be one...
				results = recRangeQuery(q, node.children[0], distanceThreshold, smin, smax, results)
			}
		}
	}
	return results
}

func recParallelRangeQuery(q string, node *bPlusTreeNode, distanceThreshold float64, smin string, smax string, resultChan chan *RangeResult) int {
	njobs := 0
	if node.isLeafNode() {
		njobs += len(node.splits)
		for j, leaf := range node.splits {
			go func(j int, leaf string) {
				if success, editDistance := VerifyEditDistance(q, leaf, denormalizeDistance(distanceThreshold, q, leaf)); success {
					resultChan <- &RangeResult{leaf, node.data[j], normalizeDistance(editDistance, q, leaf)}
				} else {
					resultChan <- nil
				}
			}(j, leaf)
		}
	} else {
		if len(node.splits) > 0 {
			if VerifyLowerBound(q, smin, node.splits[0], denormalizeDistance(distanceThreshold, q, smin)) {
				njobs += recParallelRangeQuery(q, node.children[0], distanceThreshold, smin, node.splits[0], resultChan)
			}

			for j, m := 1, len(node.splits); j < m; j++ {
				if VerifyLowerBound(q, node.splits[j-1], node.splits[j], denormalizeDistance(distanceThreshold, q, node.splits[j-1])) {
					njobs += recParallelRangeQuery(q, node.children[j], distanceThreshold, node.splits[j-1], node.splits[j], resultChan)
				}
			}

			// I want smax == "" to be interpretted like the last possible word in the alphabet?
			// which would pretty much guarantee an lcp of 0, which the empty string achieves
			if VerifyLowerBound(q, node.splits[len(node.splits)-1], smax, denormalizeDistance(distanceThreshold, q, node.splits[len(node.splits)-1])) {
				njobs += recParallelRangeQuery(q, node.children[len(node.splits)], distanceThreshold, node.splits[len(node.splits)-1], smax, resultChan)
			}
		} else {
			if len(node.children) > 0 { // should only ever be one...
				njobs += recParallelRangeQuery(q, node.children[0], distanceThreshold, smin, smax, resultChan)
			}
		}
	}

	return njobs
}

func denormalizeDistance(threshold float64, si string, sj string) int {
	return int(threshold * float64(intMax(len(si), len(sj))))
}

func normalizeDistance(editDistance int, si string, sj string) float64 {
	return float64(editDistance) / float64(intMax(len(si), len(sj)))
}
