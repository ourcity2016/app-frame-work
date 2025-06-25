package util

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

func JSONRPCWithCtx(obj interface{}, methodName string, ctx context.Context, jsonParams string) (interface{}, error) {
	// 获取方法
	method := reflect.ValueOf(obj).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	methodType := method.Type()
	if methodType.NumIn() < 1 || methodType.NumIn() > 2 {
		return nil, fmt.Errorf("method %s must have 1 or 2 parameters", methodName)
	}

	// 准备调用参数
	var callArgs []reflect.Value
	var paramType reflect.Type
	argOffset := 0

	// 检查第一个参数是否是 context.Context
	if methodType.NumIn() > 1 && methodType.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
		argOffset = 1
	}

	// 获取参数类型
	paramType = methodType.In(argOffset)
	if paramType.Kind() != reflect.Ptr && paramType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("method %s parameter must be a pointer or struct", methodName)
	}

	// 创建参数实例
	var paramValue reflect.Value
	if paramType.Kind() == reflect.Ptr {
		// 指针类型: 创建指向新实例的指针
		paramValue = reflect.New(paramType.Elem())
	} else {
		// 值类型: 创建新实例
		paramValue = reflect.New(paramType)
	}

	// 解析 JSON 到参数
	if err := json.Unmarshal([]byte(jsonParams), paramValue.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// 如果是值类型，获取指向的值
	if paramType.Kind() != reflect.Ptr {
		paramValue = paramValue.Elem()
	}

	callArgs = append(callArgs, paramValue)

	// 验证所有参数
	for i, arg := range callArgs {
		if !arg.IsValid() {
			return nil, fmt.Errorf("invalid argument at position %d", i)
		}
		//if methodType.In(i) != arg.Type() {
		//	return nil, fmt.Errorf("argument type mismatch at position %d: expected %v, got %v",
		//		i, methodType.In(i), arg.Type())
		//}
	}

	// 调用方法
	results := method.Call(callArgs)
	if results == nil {
		return nil, nil
	}
	// 处理返回值
	switch len(results) {
	case 2:
		// (result, error) 签名
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return results[0].Interface(), err
	case 1:
		// (error) 或 (result) 签名
		if methodType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !results[0].IsNil() {
				return nil, results[0].Interface().(error)
			}
			return nil, nil
		}
		return results[0].Interface(), nil
	default:
		return nil, fmt.Errorf("unsupported return signature for method %s", methodName)
	}
}
func JSONRPC(obj interface{}, methodName string, jsonParams string) (interface{}, error) {
	// 获取服务实例
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	rpcService := obj

	// 获取方法类型信息以确定参数类型
	method := reflect.ValueOf(rpcService).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	// 创建参数实例
	paramType := method.Type().In(0)
	param := reflect.New(paramType).Interface()

	// 解析 JSON
	if err := json.Unmarshal([]byte(jsonParams), param); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// 调用方法
	results := method.Call([]reflect.Value{reflect.ValueOf(param).Elem()})
	return results, nil
}
func DynamicInvoke(obj interface{}, methodName string, params ...interface{}) (interface{}, error) {
	objValue := reflect.ValueOf(obj)

	// 验证是否是结构体指针
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("必须传入结构体指针")
	}

	method := objValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("方法不存在")
	}

	// 准备参数
	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}

	// 调用方法
	out := method.Call(in)

	// 处理返回值
	if len(out) == 0 {
		return nil, nil
	}

	// 只返回第一个值
	return out[0], nil
}

func GetTypeFullName(v interface{}) string {
	t := reflect.TypeOf(v)

	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 获取包路径和类型名
	pkgPath := t.PkgPath()
	typeName := t.Name()

	if pkgPath == "" {
		return typeName // 内置类型
	}
	return pkgPath + "." + typeName
}
