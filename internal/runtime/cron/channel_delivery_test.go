package cron

import (
	"context"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeChannelDeliverer struct {
	job  *agentsv1.CronJob
	exec *agentsv1.CronExecution
}

func (d *fakeChannelDeliverer) DeliverCronResult(_ context.Context, job *agentsv1.CronJob, exec *agentsv1.CronExecution) error {
	d.job = job
	d.exec = exec
	return nil
}

func TestDeliverChannelUsesConfiguredDeliverer(t *testing.T) {
	deliverer := &fakeChannelDeliverer{}
	s := &Scheduler{ctx: context.Background(), deliverer: deliverer}
	job := &agentsv1.CronJob{
		Name: "daily-report",
		Delivery: &agentsv1.CronDelivery{
			Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
			ChannelName: "ops",
			ChatId:      "123",
		},
	}
	exec := &agentsv1.CronExecution{JobName: "daily-report"}

	s.deliver(job, exec)

	if deliverer.job != job {
		t.Fatal("deliverer did not receive job")
	}
	if deliverer.exec != exec {
		t.Fatal("deliverer did not receive execution")
	}
}
