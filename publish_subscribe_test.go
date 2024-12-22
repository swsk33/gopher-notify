package gopher_notify

import (
	"fmt"
	"testing"
	"time"
)

// 实现订阅者接口
type MessageSubscriber struct {
	// 名字
	name string
}

// 自定义接收到订阅事件的处理逻辑
func (subscriber *MessageSubscriber) OnSubscribe(e *Event[string, string]) {
	fmt.Printf("[%s]接收到事件，主题：%s，内容：%s\n", subscriber.name, e.GetTopic(), e.GetData())
}

// 测试发布-订阅功能
func TestPublish_Subscribe(t *testing.T) {
	// 1.创建事件总线
	broker := NewBroker[string, string]()
	// 2.创建发布者
	publisher := NewBasePublisher[string, string](broker)
	// 3.创建订阅者
	s1, s2 := &MessageSubscriber{"订阅者1"}, &MessageSubscriber{"订阅者2"}
	// 4.通过事件总线订阅主题
	const topicOne, topicTwo = "topic-1", "topic-2"
	broker.Subscribe("topic-1", s1)
	broker.Subscribe("topic-2", s2)
	// 5.发布者发布事件
	publisher.Publish(NewEvent(topicOne, "aaa"), true)
	publisher.Publish(NewEvent(topicTwo, "bbb"), true)
	time.Sleep(100 * time.Millisecond)
}