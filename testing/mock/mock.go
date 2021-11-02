package mock

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v2/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	"github.com/cosmos/ibc-go/v2/modules/core/exported"
)

const (
	ModuleName = "mock"

	Version = "mock-version"
)

var (
	MockAcknowledgement             = channeltypes.NewResultAcknowledgement([]byte("mock acknowledgement"))
	MockFailAcknowledgement         = channeltypes.NewErrorAcknowledgement("mock failed acknowledgement")
	MockPacketData                  = []byte("mock packet data")
	MockFailPacketData              = []byte("mock failed packet data")
	MockAsyncPacketData             = []byte("mock async packet data")
	MockRecvCanaryCapabilityName    = "mock receive canary capability name"
	MockAckCanaryCapabilityName     = "mock acknowledgement canary capability name"
	MockTimeoutCanaryCapabilityName = "mock timeout canary capability name"
)

var _ porttypes.IBCModule = AppModule{}

// Expected Interface
// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// AppModuleBasic is the mock AppModuleBasic.
type AppModuleBasic struct{}

// Name implements AppModuleBasic interface.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec implements AppModuleBasic interface.
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces implements AppModuleBasic interface.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// DefaultGenesis implements AppModuleBasic interface.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis implements the AppModuleBasic interface.
func (AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes implements AppModuleBasic interface.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes implements AppModuleBasic interface.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// GetTxCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule represents the AppModule for the mock module.
type AppModule struct {
	AppModuleBasic
	scopedKeeper capabilitykeeper.ScopedKeeper
	portKeeper   PortKeeper
	IBCApp       MockIBCApp // base application of an IBC middleware stack
}

// NewAppModule returns a mock AppModule instance.
func NewAppModule(sk capabilitykeeper.ScopedKeeper, pk PortKeeper) AppModule {
	return AppModule{
		scopedKeeper: sk,
		portKeeper:   pk,
	}
}

// RegisterInvariants implements the AppModule interface.
func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route implements the AppModule interface.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(ModuleName, nil)
}

// QuerierRoute implements the AppModule interface.
func (AppModule) QuerierRoute() string {
	return ""
}

// LegacyQuerierHandler implements the AppModule interface.
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices implements the AppModule interface.
func (am AppModule) RegisterServices(module.Configurator) {}

// InitGenesis implements the AppModule interface.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	// bind mock port ID
	cap := am.portKeeper.BindPort(ctx, ModuleName)
	am.scopedKeeper.ClaimCapability(ctx, cap, host.PortPath(ModuleName))

	return []abci.ValidatorUpdate{}
}

// ExportGenesis implements the AppModule interface.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock implements the AppModule interface
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
}

// EndBlock implements the AppModule interface
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// OnChanOpenInit implements the IBCModule interface.
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string,
	channelID string, chanCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string,
) error {
	if am.IBCApp.OnChanOpenInit != nil {
		return am.IBCApp.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)

	}

	// Claim channel capability passed back by IBC module
	if err := am.scopedKeeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	return nil
}

// OnChanOpenTry implements the IBCModule interface.
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string,
	channelID string, chanCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version, counterpartyVersion string,
) error {
	if am.IBCApp.OnChanOpenTry != nil {
		return am.IBCApp.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version, counterpartyVersion)

	}
	// Claim channel capability passed back by IBC module
	if err := am.scopedKeeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	return nil
}

// OnChanOpenAck implements the IBCModule interface.
func (am AppModule) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyVersion string) error {
	if am.IBCApp.OnChanOpenAck != nil {
		return am.IBCApp.OnChanOpenAck(ctx, portID, channelID, counterpartyVersion)
	}

	return nil
}

// OnChanOpenConfirm implements the IBCModule interface.
func (am AppModule) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	if am.IBCApp.OnChanOpenConfirm != nil {
		return am.IBCApp.OnChanOpenConfirm(ctx, portID, channelID)
	}

	return nil
}

// OnChanCloseInit implements the IBCModule interface.
func (am AppModule) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	if am.IBCApp.OnChanCloseInit != nil {
		return am.IBCApp.OnChanCloseInit(ctx, portID, channelID)
	}

	return nil
}

// OnChanCloseConfirm implements the IBCModule interface.
func (am AppModule) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	if am.IBCApp.OnChanCloseConfirm != nil {
		return am.IBCApp.OnChanCloseConfirm(ctx, portID, channelID)
	}

	return nil
}

// OnRecvPacket implements the IBCModule interface.
func (am AppModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	if am.IBCApp.OnRecvPacket != nil {
		return am.IBCApp.OnRecvPacket(ctx, packet, relayer)
	}

	// set state by claiming capability to check if revert happens return
	_, err := am.scopedKeeper.NewCapability(ctx, MockRecvCanaryCapabilityName+strconv.Itoa(int(packet.GetSequence())))
	if err != nil {
		// application callback called twice on same packet sequence
		// must never occur
		panic(err)
	}
	if bytes.Equal(MockPacketData, packet.GetData()) {
		return MockAcknowledgement
	} else if bytes.Equal(MockAsyncPacketData, packet.GetData()) {
		return nil
	}

	return MockFailAcknowledgement
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (am AppModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	if am.IBCApp.OnAcknowledgementPacket != nil {
		return am.IBCApp.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	_, err := am.scopedKeeper.NewCapability(ctx, MockAckCanaryCapabilityName+strconv.Itoa(int(packet.GetSequence())))
	if err != nil {
		// application callback called twice on same packet sequence
		// must never occur
		panic(err)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface.
func (am AppModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	if am.IBCApp.OnTimeoutPacket != nil {
		return am.IBCApp.OnTimeoutPacket(ctx, packet, relayer)
	}

	_, err := am.scopedKeeper.NewCapability(ctx, MockTimeoutCanaryCapabilityName+strconv.Itoa(int(packet.GetSequence())))
	if err != nil {
		// application callback called twice on same packet sequence
		// must never occur
		panic(err)
	}

	return nil
}

// NegotiateAppVersion implements the IBCModule interface.
func (am AppModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (string, error) {
	if am.IBCApp.NegotiateAppVersion != nil {
		return am.IBCApp.NegotiateAppVersion(ctx, order, connectionID, portID, counterparty, proposedVersion)
	}

	if proposedVersion != Version { // allow testing of error scenarios
		return "", errors.New("failed to negotiate app version")
	}

	return Version, nil
}