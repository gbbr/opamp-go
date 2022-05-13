package types

import (
	"context"

	"github.com/open-telemetry/opamp-go/protobufs"
)

type OwnTelemetryType int

const (
	OwnMetrics OwnTelemetryType = iota
	OwnTraces
	OwnLogs
)

type Callbacks interface {
	OnConnect()

	// OnConnectFailed is called when the connection to the Server cannot be established.
	// May be called after Start() is called and tries to connect to the Server.
	// May also be called if the connection is lost and reconnection attempt fails.
	OnConnectFailed(err error)

	// OnError is called when the Server reports an error in response to some previously
	// sent request. Useful for logging purposes. The Agent should not attempt to process
	// the error by reconnecting or retrying previous operations. The client handles the
	// ErrorResponse_UNAVAILABLE case internally by performing retries as necessary.
	OnError(err *protobufs.ServerErrorResponse)

	// OnRemoteConfig is called when the Agent receives a remote config from the Server.
	// The config parameter will not be nil.
	//
	// The Agent should process the config and return the effective config if processing
	// succeeded or an error if processing failed.
	//
	// configChanged must be set to true if as a result of applying the remote config
	// the effective config has changed.
	//
	// The returned effective config or the error will be reported back to the Server
	// via StatusReport message (using EffectiveConfig and RemoteConfigStatus fields).
	//
	// Only one OnRemoteConfig call can be active at any time. Until OnRemoteConfig
	// returns it will not be called again. Any other remote configs received from
	// the Server while OnRemoteConfig call has not returned will be remembered.
	// Once OnRemoteConfig call returns it will be called again with the most recent
	// remote config received.
	//
	// The EffectiveConfig.Hash field will be calculated and updated from the content
	// of the rest of the fields.
	OnRemoteConfig(
		ctx context.Context,
		remoteConfig *protobufs.AgentRemoteConfig,
	) (effectiveConfig *protobufs.EffectiveConfig, configChanged bool, err error)

	// SaveRemoteConfigStatus is called after OnRemoteConfig returns. The status
	// will be set either as APPLIED or FAILED depending on whether OnRemoteConfig
	// returned a success or error.
	// The Agent must remember this RemoteConfigStatus and supply in the future
	// calls to Start() in StartSettings.RemoteConfigStatus.
	SaveRemoteConfigStatus(ctx context.Context, status *protobufs.RemoteConfigStatus)

	// GetEffectiveConfig returns the current effective config. Only one
	// GetEffectiveConfig call can be active at any time. Until GetEffectiveConfig
	// returns it will not be called again.
	//
	// The Hash field in the returned EffectiveConfig will be calculated and updated
	// by the caller from the content of the rest of the fields.
	GetEffectiveConfig(ctx context.Context) (*protobufs.EffectiveConfig, error)

	// OnOpampConnectionSettings is called when the Agent receives an OpAMP
	// connection settings offer from the Server. Typically the settings can specify
	// authorization headers or TLS certificate, potentially also a different
	// OpAMP destination to work with.
	//
	// The Agent should process the offer and return an error if the Agent does not
	// want to accept the settings (e.g. if the TSL certificate in the settings
	// cannot be verified). The returned error will be reported back to the Server
	// via StatusReport message (using ConnectionStatuses field).
	//
	// If OnOpampConnectionSettings returns nil and then the caller will
	// attempt to reconnect to the OpAMP Server using the new settings.
	// If the connection fails the settings will be rejected and an error will
	// be reported to the Server. If the connection succeeds the new settings
	// will be used by the client from that moment on.
	//
	// Accepted or rejected settings will be reported back to the Server via
	// ConnectionStatuses message.
	//
	// Only one OnOpampConnectionSettings call can be active at any time.
	// See OnRemoteConfig for the behavior.
	OnOpampConnectionSettings(
		ctx context.Context,
		settings *protobufs.ConnectionSettings,
	) error

	// OnOpampConnectionSettingsAccepted will be called after the settings are
	// verified and accepted (OnOpampConnectionSettingsOffer and connection using
	// new settings succeeds). The Agent should store the settings and use them
	// in the future. Old connection settings should be forgotten.
	OnOpampConnectionSettingsAccepted(
		settings *protobufs.ConnectionSettings,
	)

	// OnOwnTelemetryConnectionSettings is called when the Agent receives a
	// connection settings to be used for reporting Agent's own telemetry.
	//
	// The Agent should process the settings and return an error if the Agent does not
	// want to accept the settings (e.g. if the TSL certificate in the settings
	// cannot be verified). The returned error will be reported back to the Server
	// via StatusReport message (using ConnectionStatuses field).
	// If the Agent accepts the settings it should return nil and begin sending
	// its own telemetry to the destination specified in the settings.
	// We currently support 3 types of Agent's own telemetry: metrics, traces, logs.
	// The Agent can support any subset of these types.
	OnOwnTelemetryConnectionSettings(
		ctx context.Context,
		telemetryType OwnTelemetryType,
		settings *protobufs.ConnectionSettings,
	) error

	// OnOtherConnectionSettings is called when the Agent receives a
	// connection settings to be used for by the Agent for any other purposes.
	// Typically these are used by the Agent to send collected data to the destinations
	// it is configured to. The name is typically the name of the destination.
	//
	// The Agent should process the settings and return an error if the Agent does not
	// want to accept the settings (e.g. if the TSL certificate in the settings
	// cannot be verified). The returned error will be reported back to the Server
	// via StatusReport message (using ConnectionStatuses field).
	OnOtherConnectionSettings(
		ctx context.Context,
		name string,
		certificate *protobufs.ConnectionSettings,
	) error

	// OnPackagesAvailable is called when the Server has packages available which are
	// different from what the Agent indicated it has via
	// LastServerProvidedAllPackagesHash.
	// syncer can be used to initiate syncing the packages from the Server.
	OnPackagesAvailable(ctx context.Context, packages *protobufs.PackagesAvailable, syncer PackagesSyncer) error

	// OnAgentIdentification is called when the Server requests changing identification of the Agent.
	// Agent should be updated with new id and use it for all further communication.
	// If Agent does not support the identification override from the Server, the function should return an error.
	OnAgentIdentification(ctx context.Context, agentId *protobufs.AgentIdentification) error

	// OnCommand is called when the Server requests that the connected Agent perform a command.
	OnCommand(command *protobufs.ServerToAgentCommand) error

	// For all methods that accept a context parameter the caller may cancel the
	// context if processing takes too long. In that case the method should return
	// as soon as possible with an error.
}

