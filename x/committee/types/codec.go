package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	communitytypes "github.com/zeta-protocol/black/x/community/types"
	blackdisttypes "github.com/zeta-protocol/black/x/blackdist/types"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/committee module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/committee and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	// amino is not sealed so that other modules can register their own pubproposal and/or permission types.

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)

	// CommitteeChange/Delete proposals along with Permission types are
	// registered on gov's ModuleCdc
	RegisterLegacyAminoCodec(govv1beta1.ModuleCdc.LegacyAmino)

	// Register external module pubproposal types. Ideally these would be registered within the modules' types pkg init function.
	// However registration happens here as a work-around.
	RegisterProposalTypeCodec(distrtypes.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal")
	RegisterProposalTypeCodec(proposaltypes.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal")
	RegisterProposalTypeCodec(govv1beta1.TextProposal{}, "cosmos-sdk/TextProposal")
	RegisterProposalTypeCodec(upgradetypes.SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal")
	RegisterProposalTypeCodec(upgradetypes.CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal")
	RegisterProposalTypeCodec(communitytypes.CommunityCDPRepayDebtProposal{}, "black/CommunityCDPRepayDebtProposal")
	RegisterProposalTypeCodec(communitytypes.CommunityCDPWithdrawCollateralProposal{}, "black/CommunityCDPWithdrawCollateralProposal")
	RegisterProposalTypeCodec(communitytypes.CommunityPoolLendWithdrawProposal{}, "black/CommunityPoolLendWithdrawProposal")
	RegisterProposalTypeCodec(blackdisttypes.CommunityPoolMultiSpendProposal{}, "black/CommunityPoolMultiSpendProposal")
}

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Proposals
	cdc.RegisterInterface((*PubProposal)(nil), nil)
	cdc.RegisterConcrete(CommitteeChangeProposal{}, "black/CommitteeChangeProposal", nil)
	cdc.RegisterConcrete(CommitteeDeleteProposal{}, "black/CommitteeDeleteProposal", nil)

	// Committees
	cdc.RegisterInterface((*Committee)(nil), nil)
	cdc.RegisterConcrete(BaseCommittee{}, "black/BaseCommittee", nil)
	cdc.RegisterConcrete(MemberCommittee{}, "black/MemberCommittee", nil)
	cdc.RegisterConcrete(TokenCommittee{}, "black/TokenCommittee", nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "black/GodPermission", nil)
	cdc.RegisterConcrete(TextPermission{}, "black/TextPermission", nil)
	cdc.RegisterConcrete(SoftwareUpgradePermission{}, "black/SoftwareUpgradePermission", nil)
	cdc.RegisterConcrete(ParamsChangePermission{}, "black/ParamsChangePermission", nil)
	cdc.RegisterConcrete(CommunityCDPRepayDebtPermission{}, "black/CommunityCDPRepayDebtPermission", nil)
	cdc.RegisterConcrete(CommunityCDPWithdrawCollateralPermission{}, "black/CommunityCDPWithdrawCollateralPermission", nil)
	cdc.RegisterConcrete(CommunityPoolLendWithdrawPermission{}, "black/CommunityPoolLendWithdrawPermission", nil)

	// Msgs
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "black/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "black/MsgVote")
}

// RegisterProposalTypeCodec allows external modules to register their own pubproposal types on the
// internal ModuleCdc. This allows the MsgSubmitProposal to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	registry.RegisterInterface(
		"black.committee.v1beta1.Committee",
		(*Committee)(nil),
		&BaseCommittee{},
		&TokenCommittee{},
		&MemberCommittee{},
	)

	registry.RegisterInterface(
		"black.committee.v1beta1.Permission",
		(*Permission)(nil),
		&GodPermission{},
		&TextPermission{},
		&SoftwareUpgradePermission{},
		&ParamsChangePermission{},
		&CommunityCDPRepayDebtPermission{},
		&CommunityCDPWithdrawCollateralPermission{},
		&CommunityPoolLendWithdrawPermission{},
	)

	// Need to register PubProposal here since we use this as alias for the x/gov Content interface for all the proposal implementations used in this module.
	// Note that all proposals supported by x/committee needed to be registered here, including the proposals from x/gov.
	registry.RegisterInterface(
		"black.committee.v1beta1.PubProposal",
		(*PubProposal)(nil),
		&Proposal{},
		&distrtypes.CommunityPoolSpendProposal{},
		&govv1beta1.TextProposal{},
		&blackdisttypes.CommunityPoolMultiSpendProposal{},
		&proposaltypes.ParameterChangeProposal{},
		&upgradetypes.SoftwareUpgradeProposal{},
		&upgradetypes.CancelSoftwareUpgradeProposal{},
		&communitytypes.CommunityCDPRepayDebtProposal{},
		&communitytypes.CommunityCDPWithdrawCollateralProposal{},
		&communitytypes.CommunityPoolLendWithdrawProposal{},
	)

	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&CommitteeChangeProposal{},
		&CommitteeDeleteProposal{},
	)
}
