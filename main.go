package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

// mergeJSON 递归合并两个 JSON 对象，B 的内容合并到 A 中
func mergeJSON(a map[string]interface{}, b map[string]interface{}) {
	for bKey, bValue := range b {
		// 如果 A 中存在该 key，尝试递归合并
		if aValue, exists := a[bKey]; exists {
			// 如果值是字典类型，递归合并
			if aValueMap, ok := aValue.(map[string]interface{}); ok {
				if bValueMap, ok := bValue.(map[string]interface{}); ok {
					mergeJSON(aValueMap, bValueMap)
					a[bKey] = aValueMap
				}
			} else if aValueArray, ok := aValue.([]interface{}); ok {
				// 如果值是数组，处理数组合并
				if bValueArray, ok := bValue.([]interface{}); ok {
					// 如果数组元素是基础类型，直接赋值
					if isBasicTypeArray(bValueArray) {
						// 基础类型数组直接替换
						a[bKey] = bValueArray
					} else {
						// 递归合并数组
						mergeArray(&aValueArray, bValueArray)
						a[bKey] = aValueArray
					}
				}
			} else {
				// 如果值是普通类型，直接替换
				a[bKey] = bValue
			}
		} else {
			// 如果 A 中不存在该 key，直接添加
			a[bKey] = bValue
		}
	}
}

// isBasicTypeArray 判断一个数组是否是基础类型的数组
func isBasicTypeArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}
	// 判断第一个元素类型
	switch arr[0].(type) {
	case string, bool, float64, int, float32:
		return true
	default:
		return false
	}
}

// mergeArray 处理数组类型的合并
func mergeArray(a *[]interface{}, b []interface{}) {
	for _, bItem := range b {
		// 如果 bItem 是字典类型
		if bItemMap, ok := bItem.(map[string]interface{}); ok {
			// 获取主键名称（id, name, identification, config等）
			var keyName string
			var keyVal interface{}
			var identificationKeyName string

			// 确定主键名称和对应的值
			if keyVal = bItemMap["id"]; keyVal != nil {
				keyName = "id"
			} else if keyVal = bItemMap["name"]; keyVal != nil {
				keyName = "name"
			} else if keyVal = bItemMap["identification"]; keyVal != nil {
				keyName = "identification"
				// 处理 identification 类型，获取其中的 id 或 name
				if identificationMap, ok := bItemMap["identification"].(map[string]interface{}); ok {
					if idVal, ok := identificationMap["id"].(string); ok {
						identificationKeyName = "id"
						keyVal = idVal
					} else if nameVal, ok := identificationMap["name"].(string); ok {
						identificationKeyName = "name"
						keyVal = nameVal
					}
				}
			} else if keyVal = bItemMap["config"]; keyVal != nil {
				keyName = "config"
				// 处理 config 类型，获取其中的 id 或 name
				if configMap, ok := bItemMap["config"].(map[string]interface{}); ok {
					if idVal, ok := configMap["id"].(string); ok {
						identificationKeyName = "id"
						keyVal = idVal
					} else if nameVal, ok := configMap["name"].(string); ok {
						identificationKeyName = "name"
						keyVal = nameVal
					}
				}
			}

			// 查找 A 中是否存在相同主键的元素
			found := false
			if keyName == "" {
				// 如果没有找到主键名，则直接插入到数组
				*a = append(*a, bItem)
			} else {
				// 遍历 A 中的数组元素
				for j, aItem := range *a {
					if aItemMap, ok := aItem.(map[string]interface{}); ok {
						// 先取出对应的 keyName，再根据 identificationKeyName 查找
						var aKeyVal interface{}
						if keyName == "identification" || keyName == "config" {
							// 处理 identification 或 config 类型，进一步查找 id 或 name
							if aItemMapValue, exists := aItemMap[keyName]; exists {
								if aItemMapValueMap, ok := aItemMapValue.(map[string]interface{}); ok {
									aKeyVal = aItemMapValueMap[identificationKeyName]
								}
							}
						} else {
							// 否则直接通过 id 或 name 查找
							aKeyVal = aItemMap[keyName]
						}

						// 如果主键匹配
						if aKeyVal == keyVal {
							// 找到匹配项，递归合并
							mergeJSON(aItemMap, bItemMap)
							(*a)[j] = aItemMap
							found = true
							break
						}
					}
				}

				// 如果没有找到匹配的元素，直接新增
				if !found {
					// 插入元素到数组
					*a = append(*a, bItem)
					// 如果 keyName 是 "id" 且 keyVal 是数字，则对 A 数组进行排序
					if keyName == "id" {
						if _, ok := keyVal.(float64); ok {
							// 排序，按照 id 从小到大
							sort.SliceStable(*a, func(i, j int) bool {
								// 比较 id 值
								iVal, iOk := (*a)[i].(map[string]interface{})["id"].(float64)
								jVal, jOk := (*a)[j].(map[string]interface{})["id"].(float64)
								return iOk && jOk && iVal < jVal
							})
						}
					}
				}
			}
		} else {
			// 如果 bItem 不是字典类型（基础类型），直接用 B 的值替换 A 中的值
			*a = append(*a, bItem)
		}
	}
}

// main 函数，执行命令行参数解析和文件合并操作
func main() {
	// 定义命令行参数
	target := flag.String("t", "/data/udapi-config/ubios-udapi-server/ubios-udapi-server.state", "Path to the target file (A).")
	source := flag.String("s", "/data/myconfig_gateway_json/myconfig.gateway.json", "Path to the source file (B).")
	output := flag.String("o", "/data/udapi-config/ubios-udapi-server/ubios-udapi-server.state", "Path to the output file (merged).")
	help := flag.Bool("h", false, "Display help.")

	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if *help {
		fmt.Printf("Target file: %s (default: %s)\n", *target, "/data/udapi-config/ubios-udapi-server/ubios-udapi-server.state")
		fmt.Printf("Source file: %s (default: %s)\n", *source, "/data/myconfig_gateway_json/myconfig.gateway.json")
		fmt.Printf("Output file: %s (default: %s)\n", *output, "/data/udapi-config/ubios-udapi-server/ubios-udapi-server.state")
		return
	}

	// 读取目标文件 (A)
	aFile, err := ioutil.ReadFile(*target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading target file: %v\n", err)
		os.Exit(1)
	}

	// 读取源文件 (B)
	bFile, err := ioutil.ReadFile(*source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
		os.Exit(1)
	}

	// 定义目标和源的结构
	var aData, bData map[string]interface{}

	// 解析 JSON 数据
	if err := json.Unmarshal(aFile, &aData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing target file JSON: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(bFile, &bData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing source file JSON: %v\n", err)
		os.Exit(1)
	}

	// 合并数据
	mergeJSON(aData, bData)

	// 将合并后的数据写入输出文件
	mergedFile, err := json.MarshalIndent(aData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling merged data: %v\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(*output, mergedFile, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing merged file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Merge completed successfully.")
}