type CallbacksStruct struct {
	OnConnectFunc       func()
	OnConnectFailedFunc func(err error)
	OnErrorFunc         func(err *protobufs.ServerErrorResponse)

	OnRemoteConfigFunc func(
		ctx context.Context,
		remoteConfig *protobufs.AgentRemoteConfig,
	) (effectiveConfig *protobufs.EffectiveConfig, configChanged bool, err error)

	SaveRemoteConfigStatusFunc func(ctx context.Context, status *protobufs.RemoteConfigStatus)

	GetEffectiveConfigFunc func(ctx context.Context) (*protobufs.EffectiveConfig, error)

	OnOpampConnectionSettingsFunc func(
		ctx context.Context,
		settings *protobufs.ConnectionSettings,
	) error
	OnOpampConnectionSettingsAcceptedFunc func(
		settings *protobufs.ConnectionSettings,
	)

	OnOwnTelemetryConnectionSettingsFunc func(
		ctx context.Context,
		telemetryType OwnTelemetryType,
		settings *protobufs.ConnectionSettings,
	) error

	OnOtherConnectionSettingsFunc func(
		ctx context.Context,
		name string,
		settings *protobufs.ConnectionSettings,
	) error

	OnPackagesAvailableFunc   func(ctx context.Context, packages *protobufs.PackagesAvailable, syncer PackagesSyncer) error
	OnAgentIdentificationFunc func(ctx context.Context, agentId *protobufs.AgentIdentification) error

	OnCommandFunc func(command *protobufs.ServerToAgentCommand) error
}

