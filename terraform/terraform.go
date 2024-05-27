package main

import (
	"encoding/json"
	"fmt"
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	plugin2 "github.com/kaytu-io/kaytu/pkg/plugin"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/zclconf/go-cty/cty"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TFState struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Module    string     `json:"module"`
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	Arn string `json:"arn"`
}

func main() {
	manager := plugin2.New()
	manager.SetNonInteractiveView()
	err := manager.StartServer()
	if err != nil {
		panic(err)
	}
	err = manager.StartPlugin("rds-instance")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		runningPlg := manager.GetPlugin("kaytu-io/plugin-aws")
		if runningPlg != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	runningPlg := manager.GetPlugin("kaytu-io/plugin-aws")
	if runningPlg == nil {
		panic("running plugin not found")
	}
	cfg, err := server.GetConfig()
	if err != nil {
		panic(err)
	}

	for _, rcmd := range runningPlg.Plugin.Config.Commands {
		if rcmd.Name == "rds-instance" {
			preferences.Update(rcmd.DefaultPreferences)

			if rcmd.LoginRequired && cfg.AccessToken == "" {
				// login
				panic("please login")
			}
			break
		}
	}
	err = runningPlg.Stream.Send(&golang.ServerMessage{
		ServerMessage: &golang.ServerMessage_Start{
			Start: &golang.StartProcess{
				Command:          "rds-instance",
				Flags:            nil,
				KaytuAccessToken: cfg.AccessToken,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	jsonOutput, err := manager.NonInteractiveView.WaitAndReturnResults("json")
	if err != nil {
		panic(err)
	}

	var jsonObj struct {
		Items []*golang.OptimizationItem
	}
	err = json.Unmarshal([]byte(jsonOutput), &jsonObj)
	if err != nil {
		panic(err)
	}

	recommendation := map[string]string{}
	for _, item := range jsonObj.Items {
		var recommendedInstanceSize string
		for _, device := range item.Devices {
			for _, property := range device.Properties {
				if property.Key == "Instance Size" {
					recommendedInstanceSize = property.Recommended
				}
			}
		}
		recommendation[item.Id] = recommendedInstanceSize
	}

	folder := "/home/saleh/projects/keibi/terraform-aws-rds/examples/complete-postgres/"
	err = filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".tf") {
			return nil
		}
		if strings.Contains(path, ".terraform") {
			return nil
		}
		if info.Name() != "main.tf" {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		file, diags := hclwrite.ParseConfig(contentBytes, "filename.tf", hcl2.InitialPos)
		if diags.HasErrors() {
			return fmt.Errorf("%s", diags.Error())
		}

		body := file.Body()
		localVars := map[string]string{}
		for _, block := range body.Blocks() {
			if block.Type() == "locals" {
				for k, v := range block.Body().Attributes() {
					value := strings.TrimSpace(string(v.Expr().BuildTokens(hclwrite.Tokens{}).Bytes()))

					localVars[k] = value
				}
			}
			if block.Type() == "module" {
				identifier := block.Body().GetAttribute("identifier")
				if identifier == nil {
					continue
				}

				value := strings.TrimSpace(string(identifier.Expr().BuildTokens(hclwrite.Tokens{}).Bytes()))
				value = resolveValue(localVars, value)

				var instanceUseIdentifierPrefixBool bool
				instanceUseIdentifierPrefix := block.Body().GetAttribute("instance_use_identifier_prefix")
				if instanceUseIdentifierPrefix != nil {
					boolValue := strings.TrimSpace(string(instanceUseIdentifierPrefix.Expr().BuildTokens(hclwrite.Tokens{}).Bytes()))
					instanceUseIdentifierPrefixBool = boolValue == "true"
				}

				if instanceUseIdentifierPrefixBool {
					for k, v := range recommendation {
						if strings.HasPrefix(value, k) {
							dbNameAttr := block.Body().GetAttribute("db_name")
							if dbNameAttr != nil {
								block.Body().SetAttributeValue("instance_class", cty.StringVal(v))
							}
						}
					}
				} else {
					if _, ok := recommendation[value]; ok {
						dbNameAttr := block.Body().GetAttribute("db_name")
						if dbNameAttr != nil {
							block.Body().SetAttributeValue("instance_class", cty.StringVal(recommendation[value]))
						}
					}
				}
			}
		}

		err = os.WriteFile(path, file.Bytes(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error writing HCL file: %v", err)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking:", err)
		return
	}
}

func resolveValue(vars map[string]string, value string) string {
	varRegEx, err := regexp.Compile("local\\.(\\w+)")
	if err != nil {
		panic(err)
	}

	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = strings.TrimPrefix(value, "\"")
		value = strings.TrimSuffix(value, "\"")

		exprRegEx, err := regexp.Compile("\\$\\{([\\w.]+)}")
		if err != nil {
			panic(err)
		}

		items := exprRegEx.FindAllString(value, 100)
		for _, item := range items {
			resolvedItem := resolveValue(vars, item)
			value = strings.ReplaceAll(value, item, resolvedItem)
		}
		return value
	} else {
		if varRegEx.MatchString(value) {
			subMatch := varRegEx.FindStringSubmatch(value)
			value = vars[subMatch[1]]
			return resolveValue(vars, value)
		} else {
			return value
		}
	}
}
