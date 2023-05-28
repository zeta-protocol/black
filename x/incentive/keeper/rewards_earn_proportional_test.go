package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-protocol/black/x/incentive/keeper"
	"github.com/zeta-protocol/black/x/incentive/types"
	"github.com/stretchr/testify/require"
)

func TestGetProportionalRewardPeriod(t *testing.T) {
	tests := []struct {
		name                  string
		giveRewardPeriod      types.MultiRewardPeriod
		giveTotalBblackSupply  sdkmath.Int
		giveSingleBblackSupply sdkmath.Int
		wantRewardsPerSecond  sdk.DecCoins
	}{
		{
			"full amount",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ublack", 100), c("hard", 200)),
			),
			i(100),
			i(100),
			toDcs(c("ublack", 100), c("hard", 200)),
		},
		{
			"3/4 amount",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ublack", 100), c("hard", 200)),
			),
			i(10_000000),
			i(7_500000),
			toDcs(c("ublack", 75), c("hard", 150)),
		},
		{
			"half amount",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ublack", 100), c("hard", 200)),
			),
			i(100),
			i(50),
			toDcs(c("ublack", 50), c("hard", 100)),
		},
		{
			"under 1 unit",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ublack", 100), c("hard", 200)),
			),
			i(1000), // total bblack
			i(1),    // bblack supply of this specific vault
			dcs(dc("ublack", "0.1"), dc("hard", "0.2")), // rewards per second rounded to 0 if under 1ublack/1hard
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewardsPerSecond := keeper.GetProportionalRewardsPerSecond(
				tt.giveRewardPeriod,
				tt.giveTotalBblackSupply,
				tt.giveSingleBblackSupply,
			)

			require.Equal(t, tt.wantRewardsPerSecond, rewardsPerSecond)
		})
	}
}
