package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetNewRootContext(newRootContext)
	proxywasm.SetNewStreamContext(newStreamContext)
}

var (
	counter proxywasm.MetricCounter
)
const (
	metricsName = "proxy_wasm_go.connection_counter"
)

type (
	wasmFilterDemoRootContext struct {
		// you must embed the default context so that you need not to reimplement all the methods by yourself
		proxywasm.DefaultRootContext
	}

	wasmFilterDemoStreamContext struct {
		// you must embed the default context so that you need not to reimplement all the methods by yourself
		proxywasm.DefaultStreamContext
		contextID uint32
	}
)

func newRootContext(contextID uint32) proxywasm.RootContext {
	return &wasmFilterDemoRootContext{}
}

func newStreamContext(rootContextID, contextID uint32) proxywasm.StreamContext {
	return &wasmFilterDemoStreamContext{}
}

// override
func (ctx *wasmFilterDemoRootContext) OnVMStart(vmConfigurationSize int) bool {
	data, err := proxywasm.GetVMConfiguration(vmConfigurationSize)
	if err != nil {
		proxywasm.LogCriticalf("error reading vm configuration: %v", err)
	}
	proxywasm.LogInfof("vm config: %s\n", string(data))
	counter = proxywasm.DefineCounterMetric(metricsName)
	return true
}

func (ctx wasmFilterDemoRootContext) OnPluginStart(pluginConfigurationSize int) bool {
	data, err := proxywasm.GetPluginConfiguration(pluginConfigurationSize)
	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
	}

	proxywasm.LogInfof("plugin config: %s\n", string(data))
	return true
}

func (ctx *wasmFilterDemoStreamContext) OnNewConnection() types.Action {
	proxywasm.LogInfo("new connection!")
	return types.ActionContinue
}

func (ctx *wasmFilterDemoStreamContext) OnDownstreamData(dataSize int, endOfStream bool) types.Action {
	if dataSize == 0 {
		return types.ActionContinue
	}

	data, err := proxywasm.GetDownStreamData(0, dataSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("failed to get downstream data: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogInfof(">>>>>> downstream data received >>>>>>\n%s", string(data))
	return types.ActionContinue
}

func (ctx *wasmFilterDemoStreamContext) OnDownstreamClose(types.PeerType) {
	proxywasm.LogInfo("downstream connection close!")
	return
}

func (ctx *wasmFilterDemoStreamContext) OnUpstreamData(dataSize int, endOfStream bool) types.Action {
	if dataSize == 0 {
		return types.ActionContinue
	}

	ret, err := proxywasm.GetProperty([]string{"upstream", "address"})
	if err != nil {
		proxywasm.LogCriticalf("failed to get downstream data: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogInfof("remote address: %s", string(ret))

	data, err := proxywasm.GetUpstreamData(0, dataSize)
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCritical(err.Error())
	}

	proxywasm.LogInfof("<<<<<< upstream data received <<<<<<\n%s", string(data))
	return types.ActionContinue
}

func (ctx *wasmFilterDemoStreamContext) OnStreamDone() {
	counter.Increment(1)
	proxywasm.LogInfo("connection complete!")
}
