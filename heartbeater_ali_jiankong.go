package spirit

import (
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
func (p *AliJiankong) Start(options Options) (err error) {
	if options == nil {
		err = ERR_HEARTBEAT_OPTIONS_IS_NIL.New(errors.Params{"name": p.Name()})
		return
	}

	var uid, metricName string = "", ""
	var reportPeriod, timeout int64 = 0, 0

	if uid, err = options.GetStringValue("uid"); err != nil {
		return
	}

	if metricName, err = options.GetStringValue("metric_name"); err != nil {
		return
	}

	if reportPeriod, err = options.GetInt64Value("report_period"); err != nil {
		return
	}

	if timeout, err = options.GetInt64Value("timeout"); err != nil {
		return
	}

	uid = strings.TrimSpace(uid)
	if uid == "" {
		err = ERR_HEARTBEAT_ALI_JIANKONG_UID_NOT_EXIST.New()
		return
	}

	metricName = strings.TrimSpace(metricName)
	if metricName == "" {
		metricName = "component_heartbeat"
	}

	if timeout == 0 {
		timeout = 1000
	}

	p.client = ali_jiankong.NewAliJianKong(uid, time.Duration(timeout)*time.Microsecond)

	if reportPeriod <= 60000 {
		reportPeriod = 60000
	}

	p.reportPeriod = time.Duration(reportPeriod) * time.Millisecond

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
				"instance_name": heartbeatMessage.InstanceName,
				"process_id":    strconv.Itoa(int(heartbeatMessage.PID)),
				"host_name":     heartbeatMessage.HostName,
				"start_time":    heartbeatMessage.StartTime.Format("2006-01-02 15:04:05"),
			},
			DimensionsOrder: []string{"instance_name", "process_id", "host_name", "start_time"},
		}

		EventCenter.PushEvent(EVENT_HEARTBEAT, item)

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
