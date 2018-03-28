package builtin

import (
	"errors"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/go-funk"
)

type builtinConfig struct {
	oConnFunc core.CFunc
	oFunc     core.OFunc
	oDef      core.OperatorDef
}

var cfgs map[string]*builtinConfig

func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	cfg := getBuiltinCfg(def.Operator)

	if cfg == nil {
		return nil, errors.New("unknown builtin operator")
	}

	for prop, propType := range cfg.oDef.PropertyDefs {
		if err := propType.VerifyData(def.Properties[prop]); err != nil {
			return nil, err
		}
	}

	srvs := make(map[string]*core.ServiceDef)
	for srvName, srv := range cfg.oDef.ServiceDefs {
		srvCpy := srv.Copy()
		srvs[srvName] = &srvCpy
	}

	dels := make(map[string]*core.DelegateDef)
	for delName, del := range cfg.oDef.DelegateDefs {
		delCpy := del.Copy()
		dels[delName] = &delCpy
	}

	o, err := core.NewOperator(def.Name, cfg.oFunc, cfg.oConnFunc, def.Generics, def.Properties, srvs, dels)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func GetOperatorDef(insDef *core.InstanceDef) (core.OperatorDef, error) {
	cfg, ok := cfgs[insDef.Operator]
	oDef := cfg.oDef

	if !ok {
		return oDef, errors.New("builtin operator not found")
	}

	// We must not change oDef in any way as this would affect other instances of this builtin operator
	if err := oDef.SpecifyGenericPorts(insDef.Generics); err != nil {
		return oDef, err
	}

	return oDef, nil
}

func IsRegistered(name string) bool {
	_, b := cfgs[name]
	return b
}

func Register(name string, cfg *builtinConfig) {
	cfgs[name] = cfg
}

func GetBuiltinNames() []string {
	return funk.Keys(cfgs).([]string)
}

func init() {
	cfgs = make(map[string]*builtinConfig)
	Register("slang.const", constOpCfg)
	Register("slang.eval", evalOpCfg)

	Register("slang.fork", forkOpCfg)
	Register("slang.syncFork", syncForkOpCfg)
	Register("slang.merge", mergeOpCfg)
	Register("slang.syncMerge", syncMergeOpCfg)
	Register("slang.take", takeOpCfg)

	Register("slang.loop", loopOpCfg)
	Register("slang.aggregate", aggregateOpCfg)
	Register("slang.reduce", reduceOpCfg)

	Register("slang.window.count", windowCountOpCfg)
	Register("slang.window.triggered", windowTriggeredOpCfg)

	Register("slang.net.httpServer", httpServerOpCfg)

	Register("slang.files.read", fileReadOpCfg)

	Register("slang.encoding.csv.read", csvReadOpCfg)
	Register("slang.encoding.xlsx.read", xlsxReadOpCfg)

	Register("slang.time.delay", delayOpCfg)
}

func getBuiltinCfg(name string) *builtinConfig {
	c, _ := cfgs[name]
	return c
}
