package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *KeeperTestSuite) TestExportAndInitGenesis() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	s.Require().NoError(keeper.SetParams(ctx, testutil.TestParams()))

	consAddr1 := sdk.ConsAddress("addr1_______________")
	consAddr2 := sdk.ConsAddress("addr2_______________")

	info1 := types.NewValidatorSigningInfo(consAddr1, int64(4), int64(3),
		time.Now().UTC().Add(100000000000), false, int64(10))
	info2 := types.NewValidatorSigningInfo(consAddr2, int64(5), int64(4),
		time.Now().UTC().Add(10000000000), false, int64(10))

	require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr1, info1))
	require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr2, info2))
	genesisState := keeper.ExportGenesis(ctx)

	require.Equal(genesisState.Params, testutil.TestParams())
	require.Len(genesisState.SigningInfos, 2)
	require.Equal(genesisState.SigningInfos[0].ValidatorSigningInfo, info1)

	// Tombstone validators after genesis shouldn't effect genesis state
	err := keeper.Tombstone(ctx, consAddr1)
	require.NoError(err)
	err = keeper.Tombstone(ctx, consAddr2)
	require.NoError(err)

	ok := keeper.IsTombstoned(ctx, consAddr1)
	require.True(ok)

	newInfo1, _ := keeper.GetValidatorSigningInfo(ctx, consAddr1)
	require.NotEqual(info1, newInfo1)

	// Initialize genesis with genesis state before tombstone
	s.stakingKeeper.EXPECT().IterateValidators(ctx, gomock.Any()).Return(nil)
	keeper.InitGenesis(ctx, s.stakingKeeper, genesisState)

	// Validator isTombstoned should return false as GenesisState is initialized
	ok = keeper.IsTombstoned(ctx, consAddr1)
	require.False(ok)

	newInfo1, _ = keeper.GetValidatorSigningInfo(ctx, consAddr1)
	newInfo2, _ := keeper.GetValidatorSigningInfo(ctx, consAddr2)
	require.Equal(info1, newInfo1)
	require.Equal(info2, newInfo2)
}
