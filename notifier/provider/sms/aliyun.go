package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/ntshibin/core/errorx"
	"github.com/ntshibin/core/notifier/provider"
)

// AliyunSMS 阿里云短信实现
type AliyunSMS struct {
	conf   *Config
	client *dysmsapi20170525.Client
}

// NewAliyunSMS 创建阿里云短信实例
func NewAliyunSMS(conf *Config) provider.NotificationSender {
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: &conf.AccessKeyID,
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: &conf.AccessKeySecret,
		RegionId:        &conf.RegionID,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	client := &dysmsapi20170525.Client{}
	client, err := dysmsapi20170525.NewClient(config)

	if err != nil {
		panic(fmt.Sprintf("failed to create aliyun sms client: %v", err))
	}
	return &AliyunSMS{
		conf:   conf,
		client: client,
	}
}

// Send 发送短信
func (a *AliyunSMS) Send(ctx context.Context, msg provider.Message) (res *provider.MessageRes, err error) {
	notice, ok := msg.Content.(*provider.MessageSMS)
	if !ok {
		return nil, errorx.New(errorx.HTTPCodeError, "消息内容类型必须为SMS类型，当前类型无效")
	}
	// 构建请求参数
	templateParamJSON, err := json.Marshal(notice.TemplateCode)
	if err != nil {
		return nil, err
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers: tea.String(msg.To),
		SignName:     tea.String(notice.SignName),
		TemplateCode: tea.String(notice.TemplateCode),

		TemplateParam: tea.String(string(templateParamJSON)),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, _err := a.client.SendSmsWithOptions(sendSmsRequest, runtime)
		if _err != nil {
			return _err
		}

		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		fmt.Println(tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			fmt.Println(recommend)
		}
		_, _err := util.AssertAsString(error.Message)
		if _err != nil {
			return nil, _err
		}
	}
	return
}
