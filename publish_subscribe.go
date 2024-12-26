package gopher_notify

import "sync"

// 发布-订阅模式实现

// Event 事件对象
// 事件包含主题(Topic)和数据(data)两部分，其中：
//   - 主题类似于频道，区分不同关注点的订阅者
//   - 数据即为事件包含的内容
//
// 泛型：
//   - T 事件的主题类型
//   - D 事件包含的内容类型
type Event[T comparable, D any] struct {
	// 主题
	topic T
	// 内容
	data D
	// 是否异步调用
	async bool
}

// NewEvent 事件对象构造函数
//
// 泛型：
//   - T 事件的主题类型
//   - D 事件包含的内容类型
//
// 参数：
//   - topic 事件主题
//   - data 事件内容
func NewEvent[T comparable, D any](topic T, data D) *Event[T, D] {
	return &Event[T, D]{
		topic: topic,
		data:  data,
		async: false,
	}
}

// GetTopic 获取事件的主题
func (e *Event[T, D]) GetTopic() T {
	return e.topic
}

// GetData 获取事件的内容
func (e *Event[T, D]) GetData() D {
	return e.data
}

// Subscriber 订阅者接口
//
// 泛型：
//   - T 订阅的事件的主题类型
//   - D 订阅的事件包含的内容类型
type Subscriber[T comparable, D any] interface {
	// OnSubscribe 订阅到对应主题的新事件时，该方法被调用
	//
	// 参数：
	//  - e 订阅到的事件对象
	OnSubscribe(e *Event[T, D])
}

// Broker 事件总线
//
// 泛型：
//   - T 处理的事件的主题类型
//   - D 处理的事件包含的内容类型
type Broker[T comparable, D any] struct {
	// 全部订阅者列表，其中：
	//  - 键：订阅的主题，类型：T
	//  - 值：对应主题的全部订阅者集合，类型：*sync.Map 键： Subscriber 值： void
	subscribers sync.Map
	// 消息队列
	queue chan *Event[T, D]
}

// NewBroker 事件总线的构造函数
//
// 泛型：
//   - T 处理的事件的主题类型
//   - D 处理的事件包含的内容类型
//
// 参数：
//   - buffer 消息队列通道缓冲区大小
//     在订阅者处理消息逻辑耗时的情况下，可能导致发布操作被阻塞，可设定一定大小的通道缓冲区
func NewBroker[T comparable, D any](buffer int) *Broker[T, D] {
	// 创建一个Broker
	broker := &Broker[T, D]{
		subscribers: sync.Map{},
		queue:       make(chan *Event[T, D], buffer),
	}
	// 在一个新的线程中准备接收事件并广播
	go func() {
		for e := range broker.queue {
			broker.broadcast(e)
		}
	}()
	return broker
}

// 广播事件给全部订阅者
//
//   - event 发布的事件对象
func (broker *Broker[T, D]) broadcast(event *Event[T, D]) {
	// 获取主题对应的订阅者列表
	topicMap, ok := broker.subscribers.Load(event.GetTopic())
	if !ok {
		return
	}
	// 执行事件发布
	if event.async {
		topicMap.(*sync.Map).Range(func(key, value any) bool {
			go key.(Subscriber[T, D]).OnSubscribe(event)
			return true
		})
	} else {
		topicMap.(*sync.Map).Range(func(key, value any) bool {
			key.(Subscriber[T, D]).OnSubscribe(event)
			return true
		})
	}
}

// Subscribe 订阅一个主题
//
//   - topic 要订阅的主题，不存在会创建
//   - subscribers 订阅该主题的订阅者，不定长参数
func (broker *Broker[T, D]) Subscribe(topic T, subscribers ...Subscriber[T, D]) {
	// 主题不存在则创建
	topicMap, ok := broker.subscribers.Load(topic)
	if !ok {
		topicMap = &sync.Map{}
		broker.subscribers.Store(topic, topicMap)
	}
	// 加入主题列表
	topicList := topicMap.(*sync.Map)
	for _, subscriber := range subscribers {
		topicList.Store(subscriber, void{})
	}
}

// UnSubscribe 取消订阅一个主题
//
//   - topic 要取消订阅的主题，不存在则不会做任何操作
//   - subscriber 订阅该主题的订阅者
func (broker *Broker[T, D]) UnSubscribe(topic T, subscriber Subscriber[T, D]) {
	// 移出订阅者列表
	topicMap, ok := broker.subscribers.Load(topic)
	if ok {
		topicMap.(*sync.Map).Delete(subscriber)
	}
}

// RemoveTopic 移除某个主题，该主题全部的订阅者也会被全部取消订阅
//
//   - topic 要移除的主题
func (broker *Broker[T, D]) RemoveTopic(topic T) {
	topicMap, ok := broker.subscribers.Load(topic)
	if ok {
		topicMap.(*sync.Map).Clear()
		broker.subscribers.Delete(topic)
	}
}

// RemoveAll 移除全部主题及其订阅者
func (broker *Broker[T, D]) RemoveAll() {
	broker.subscribers.Range(func(key, value any) bool {
		value.(*sync.Map).Clear()
		return true
	})
	broker.subscribers.Clear()
}

// Close 关闭Broker的消息队列，释放资源，关闭后该Broker无法再被用于发布消息
func (broker *Broker[T, D]) Close() {
	broker.RemoveAll()
	close(broker.queue)
}

// BasePublisher 基本发布者对象，可进行组合扩展
//
// 泛型：
//   - T 发布的事件的主题类型
//   - D 发布的事件包含的内容类型
type BasePublisher[T comparable, D any] struct {
	// 对应的事件总线
	broker *Broker[T, D]
}

// NewBasePublisher 创建一个基本发布者对象
//
// 泛型：
//   - T 发布的事件的主题类型
//   - D 发布的事件包含的内容类型
//
// 参数：
//   - broker 该发布者对应的事件总线对象
func NewBasePublisher[T comparable, D any](broker *Broker[T, D]) *BasePublisher[T, D] {
	return &BasePublisher[T, D]{
		broker: broker,
	}
}

// Publish 发布事件
//
//   - event 发布的事件对象
//   - async 是否异步通知订阅者
func (publisher *BasePublisher[T, D]) Publish(event *Event[T, D], async bool) {
	// 捕获可能出现的panic
	defer func() {
		_ = recover()
	}()
	event.async = async
	publisher.broker.queue <- event
}