package classic

import (
	"sync"
	"time"

	"github.com/gogap/spirit"
)

const (
	inboxURN = "urn:spirit:inbox:classic"
)

var _ spirit.Inbox = new(ClassicInbox)

type ClassicInboxConfig struct {
	Size       int `json:"size"`
	PutTimeout int `json:"put_timeout"`
	GetTimeout int `json:"get_timeout"`
}

type ClassicInbox struct {
	statusLocker sync.Mutex
	status       spirit.Status

	receivers    []spirit.Receiver
	receiverLock sync.Mutex

	deliveriesChan chan []spirit.Delivery

	conf ClassicInboxConfig
}

func init() {
	spirit.RegisterInbox(inboxURN, NewClassicInbox)
}

func NewClassicInbox(config spirit.Map) (box spirit.Inbox, err error) {
	conf := ClassicInboxConfig{}

	if err = config.ToObject(&conf); err != nil {
		return
	}

	box = &ClassicInbox{
		conf: conf,
	}
	return

}

func (p *ClassicInbox) Start() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().WithField("actor", spirit.ActorInbox).
		WithField("urn", inboxURN).
		WithField("event", "start").
		Debugln("enter start")

	if p.status == spirit.StatusRunning {
		return
	}

	p.deliveriesChan = make(chan []spirit.Delivery, p.conf.Size)

	p.status = spirit.StatusRunning

	for _, receiver := range p.receivers {

		go func(receiver spirit.Receiver) {
			if receiver.Status() == spirit.StatusStopped {
				if err = receiver.Start(); err != nil {
					spirit.Logger().WithField("actor", spirit.ActorInbox).
						WithField("urn", inboxURN).
						WithField("event", "start receiver").
						Errorln(err)
				}

				spirit.Logger().WithField("actor", spirit.ActorInbox).
					WithField("urn", inboxURN).
					WithField("event", "start receiver").
					Debugln("receiver started")
			}
		}(receiver)
	}

	spirit.Logger().WithField("actor", spirit.ActorInbox).
		WithField("urn", inboxURN).
		WithField("event", "start").
		Infoln("started")

	return
}

func (p *ClassicInbox) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().WithField("actor", spirit.ActorInbox).
		WithField("urn", inboxURN).
		WithField("event", "stop").
		Debugln("enter stop")

	if p.status == spirit.StatusStopped {
		return
	}

	p.status = spirit.StatusStopped

	close(p.deliveriesChan)

	spirit.Logger().WithField("actor", spirit.ActorInbox).
		WithField("urn", inboxURN).
		WithField("event", "stop").
		Infoln("stopped")

	return
}

func (p *ClassicInbox) Status() (status spirit.Status) {
	return p.status
}

func (p *ClassicInbox) Put(deliveries []spirit.Delivery) (err error) {
	if deliveries == nil || len(deliveries) == 0 {
		return
	}

	if p.conf.PutTimeout < 0 {
		p.deliveriesChan <- deliveries
	} else {
		select {
		case p.deliveriesChan <- deliveries:
			{
				spirit.Logger().WithField("actor", spirit.ActorInbox).
					WithField("urn", inboxURN).
					WithField("event", "put deliveries").
					WithField("length", len(deliveries)).
					WithField("chan_size", len(p.deliveriesChan)).
					WithField("chan_cap", cap(p.deliveriesChan)).
					Debugln("put deliveries to delivery chan")
			}
		case <-time.After(time.Duration(p.conf.PutTimeout) * time.Millisecond):
			{
				spirit.Logger().WithField("actor", spirit.ActorInbox).
					WithField("urn", inboxURN).
					WithField("event", "put deliveries").
					WithField("chan_size", len(p.deliveriesChan)).
					WithField("chan_cap", cap(p.deliveriesChan)).
					Debugln("put deliveries timeout")
			}
		}
	}

	return
}

func (p *ClassicInbox) Get() (deliveries []spirit.Delivery, err error) {
	if p.conf.GetTimeout < 0 {
		deliveries = <-p.deliveriesChan
	} else {
		select {
		case deliveries = <-p.deliveriesChan:
			{
			}
		case <-time.After(time.Duration(p.conf.GetTimeout) * time.Millisecond):
			{
				spirit.Logger().WithField("actor", spirit.ActorInbox).
					WithField("urn", inboxURN).
					WithField("event", "get deliveries").
					Debugln("get deliveries timeout")
			}
		}

	}

	return
}
