package spirit

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogap/ali_jiankong"
	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

type AliJiankong struct {
	client *ali_jiankong.AliJianKong

	count        int64
	countLocker  sync.Mutex
	reportPeriod time.Duration

	lastReportTime time.Time
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
			UID          string `json:"uid"`
			MetricName   string `json:"metric_name"`
			ReportPeriod int64  `json:"report_period"`
			Timeout      int64  `json:"timeout"`
		} `json:"ali_jiankong"`
	}

	if data, e := ioutil.ReadFile(configFile); e != nil {
		err = ERR_READE_FILE_ERROR.New(errors.Params{"err": e, "file": configFile})
		return
	} else if e := json.Unmarshal(data, &tmp); e != nil {
		err = ERR_UNMARSHAL_DATA_ERROR.New(errors.Params{"err": e})
		return
	}

	tmp.AliJIankongConfig.UID = strings.TrimSpace(tmp.AliJIankongConfig.UID)
	if tmp.AliJIankongConfig.UID == "" {
		err = ERR_HEARTBEAT_ALI_JIANKONG_UID_NOT_EXIST.New()
		return
	}

	tmp.AliJIankongConfig.MetricName = strings.TrimSpace(tmp.AliJIankongConfig.MetricName)
	if tmp.AliJIankongConfig.MetricName == "" {
		tmp.AliJIankongConfig.MetricName = "component_heartbeat"
	}

	if tmp.AliJIankongConfig.Timeout == 0 {
		tmp.AliJIankongConfig.Timeout = 1000
	}

	p.client = ali_jiankong.NewAliJianKong(tmp.AliJIankongConfig.UID, time.Duration(tmp.AliJIankongConfig.Timeout)*time.Microsecond)

	if tmp.AliJIankongConfig.ReportPeriod <= 60000 {
		tmp.AliJIankongConfig.ReportPeriod = 60000
	}

	p.reportPeriod = time.Duration(tmp.AliJIankongConfig.ReportPeriod) * time.Millisecond

	p.lastReportTime = time.Now()

	return
}
func (p *AliJiankong) Heartbeat(heartbeatMessage HeartbeatMessage) {
	p.countLocker.Lock()
	defer p.countLocker.Unlock()

	now := time.Now()

	if now.Sub(p.lastReportTime) >= p.reportPeriod {
		item := ali_jiankong.ReportItem{
			MetricName:  "component_heartbeat",
			MetricValue: strconv.Itoa(int(p.count)),
			Dimensions: ali_jiankong.Dimensions{
				"component_name": heartbeatMessage.Component,
				"process_id":     strconv.Itoa(int(heartbeatMessage.PID)),
				"host_name":      heartbeatMessage.HostName,
				"start_time":     heartbeatMessage.StartTime.Format("2006-01-02 15:04:05"),
			},
			DimensionsOrder: []string{"component_name", "process_id", "host_name", "start_time"},
		}

		if err := p.client.Report(item); err != nil {
			logs.Error(err)
		} else {
			p.count = 0
			p.lastReportTime = time.Now()
		}
	} else {
		p.count++
	}

	return
}
