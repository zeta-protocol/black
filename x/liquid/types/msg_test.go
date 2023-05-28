package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmath "cosmossdk.io/math"

	"github.com/zeta-protocol/black/x/liquid/types"
)

func TestMsgMintDerivative_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("black1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgMintDerivative(
		address,
		validatorAddress,
		sdk.NewCoin("ufury", sdkmath.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"liquid/MsgMintDerivative","value":{"amount":{"amount":"1000000000","denom":"ufury"},"sender":"black1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgBurnDerivative_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("black1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgBurnDerivative(
		address,
		validatorAddress,
		sdk.NewCoin("bblack-blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42", sdkmath.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"liquid/MsgBurnDerivative","value":{"amount":{"amount":"1000000000","denom":"bblack-blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"},"sender":"black1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsg_Validate(t *testing.T) {
	validAddress := mustAccAddressFromBech32("black1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validValidatorAddress := mustValAddressFromBech32("blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")
	validCoin := sdk.NewInt64Coin("ufury", 1e9)

	type msgArgs struct {
		sender    string
		validator string
		amount    sdk.Coin
	}
	tests := []struct {
		name        string
		msgArgs     msgArgs
		expectedErr error
	}{
		{
			name: "normal is valid",
			msgArgs: msgArgs{
				sender:    validAddress.String(),
				validator: validValidatorAddress.String(),
				amount:    validCoin,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:    "invalid",
				validator: validValidatorAddress.String(),
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid short sender",
			msgArgs: msgArgs{
				sender:    "black1uexte6", // encoded zero length address
				validator: validValidatorAddress.String(),
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid validator",
			msgArgs: msgArgs{
				sender:    validAddress.String(),
				validator: "invalid",
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid nil coin",
			msgArgs: msgArgs{
				sender:    validAddress.String(),
				validator: validValidatorAddress.String(),
				amount:    sdk.Coin{},
			},
			expectedErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid zero coin",
			msgArgs: msgArgs{
				sender:    validAddress.String(),
				validator: validValidatorAddress.String(),
				amount:    sdk.NewInt64Coin("ufury", 0),
			},
			expectedErr: sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tc := range tests {
		msgs := []sdk.Msg{
			&types.MsgMintDerivative{
				Sender:    tc.msgArgs.sender,
				Validator: tc.msgArgs.validator,
				Amount:    tc.msgArgs.amount,
			},
			&types.MsgBurnDerivative{
				Sender:    tc.msgArgs.sender,
				Validator: tc.msgArgs.validator,
				Amount:    tc.msgArgs.amount,
			},
		}
		for _, msg := range msgs {
			t.Run(fmt.Sprintf("%s/%T", tc.name, msg), func(t *testing.T) {
				err := msg.ValidateBasic()
				if tc.expectedErr == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, tc.expectedErr, "expected error '%s' not found in actual '%s'", tc.expectedErr, err)
				}
			})
		}
	}
}

func mustAccAddressFromBech32(address string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

func mustValAddressFromBech32(address string) sdk.ValAddress {
	addr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}
