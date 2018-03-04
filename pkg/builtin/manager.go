package builtin

import (
	"errors"
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
	"github.com/Bitspark/go-funk"
)

type PropertyFunc func(*core.Operator, map[string]interface{}) error

type builtinConfig struct {
	oPropFunc PropertyFunc
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

	dels := make(map[string]*core.DelegateDef)

	in := cfg.oDef.In.Copy()
	out := cfg.oDef.Out.Copy()
	for delName, del := range cfg.oDef.Delegates {
		delCpy := del.Copy()
		dels[delName] = &delCpy
	}

	if err := in.SpecifyGenericPorts(def.Generics); err != nil {
		return nil, err
	}
	if err := out.SpecifyGenericPorts(def.Generics); err != nil {
		return nil, err
	}
	for _, del := range dels {
		if err := del.Out.SpecifyGenericPorts(def.Generics); err != nil {
			return nil, err
		}
		if err := del.In.SpecifyGenericPorts(def.Generics); err != nil {
			return nil, err
		}
	}

	if err := in.GenericsSpecified(); err != nil {
		return nil, err
	}
	if err := out.GenericsSpecified(); err != nil {
		return nil, err
	}
	for delName, del := range dels {
		if err := del.Out.GenericsSpecified(); err != nil {
			return nil, fmt.Errorf("%s: %s", delName, err.Error())
		}
		if err := del.In.GenericsSpecified(); err != nil {
			return nil, fmt.Errorf("%s: %s", delName, err.Error())
		}
	}

	o, err := core.NewOperator(def.Name, cfg.oFunc, cfg.oConnFunc, in, out, dels)
	if err != nil {
		return nil, err
	}

	if cfg.oPropFunc != nil {
		err = cfg.oPropFunc(o, def.Properties)
		if err != nil {
			return nil, err
		}
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

	Register("slang.net.httpServer", httpServerOpCfg)

	Register("slang.files.read", fileReadOpCfg)
}

func getBuiltinCfg(name string) *builtinConfig {
	c, _ := cfgs[name]
	return c
}