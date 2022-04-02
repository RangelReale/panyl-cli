package panylcli

import (
	"github.com/RangelReale/panyl"
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
	Name    string
	Enabled bool
}

func WithPluginOptions(pluginOptions []PluginOption) Option {
	return func(o *options) {
		o.pluginOptions = append(o.pluginOptions, pluginOptions...)
	}
}

type ProcessorProviderFunc func(preset string, pluginsEnabled []string, flags *pflag.FlagSet) (*panyl.Processor, error)

func WithProcessorlProvider(f ProcessorProviderFunc) Option {
	return func(o *options) {
		o.processorProvider = f
	}
}

type ResultProviderFunc func(flags *pflag.FlagSet) (panyl.ProcessResult, error)

func WithResultProvider(f ResultProviderFunc) Option {
	return func(o *options) {
		o.resultProvider = f
	}
}
