package keeper_test

import (
	gocontext "context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func (suite *TestSuite) TestGRPCQueryAuthorization() {
	app, ctx, queryClient, addrs := suite.app, suite.ctx, suite.queryClient, suite.addrs
	var (
		req              *types.QueryGrantsRequest
		expAuthorization exported.Authorization
	)
	testCases := []struct {
		msg      string
		malleate func(require *require.Assertions)
		expError string
		postTest func(require *require.Assertions, res *types.QueryGrantsResponse)
	}{
		{
			"fail invalid granter addr",
			func(require *require.Assertions) {
				req = &types.QueryGrantsRequest{}
			},
			"empty address string is not allowed",
			func(require *require.Assertions, res *types.QueryGrantsResponse) {},
		},
		{
			"fail invalid grantee addr",
			func(require *require.Assertions) {
				req = &types.QueryGrantsRequest{
					Granter: addrs[0].String(),
				}
			},
			"empty address string is not allowed",
			func(require *require.Assertions, res *types.QueryGrantsResponse) {},
		},
		{
			"fail invalid msg-type",
			func(require *require.Assertions) {
				req = &types.QueryGrantsRequest{
					Granter:    addrs[0].String(),
					Grantee:    addrs[1].String(),
					MsgTypeUrl: "unknown",
				}
			},
			"no authorization found for unknown type",
			func(require *require.Assertions, res *types.QueryGrantsResponse) {},
		},
		{
			"Success",
			func(require *require.Assertions) {
				now := ctx.BlockHeader().Time
				newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
				expAuthorization = &banktypes.SendAuthorization{SpendLimit: newCoins}
				err := app.AuthzKeeper.SaveGrant(ctx, addrs[0], addrs[1], expAuthorization, now.Add(time.Hour))
				require.NoError(err)
				req = &types.QueryGrantsRequest{
					Granter:    addrs[1].String(),
					Grantee:    addrs[0].String(),
					MsgTypeUrl: expAuthorization.MethodName(),
				}
			},
			"",
			func(require *require.Assertions, res *types.QueryGrantsResponse) {
				var auth exported.Authorization
				require.Equal(1, len(res.Grants))
				err := suite.app.InterfaceRegistry().UnpackAny(res.Grants[0].Authorization, &auth)
				require.NoError(err)
				require.NotNil(auth)
				require.Equal(auth.String(), expAuthorization.String())
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			require := suite.Require()
			tc.malleate(require)
			result, err := queryClient.Grants(gocontext.Background(), req)
			if tc.expError == "" {
				require.NoError(err)
			} else {
				require.Error(err)
				require.Contains(err.Error(), tc.expError)
			}
			tc.postTest(require, result)
		})
	}
}

func (suite *TestSuite) TestGRPCQueryAuthorizations() {
	app, ctx, queryClient, addrs := suite.app, suite.ctx, suite.queryClient, suite.addrs
	var (
		req              *types.QueryGrantsRequest
		expAuthorization exported.Authorization
	)
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryGrantsResponse)
	}{
		{
			"fail invalid granter addr",
			func() {
				req = &types.QueryGrantsRequest{}
			},
			false,
			func(res *types.QueryGrantsResponse) {},
		},
		{
			"fail invalid grantee addr",
			func() {
				req = &types.QueryGrantsRequest{
					Granter: addrs[0].String(),
				}
			},
			false,
			func(res *types.QueryGrantsResponse) {},
		},
		{
			"Success",
			func() {
				now := ctx.BlockHeader().Time
				newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
				expAuthorization = &banktypes.SendAuthorization{SpendLimit: newCoins}
				err := app.AuthzKeeper.SaveGrant(ctx, addrs[0], addrs[1], expAuthorization, now.Add(time.Hour))
				suite.Require().NoError(err)
				req = &types.QueryGrantsRequest{
					Granter: addrs[1].String(),
					Grantee: addrs[0].String(),
				}
			},
			true,
			func(res *types.QueryGrantsResponse) {
				var auth exported.Authorization
				suite.Require().Equal(1, len(res.Grants))
				err := suite.app.InterfaceRegistry().UnpackAny(res.Grants[0].Authorization, &auth)
				suite.Require().NoError(err)
				suite.Require().NotNil(auth)
				suite.Require().Equal(auth.String(), expAuthorization.String())
			},
		},
	}
	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()
			result, err := queryClient.Grants(gocontext.Background(), req)
			if testCase.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
			testCase.postTest(result)
		})
	}
}
