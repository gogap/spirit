package spirit

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/gogap/ali_jiankong"
	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

type AliJiankong struct {
	client *ali_jiankong.AliJianKong
}

func (p *AliJiankong) Name() string {
	return "ali_jiankong"
}
func (p *AliJiankong) Start(configFile string) (err error) {
	if configFile == "" {
		err = ERR_HEARTBEAT_CONFIG_FILE_IS_EMPTY.New(errors.Params{"name": p.Name()})
		return
	}

	var tmp struct {
		AliJIankongConfig struct {
			UID     string `json:"uid"`
			timeout int64  `json:"timeout"`
		} `json:"ali_jiankong"`
	}

	if data, e := ioutil.ReadFile(configFile); e != nil {
		err = ERR_READE_FILE_ERROR.New(errors.Params{"err": e, "file": configFile})
		return
	} else if e := json.Unmarshal(data, &tmp); e != nil {
		err = ERR_UNMARSHAL_DATA_ERROR.New(errors.Params{"err": e})
		return
	}

	if tmp.AliJIankongConfig.UID == "" {
		err = ERR_HEARTBEAT_ALI_JIANKONG_UID_NOT_EXIST.New()
		return
	}

	if tmp.AliJIankongConfig.timeout == 0 {
		tmp.AliJIankongConfig.timeout = 1000
	}

	p.client = ali_jiankong.NewAliJianKong(tmp.AliJIankongConfig.UID, time.Duration(tmp.AliJIankongConfig.timeout)*time.Microsecond)

	return
}
func (p *AliJiankong) Heartbeat(heartbeatMessage HeartbeatMessage) {

	item := ali_jiankong.ReportItem{
		MetricName:  "component_heartbeat",
		MetricValue: "1",
		Dimensions: ali_jiankong.Dimensions{
			"component_name": heartbeatMessage.Component,
			"process_id":     strconv.Itoa(int(heartbeatMessage.PID)),
			"host_name":      heartbeatMessage.HostName,
			"start_time":     heartbeatMessage.StartTime.Format("2006-01-02 15:04:05"),
		},
	}

	if err := p.client.Report(item); err != nil {
		logs.Error(err)
	}

	return
}
