package main

import (
	"fmt"
	"reflect"
	"strings"
)

type Base struct {
	OrganizationId int32
	regionId       int
	shortIds       []string
	UUIDs          []string
	DataSet        interface{}
	Status         map[string]interface{}
}

func (this *Base) _setError(errorCode int, errorMessage string) {
	this.Status["code"] = errorCode
	this.Status["message"] = errorMessage
}

func (this *Base) _hasError() bool {
	if this.Status["code"].(int) == 0 {
		return false
	}
	return true
}

func (this *Base) Init() *Base {
	if this.shortIds == nil {
		this.shortIds = make([]string, 0)
	}
	if this.UUIDs == nil {
		this.UUIDs = make([]string, 0)
	}
	if this.DataSet == nil {
		this.DataSet = make(map[string]interface{})
	}
	this.Status = make(map[string]interface{})
	this.Status["code"] = 0
	this.Status["message"] = ""
	return this
}

func (this *Base) Done() bool {
	return this.Status["code"].(int) == 0
}

func (this *Base) _getPath(nodePath string) (interface{}, interface{}) {
	nodes := strings.Split(nodePath, ".")
	key := nodes[len(nodes)-1 : len(nodes)][0]
	nodes = nodes[0 : len(nodes)-1]

	valueCurNode := reflect.ValueOf(this.DataSet)
	for _, nodeName := range nodes {
		if valueCurNode.Kind() == reflect.Invalid {
			// TODO maybe need to set error info
			return nil, ""
		}
		if valueCurNode.Kind() != reflect.Map {
			// TODO maybe need to set error info
			return nil, ""
		}
		valueNodeName := reflect.ValueOf(nodeName)
		// CHECK 这一步为空会怎样
		valueCurNode = valueCurNode.MapIndex(valueNodeName)
	}
	return valueCurNode.Interface(), key
}

func (this *Base) _set(nodePath string, data interface{}) {
	if this._hasError() {
		return
	}

	visitNode, visitKey := this._getPath(nodePath)
	valueVisitNode := reflect.ValueOf(visitNode)
	valueVisitKey := reflect.ValueOf(visitKey)
	// CHECK  这一步可能出现的值 可能不只nil
	if visitNode == nil {
		// TODO set error info
		return
	}

	valueLeafVisitNodeNew := reflect.ValueOf(data)
	valueVisitNode.SetMapIndex(valueVisitKey, valueLeafVisitNodeNew)
	return
}

func (this *Base) _get(nodePath string, data interface{}) interface{} {
	if this._hasError() {
		return nil
	}

	visitNode, visitKey := this._getPath(nodePath)
	valueVisitNode := reflect.ValueOf(visitNode)
	valueVisitKey := reflect.ValueOf(visitKey)
	// CHECK  这一步可能出现的值 可能不只nil
	if visitNode == nil {
		// TODO set error info
		return nil
	}

	valueLeafVisitNode := valueVisitNode.MapIndex(valueVisitKey)
	return valueLeafVisitNode.Interface()
}

func (this *Base) Filter(nodePath string, filters interface{}) *Base {
	if this._hasError() {
		return this
	}

	mapFilter := make(map[interface{}]int)
	valueFilters := reflect.ValueOf(filters)
	for i := 0; i < valueFilters.Len(); i++ {
		mapFilter[valueFilters.Index(i).Interface()] = 1
	}

	visitNode, visitKey := this._getPath(nodePath)
	valueVisitNode := reflect.ValueOf(visitNode)
	valueVisitKey := reflect.ValueOf(visitKey)
	valueLeafVisitNode := valueVisitNode.MapIndex(valueVisitKey)
	// CHECK  这一步可能出现的值 可能不只nil
	if visitNode == nil {
		// TODO set error info
		return this
	}

	if valueLeafVisitNode.Kind() == reflect.Interface {
		valueLeafVisitNode = valueLeafVisitNode.Elem()
	}
	if valueLeafVisitNode.Kind() == reflect.Slice {
		valueLeafVisitNodeNew := reflect.MakeSlice(valueLeafVisitNode.Type(), 0, 0)
		for i := 0; i < valueLeafVisitNode.Len(); i++ {
			if mapFilter[valueLeafVisitNode.Index(i).Interface()] == 1 {
				valueLeafVisitNodeNew = reflect.Append(valueLeafVisitNodeNew, valueLeafVisitNode.Index(i))
			}
		}
		valueVisitNode.SetMapIndex(valueVisitKey, valueLeafVisitNodeNew)
		return this
	}
	if valueLeafVisitNode.Kind() == reflect.Map {
		for _, valueKey := range valueLeafVisitNode.MapKeys() {
			key := valueKey.Interface()
			if mapFilter[key] == 0 {
				valueLeafVisitNode.SetMapIndex(valueKey, reflect.ValueOf(nil))
			}
		}
		// CHECK 这一步是否可省略
		valueVisitNode.SetMapIndex(valueVisitKey, valueLeafVisitNode)
		return this
	}

	// TODO set error info
	return this
}

