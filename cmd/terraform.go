package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/kaytu-io/kaytu/pkg/github"
	plugin2 "github.com/kaytu-io/kaytu/pkg/plugin"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var terraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Create pull request for right sizing opportunities on your terraform git",
	Long:  "Create pull request for right sizing opportunities on your terraform git",
	RunE: func(cmd *cobra.Command, args []string) error {
		ignoreYoungerThan := utils.ReadIntFlag(cmd, "ignore-younger-than")
		contentBytes, err := github.GetFile(
			utils.ReadStringFlag(cmd, "github-owner"),
			utils.ReadStringFlag(cmd, "github-repo"),
			utils.ReadStringFlag(cmd, "terraform-file-path"),
			utils.ReadStringFlag(cmd, "github-username"),
			utils.ReadStringFlag(cmd, "github-token"),
		)
		if err != nil {
			return err
		}

		manager := plugin2.New()
		manager.SetNonInteractiveView()
		err = manager.StartServer()
		if err != nil {
			return err
		}
		err = manager.StartPlugin("rds-instance")
		if err != nil {
			return err
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
			return fmt.Errorf("running plugin not found")
		}
		cfg, err := server.GetConfig()
		if err != nil {
			return err
		}

		for _, rcmd := range runningPlg.Plugin.Config.Commands {
			if rcmd.Name == "rds-instance" {
				preferences.Update(rcmd.DefaultPreferences)

				if rcmd.LoginRequired && cfg.AccessToken == "" {
					// login
					return fmt.Errorf("please login")
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
			return err
		}
		jsonOutput, err := manager.NonInteractiveView.WaitAndReturnResults("json")
		if err != nil {
			return err
		}

		var jsonObj struct {
			Items []*golang.OptimizationItem
		}
		err = json.Unmarshal([]byte(jsonOutput), &jsonObj)
		if err != nil {
			return err
		}

		recommendation := map[string]string{}
		rightSizingDescription := map[string]string{}
		for _, item := range jsonObj.Items {
			var recommendedInstanceSize string
			maxRuntimeHours := int64(1) // since default for ignoreYoungerThan is 1
			for _, device := range item.Devices {
				for _, property := range device.Properties {
					if property.Key == "RuntimeHours" {
						i, _ := strconv.ParseInt(property.Current, 10, 64)
						maxRuntimeHours = max(maxRuntimeHours, i)
					}
					if property.Key == "Instance Size" && property.Current != property.Recommended {
						recommendedInstanceSize = property.Recommended
					}
				}
			}

			if maxRuntimeHours < ignoreYoungerThan {
				continue
			}
			if recommendedInstanceSize == "" {
				continue
			}
			recommendation[item.Id] = recommendedInstanceSize
			rightSizingDescription[item.Id] = item.Description
		}

		file, diags := hclwrite.ParseConfig(contentBytes, "filename.tf", hcl.InitialPos)
		if diags.HasErrors() {
			return fmt.Errorf("%s", diags.Error())
		}

		body := file.Body()
		localVars := map[string]string{}
		countRightSized := 0
		var rightSizedIds []string
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
						if strings.HasPrefix(k, value) {
							dbNameAttr := block.Body().GetAttribute("db_name")
							if dbNameAttr != nil {
								block.Body().SetAttributeValue("instance_class", cty.StringVal(v))
								countRightSized++
								rightSizedIds = append(rightSizedIds, k)
							}
						}
					}
				} else {
					if _, ok := recommendation[value]; ok {
						dbNameAttr := block.Body().GetAttribute("db_name")
						if dbNameAttr != nil {
							block.Body().SetAttributeValue("instance_class", cty.StringVal(recommendation[value]))
							countRightSized++
							rightSizedIds = append(rightSizedIds, value)
						}
					}
				}
			}
		}

		description := ""
		for _, id := range rightSizedIds {
			description += fmt.Sprintf("Changing instance class of %s to %s\n", id, recommendation[id])
			description += rightSizingDescription[id] + "\n\n"
		}

		if countRightSized == 0 {
			return nil
		}
		return github.ApplyChanges(
			utils.ReadStringFlag(cmd, "github-owner"),
			utils.ReadStringFlag(cmd, "github-repo"),
			utils.ReadStringFlag(cmd, "github-username"),
			utils.ReadStringFlag(cmd, "github-token"),
			utils.ReadStringFlag(cmd, "github-base-branch"),
			fmt.Sprintf("SRE Bot right sizing %d resources", countRightSized),
			utils.ReadStringFlag(cmd, "terraform-file-path"),
			string(file.Bytes()),
			fmt.Sprintf("SRE Bot right sizing %d resources", countRightSized),
			description,
		)
	},
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
