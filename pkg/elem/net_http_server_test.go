package elem

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_HTTP__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(netHTTPServerId)
	a.NotNil(ocFork)
}

func Test_HTTP__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().In())
	a.Equal(core.TYPE_NUMBER, o.Main().In().Type())
}

func Test_HTTP__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().Out())
	a.Equal(core.TYPE_STRING, o.Main().Out().Type())
}

func Test_HTTP__Delegates(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	dlg := o.Delegate("handler")
	a.NotNil(dlg)

	a.Equal(core.TYPE_MAP, dlg.In().Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Type())

	a.Equal(core.TYPE_BINARY, dlg.In().Map("body").Type())
	a.Equal(core.TYPE_NUMBER, dlg.In().Map("status").Type())
	a.Equal(core.TYPE_STREAM, dlg.In().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.In().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.In().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STRING, dlg.In().Map("headers").Stream().Map("value").Type())

	a.Equal(core.TYPE_STRING, dlg.Out().Map("method").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("path").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("headers").Stream().Map("values").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("headers").Stream().Map("values").Stream().Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Map("params").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("params").Stream().Map("key").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("params").Stream().Map("values").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("params").Stream().Map("values").Stream().Type())
}

func Test_HTTP__Request(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.Main().In().Push(9438)

	done := false

	go func() {
		for i := 0; i < 5; i++ {
			http.Get("http://127.0.0.1:9438/test123?a=1")
			if done {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()

	a.Equal("GET", handler.Out().Map("method").Pull())
	a.Equal("/test123", handler.Out().Map("path").Pull())
	a.Equal([]interface{}{map[string]interface{}{"key": "a", "values": []interface{}{"1"}}}, handler.Out().Map("params").Pull())
	done = true
}

func Test_HTTP__Response200(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.Main().In().Push(9439)
	handler.In().Push(map[string]interface{}{"status": 200, "headers": []interface{}{}, "body": core.Binary("hallo slang!")})

	for i := 0; i < 5; i++ {
		resp, _ := http.Get("http://127.0.0.1:9439/test789")
		if resp == nil || resp.StatusCode != 200 {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		a.Equal([]byte("hallo slang!"), buf.Bytes())
		a.Equal("200 OK", resp.Status)
		return
	}
	a.Fail("no response")
}

func Test_HTTP__Response404(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.Main().In().Push(9440)
	handler.In().Push(map[string]interface{}{"status": 404, "headers": []interface{}{}, "body": core.Binary("bye slang!")})

	for i := 0; i < 5; i++ {
		resp, _ := http.Get("http://127.0.0.1:9440/test789")
		if resp == nil || resp.StatusCode != 404 {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		a.Equal([]byte("bye slang!"), buf.Bytes())
		a.Equal("404 Not Found", resp.Status)
		return
	}
	a.Fail("no response")
}
