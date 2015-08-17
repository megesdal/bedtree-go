package bedtree

import (
	"testing"
)

func TestScoreDistance(t *testing.T) {
	score := scoreDistance(0.05)
	if int64(score*100) != 15 {
		t.Errorf("Expecting score of 0.15, got: %v", score)
	}
}

func TestNormalizeEntry(t *testing.T) {
	testString := "A very long name with initial K."
	testString = normalizeEntry(testString)

	if testString != "averylongnamewithinitialk" {
		t.Errorf("didn't normalize string correctly: %v", testString)
	}
}

func TestSortMatchedPersons(t *testing.T) {
	matchedScores := make(map[int]float64)

	matchedScores[1] = 0.1
	matchedScores[2] = 0.3
	matchedScores[3] = 0.2

	sortedScores := sortMatchedPersons(matchedScores)

	if sortedScores[0].score != 0.3 || sortedScores[0].key != 2 || sortedScores[1].score != 0.2 {
		t.Errorf("%v, %v, %v", sortedScores[0], sortedScores[1], sortedScores[2])
	}

}