var _ Callbacks = (*CallbacksStruct)(nil)

func (c CallbacksStruct) OnConnect() {
	if c.OnConnectFunc != nil {
		c.OnConnectFunc()
	}
}

func (c CallbacksStruct) OnConnectFailed(err error) {
	if c.OnConnectFailedFunc != nil {
		c.OnConnectFailedFunc(err)
	}
}

func (c CallbacksStruct) OnError(err *protobufs.ServerErrorResponse) {
	if c.OnErrorFunc != nil {
		c.OnErrorFunc(err)
	}
}

func (c CallbacksStruct) OnRemoteConfig(
	ctx context.Context,
	remoteConfig *protobufs.AgentRemoteConfig,
) (effectiveConfig *protobufs.EffectiveConfig, configChanged bool, err error) {
	if c.OnRemoteConfigFunc != nil {
		return c.OnRemoteConfigFunc(ctx, remoteConfig)
	}
	return nil, false, nil
}

func (c CallbacksStruct) SaveRemoteConfigStatus(ctx context.Context, status *protobufs.RemoteConfigStatus) {
	if c.SaveRemoteConfigStatusFunc != nil {
		c.SaveRemoteConfigStatusFunc(ctx, status)
	}
}

func (c CallbacksStruct) GetEffectiveConfig(ctx context.Context) (*protobufs.EffectiveConfig, error) {
	if c.GetEffectiveConfigFunc != nil {
		return c.GetEffectiveConfigFunc(ctx)
	}
	return nil, nil
}

func (c CallbacksStruct) OnOpampConnectionSettings(
	ctx context.Context, settings *protobufs.ConnectionSettings,
) error {
	if c.OnOpampConnectionSettingsFunc != nil {
		return c.OnOpampConnectionSettingsFunc(ctx, settings)
	}
	return nil
}

func (c CallbacksStruct) OnOpampConnectionSettingsAccepted(settings *protobufs.ConnectionSettings) {
	if c.OnOpampConnectionSettingsAcceptedFunc != nil {
		c.OnOpampConnectionSettingsAcceptedFunc(settings)
	}
}

func (c CallbacksStruct) OnOwnTelemetryConnectionSettings(
	ctx context.Context, telemetryType OwnTelemetryType,
	settings *protobufs.ConnectionSettings,
) error {
	if c.OnOwnTelemetryConnectionSettingsFunc != nil {
		return c.OnOwnTelemetryConnectionSettingsFunc(ctx, telemetryType, settings)
	}
	return nil
}

func (c CallbacksStruct) OnOtherConnectionSettings(
	ctx context.Context, name string, settings *protobufs.ConnectionSettings,
) error {
	if c.OnOtherConnectionSettingsFunc != nil {
		return c.OnOtherConnectionSettingsFunc(ctx, name, settings)
	}
	return nil
}

func (c CallbacksStruct) OnPackagesAvailable(
	ctx context.Context,
	packages *protobufs.PackagesAvailable,
	syncer PackagesSyncer,
) error {
	if c.OnPackagesAvailableFunc != nil {
		return c.OnPackagesAvailableFunc(ctx, packages, syncer)
	}
	return nil
}

func (c CallbacksStruct) OnCommand(command *protobufs.ServerToAgentCommand) error {
	if c.OnCommandFunc != nil {
		return c.OnCommandFunc(command)
	}
	return nil
}

func (c CallbacksStruct) OnAgentIdentification(
	ctx context.Context,
	agentId *protobufs.AgentIdentification,
) error {
	if c.OnAgentIdentificationFunc != nil {
		return c.OnAgentIdentificationFunc(ctx, agentId)
	}
	return nil
}
