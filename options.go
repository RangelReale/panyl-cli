package panylcli

import (
	"context"

	"github.com/RangelReale/panyl/v2"
	"github.com/spf13/pflag"
)

type options struct {
	globalFlags       FlagsProvider
	presetFlags       FlagsProvider
	logFlags          FlagsProvider
	pluginOptions     []PluginOption
	processorProvider ProcessorProviderFunc
	resultProvider    ResultProviderFunc
}

type Option func(*options)

type FlagsProvider func(flags *pflag.FlagSet)

func WithDeclareGlobalFlags(f FlagsProvider) Option {
	return func(o *options) {
		o.globalFlags = f
	}
}

func WithDeclarePresetFlags(f FlagsProvider) Option {
	return func(o *options) {
		o.presetFlags = f
	}
}

func WithDeclareLogFlags(f FlagsProvider) Option {
	return func(o *options) {
		o.logFlags = f
	}
}

type PluginOption struct {
	Name          string
	Enabled       bool
	Preset        bool
	PresetEnabled bool
}

func WithPluginOptions(pluginOptions []PluginOption) Option {
	return func(o *options) {
		o.pluginOptions = append(o.pluginOptions, pluginOptions...)
	}
}

type ProcessorProviderFunc func(ctx context.Context, preset string, pluginsEnabled []string,
	flags *pflag.FlagSet) (context.Context, *panyl.Processor, []panyl.JobOption, error)

func WithProcessorProvider(f ProcessorProviderFunc) Option {
	return func(o *options) {
		o.processorProvider = f
	}
}

type ResultProviderFunc func(ctx context.Context, flags *pflag.FlagSet) (panyl.ProcessResult, error)

func WithResultProvider(f ResultProviderFunc) Option {
	return func(o *options) {
		o.resultProvider = f
	}
}
