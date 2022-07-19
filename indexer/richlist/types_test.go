package richlist

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
)

func TestRanker(t *testing.T) {
	ranker1 := Ranker{}
	ranker2 := Ranker{}

	ranker1.Enlist("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v")
	assert.Len(t, ranker1.Addresses, 1)

	ranker2.Enlist("terra17lmam6zguazs5q5u6z5mmx76uj63gldnse2pdp")
	ranker2.Enlist("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v")
	assert.Len(t, ranker2.Addresses, 2)

	ranker2.Enlist("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v")
	assert.Len(t, ranker2.Addresses, 2)

	ranker2.Unlist("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v")
	assert.Len(t, ranker2.Addresses, 1)

	ranker2.Unlist("terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v")
	assert.Len(t, ranker2.Addresses, 1)

	ranker1.Score = sdk.NewCoin("uluna", sdk.NewInt(1_000))
	ranker2.Score = sdk.NewCoin("uluna", sdk.NewInt(2_000))
	assert.True(t, ranker1.Less(ranker2))
}

func TestRichlist(t *testing.T) {
	threshold := sdk.NewCoin("uluna", sdk.NewInt(1_000))
	list := NewRichlist(0, &threshold)
	smallRanker := Ranker{Addresses: []string{"terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"}, Score: sdk.NewCoin("uluna", sdk.NewInt(999))}
	enoughRanker := Ranker{Addresses: []string{"terra1x46rqay4d3cssq8gxxvqz8xt6nwlz4td20k38v"}, Score: sdk.NewCoin("uluna", sdk.NewInt(1000))}
	enoughRanker2 := Ranker{Addresses: []string{"terra17lmam6zguazs5q5u6z5mmx76uj63gldnse2pdp"}, Score: sdk.NewCoin("uluna", sdk.NewInt(1001))}

	assert.Zero(t, list.Count())

	list.Rank(smallRanker)
	assert.Zero(t, list.Count())

	list.Rank(enoughRanker)
	assert.Equal(t, list.Count(), 1)

	list.Rank(enoughRanker2)
	assert.Equal(t, list.Count(), 2)

	extracted, _ := list.Extract(0, 1, &threshold)
	assert.Equal(t, extracted.Count(), 1)

	list.Unrank(enoughRanker)
	assert.Equal(t, list.Count(), 1)
	list.Unrank(enoughRanker2)
	assert.Zero(t, list.Count())
}
