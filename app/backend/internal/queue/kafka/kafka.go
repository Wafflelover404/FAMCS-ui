package kafka 

import (
	"context"
	"encoding/json"

	kafkago "github.com/segmentio/kafka-go"

	"famcs-ui/backend/internal/queue"
)

const topic = "famcs.jobs"

type Queue struct {
	brokers []string
	writer  *kafkago.Writer
}

func New(brokers []string) *Queue {
	return &Queue{
		brokers: brokers,
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (q *Queue) Enqueue(ctx context.Context, j queue.Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return q.writer.WriteMessages(ctx, kafkago.Message{Key: []byte(j.ID), Value: b})
}

func (q *Queue) Consume(ctx context.Context, group string, fn queue.Handler) error {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: q.brokers,
		Topic:   topic,
		GroupID: group,
	})
	defer r.Close()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		var j queue.Job
		if err := json.Unmarshal(m.Value, &j); err == nil {
			_ = fn(ctx, j)
		}
		_ = r.CommitMessages(ctx, m)
	}
}

func (q *Queue) Depth(ctx context.Context) (int64, error) {
	if len(q.brokers) == 0 {
		return 0, nil
	}
	conn, err := kafkago.DialContext(ctx, "tcp", q.brokers[0])
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return 0, nil
	}
	var total int64
	for _, p := range partitions {
		pconn, err := kafkago.DialLeader(ctx, "tcp", q.brokers[0], topic, p.ID)
		if err != nil {
			continue
		}
		first, last, err := pconn.ReadOffsets()
		pconn.Close()
		if err == nil {
			total += last - first
		}
	}
	return total, nil
}

func (q *Queue) Close() error {
	return q.writer.Close()
}
