package bedtree

import (
	"strings"
)

type Index struct {
	tree  *BPlusTree
	count int
}

func NewIndex() *Index {
	branchFactor := 32
	return &Index{
		tree:  New(branchFactor, CompareDictionaryOrder),
		count: 0,
	}
}

func (index *Index) Size() int {
	return index.count
}

// What was I doing here???
func scoreDistance(distance float64) float64 {
	return 0.2 - distance
}

// Make 0.2 a parameter...
func (index *Index) findPossibleMatches(lastName string) (possibleIds map[int]float64) {
	results := index.tree.RangeQuery(lastName, 0.2)
	possibleIds = make(map[int]float64)

	// fmt.Printf("found %v possible last name matches for %v\n", len(lastNameResults), lastName)
	for _, result := range results {
		for _, entryId := range result.Values {
			possibleIds[entryId.(int)] = scoreDistance(result.Distance)
		}
	}

	return
}

// Does this belong here?
func normalizeEntry(name string) string {
	// remove whitespace
	name = strings.Replace(name, " ", "", -1)
	// remove full-stops
	name = strings.Replace(name, ".", "", -1)
	// lower case
	name = strings.ToLower(name)

	return name
}

type keyScorePair struct {
	key   int
	score float64
}

func sortMatchedPersons(matchedScores map[int]float64) (sortedMatches []*keyScorePair) {
	sortedMatches = make([]*keyScorePair, 0, len(matchedScores))

	for key, score := range matchedScores {
		keyScore := &keyScorePair{key, score}
		inserted := false

		for i, pair := range sortedMatches {
			if pair.score < score {
				//insert before pair
				sortedMatches = append(sortedMatches, nil)
				copy(sortedMatches[i+1:], sortedMatches[i:])
				sortedMatches[i] = keyScore
				inserted = true
				break
			}
		}

		if !inserted {
			// append to end
			sortedMatches = append(sortedMatches, keyScore)
		}
	}

	return
}

// returns key of entry or -1 if not found
func (index *Index) Find(entry string) int {
	// TODO...
	return -1
}

// returns key of entry and boolean if new
func (index *Index) Insert(entry string) (entryKey int, isNew bool) {

	normEntry := normalizeEntry(entry)
	possibleMatches := index.findPossibleMatches(normEntry)
	isNew = true

	if len(possibleMatches) > 0 {
		// fmt.Printf("found %v potential matches", len(possibleAddressIds))
		sortedMatches := sortMatchedPersons(possibleMatches)

		// for _, match := range sortedMatches {
		// 	fmt.Printf("key: %v, score: %.2f\n", match.key, match.score)
		// }

		entryKey = sortedMatches[0].key
		highScore := sortedMatches[0].score

		if highScore > 1 {
			isNew = false
			// return existing person
		}
	} else {
		// fmt.Printf("No Possible Matches for: %v, %v\n", lastName, firstName)
	}

	if isNew {
		entryKey = index.count
		index.count++
		index.tree.Put(normEntry, entryKey)
	}

	return
}
