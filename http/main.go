package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetNewRootContext(newRootContext)
	proxywasm.SetNewHttpContext(newHttpContext)
}

var (
	counter proxywasm.MetricCounter
)
const (
	metricsName = "proxy_wasm_go.request_counter"
)

type (
	wasmFilterDemoRootContext struct {
		// you must embed the default context so that you need not to reimplement all the methods by yourself
		proxywasm.DefaultRootContext
	}

	wasmFilterDemoHttpContext struct {
		// you must embed the default context so that you need not to reimplement all the methods by yourself
		proxywasm.DefaultHttpContext
		contextID uint32
	}
)

func newRootContext(contextID uint32) proxywasm.RootContext {
	return &wasmFilterDemoRootContext{}
}

func newHttpContext(rootContextID, contextID uint32) proxywasm.HttpContext {
	return &wasmFilterDemoHttpContext{}
}

const sharedDataKey = "shared_data_key"

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

// override
func (ctx *wasmFilterDemoHttpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	prev := counter.Get()
	proxywasm.LogInfof("previous value of %s: %d", metricsName, prev)

	counter.Increment(1)
	proxywasm.LogInfo("incremented")

	uuid, err := proxywasm.GetHttpRequestHeader("X-Request-Id")
	if err := proxywasm.SetSharedData(sharedDataKey, []byte(uuid), 0); err != nil {
		proxywasm.LogWarnf("error setting shared data on OnHttpRequestHeaders: %v", err)
		return types.ActionContinue
	}

	hs, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
	}

	for _, h := range hs {
		proxywasm.LogInfof("request header --> %s: %s", h[0], h[1])
	}
	return types.ActionContinue
}

// override
func (ctx *wasmFilterDemoHttpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	hs, err := proxywasm.GetHttpResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
	}
	value, _, err := proxywasm.GetSharedData(sharedDataKey)
	if err != nil {
		proxywasm.LogWarnf("error getting shared data on OnHttpRequestHeaders: %v", err)
		return types.ActionContinue
	}
	err = proxywasm.SetHttpResponseHeader("UUID", string(value))
	if err != nil {
		proxywasm.LogErrorf("failed to set request header: %v", err)
		return types.ActionContinue
	}
	for _, h := range hs {
		proxywasm.LogInfof("response header <-- %s: %s", h[0], h[1])
	}
	return types.ActionContinue
}

// override
func (ctx *wasmFilterDemoHttpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	proxywasm.LogInfof("body size: %d", bodySize)
	if bodySize != 0 {
		initialBody, err := proxywasm.GetHttpRequestBody(0, bodySize)
		if err != nil {
			proxywasm.LogErrorf("failed to get request body: %v", err)
			return types.ActionContinue
		}
		proxywasm.LogInfof("initial request body: %s", string(initialBody))

		b := []byte(`{ "another": "body" }`)

		err = proxywasm.SetHttpRequestBody(b)
		if err != nil {
			proxywasm.LogErrorf("failed to set request body: %v", err)
			return types.ActionContinue
		}

		proxywasm.LogInfof("on http request body finished")
	}

	return types.ActionContinue
}

// override
func (ctx *wasmFilterDemoHttpContext) OnHttpStreamDone() {
	proxywasm.LogInfof("%d finished", ctx.contextID)
}