func (this *Base) GroupByKey(nodePath string, selectedKey string) *Base {
	if this._hasError() {
		return this
	}

	//if len(args) != 2 {
	//// TODO set error info
	//return this
	//}
	//F = function.(func(...interface{}))(args)

	visitNode, visitKey := this._getPath(nodePath)
	valueVisitNode := reflect.ValueOf(visitNode)
	valueVisitKey := reflect.ValueOf(visitKey)
	valueLeafNodeItems := valueVisitNode.MapIndex(valueVisitKey)
	valueSelectedKey := reflect.ValueOf(selectedKey)
	// CHECK  这一步可能出现的值 可能不只nil
	if visitNode == nil {
		// TODO set error info
		return this
	}

	if valueLeafNodeItems.Kind() != reflect.Interface {
		// 只有当路径访问的节点是interface{}的时候才能groupby,例如map[string]interface{}
		// 因为groupby后interface{}的数据类型会变
		// TODO set error info
		return this
	}
	valueLeafNodeItems = valueLeafNodeItems.Elem()
	if valueLeafNodeItems.Kind() == reflect.Slice {
		if valueLeafNodeItems.Len() > 0 {
			valueLeafNodeItem := valueLeafNodeItems.Index(0)
			valueNodeItemValue := valueLeafNodeItem.MapIndex(valueSelectedKey)

			valueLeafNodeItemsNew := reflect.MakeMap(reflect.MapOf(valueNodeItemValue.Type(), valueLeafNodeItems.Type()))
			for i := 0; i < valueLeafNodeItems.Len(); i++ {
				// valueLeafNodeItem: {"a":"1", "b":"2"}
				valueLeafNodeItem = valueLeafNodeItems.Index(i)
				// valueNodeItemValue: "1" selectedKey "a"
				valueNodeItemValue = valueLeafNodeItem.MapIndex(valueSelectedKey)
				valueLeafNodeItemsNewValue := valueLeafNodeItemsNew.MapIndex(valueNodeItemValue)

				if !valueLeafNodeItemsNewValue.IsValid() {
					valueLeafNodeItemsNewValue = reflect.MakeSlice(valueLeafNodeItems.Type(), 0, 0)
				}
				valueLeafNodeItemsNewValue = reflect.Append(valueLeafNodeItemsNewValue, valueLeafNodeItem)
				valueLeafNodeItemsNew.SetMapIndex(valueNodeItemValue, valueLeafNodeItemsNewValue)
			}
			valueVisitNode.SetMapIndex(valueVisitKey, valueLeafNodeItemsNew)
		}
	}

	// TODO set error info
	return this
}

type Policy struct {
	Base
}

func main() {
	a := Base{
		DataSet: map[string]interface{}{
			"path": []map[string]int{
				map[string]int{"a": 1, "d": 11},
				map[string]int{"a": 2, "d": 22},
				map[string]int{"a": 3, "d": 33},
				map[string]int{"a": 4, "d": 44},
				map[string]int{"a": 4, "d": 55},
				map[string]int{"a": 4, "d": 66},
				map[string]int{"a": 4, "d": 77},
			},
		},
	}
	a.Init().GroupByKey("path", "a").Done()
	fmt.Println(a)

	aa := Base{
		DataSet: map[string]interface{}{
			"a": map[string]int{"c": 22, "d": 11, "e": 55},
			"b": []interface{}{11, 22, uint(33)},
		},
	}
	policy := Policy{aa}
	policy.Init().Filter("a", []string{"d", "e"}).Filter("b", []interface{}{11, 22, 33}).Done()
	fmt.Println(aa)

	//a := new(Base)
	//a.DataSet = map[string]interface{}{"a": map[string]int{"c": 22, "d": 11, "e": 55}}
	//a.DataSet["c"] = 1

	//m := map[string]interface{}{"a": map[string]int{"c": 22, "d": 11, "e": 55}}
	//m := map[string]interface{}{"a": map[string]int{"c": 22, "d": 11, "e": 55}}
	//valueMap := reflect.ValueOf(m)
	//valueKey := reflect.ValueOf(interface{}("a"))
	//valueMapValue := valueMap.MapIndex(valueKey)
	//fmt.Println(valueMapValue.Elem().Type())
	//fmt.Println(valueMapValue.Interface().(map[string]int))

	//d := "fa"
	//c := interface{}(d)
	//e := interface{}(c)
	//fmt.Println(reflect.ValueOf(e).Kind())
	//fmt.Println(reflect.ValueOf(cc["a"]).Kind() == reflect.Map)

	//cc = "abc"
	//fmt.Println(cc)
	//fmt.Println(reflect.ValueOf(cc).Kind())

	//c := reflect.ValueOf([]string{"a", "b", "c"})
	//fmt.Println(c)
	//b := Policy{Base{DataSet: "1"}}
	//b := new(Policy)
	//b.DataSet = "1"
	//fmt.Println(b.Say())

	//fmt.Println(reflect.ValueOf(a["b"]).Kind())
	//fmt.Println(reflect.ValueOf(a["t"]).Kind() == reflect.Invalid)

	//c := []int{1, 2, 3, 4, 5, 6}
	//fmt.Println(c)

	//filter := make(map[interface{}]interface{})
	//filter[1] = 1
	//filter["abc"] = 1
	//fmt.Println(filter)

}
