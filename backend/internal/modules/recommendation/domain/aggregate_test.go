package domain

import "testing"

func TestAggregateScoresUsesRoundedSelectedProductAverages(t *testing.T) {
	objective, breakdown := aggregateScores([]RankedProduct{
		{ObjectiveScore: 80, Breakdown: ScoreBreakdown{GoalMatch: 90, SpaceMatch: 70, BudgetMatch: 60}},
		{ObjectiveScore: 91, Breakdown: ScoreBreakdown{GoalMatch: 95, SpaceMatch: 91, BudgetMatch: 81}},
	})
	if objective != 86 || breakdown.GoalMatch != 93 || breakdown.SpaceMatch != 81 || breakdown.BudgetMatch != 71 {
		t.Fatalf("unexpected aggregate: objective=%d breakdown=%#v", objective, breakdown)
	}
}
