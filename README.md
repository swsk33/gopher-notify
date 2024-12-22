#  Go-事件驱动

## 1，介绍

一个Go实现的**观察者模式**和**发布-订阅模式**的封装，可用于数据变化/事件传递的通知，将事件/数据变化和变化监听的操作解耦合。

该软件包支持同步和异步的事件通知与处理，且都是线程安全的，可以在并发的场景下使用。

##  2，安装依赖

在项目目录中执行下列命令安装依赖：

```bash
go get gitee.com/swsk33/gopher-notify
```

在使用该类库之前，需要使用者对**观察者模式**和**发布-订阅模式**有一定的了解，包括但不限于其适用场景、组成部分等等，若不太熟悉可以先参考该文章：[传送门](https://juejin.cn/post/7426954878681677858)

## 3，观察者模式实现单主题监听

对于某个数据或者状态量的变化，可能需要一到多个观察者进行监听，并在数据发生变化时做出相应的动作，此时可调用该类库的**观察者模式**实现。

观察者模式中，包含了下列组成部分：

- `Subject` 主题，即被观察的对象，其中包含了一个状态量，该状态量变化时会通知观察该主题的观察者
- `Observer` 观察者，即观察某个主题的对象，其观察的主题状态变化时，则会被通知，且能够在被通知时实现一些自定义的数据处理逻辑

对于观察者，提供了观察者接口`Observer`定义如下：

```go
// Observer 观察者接口
//
// 泛型：
//   - T 观察的主题的数据类型
type Observer[T any] interface {
	// OnUpdate 观察的对象更新后，该方法被调用
	//
	//  - data 观察的主题更新后的数据
	OnUpdate(data T)
}
```

通过实现该接口的`OnUpdate`方法，即可自定义在观察者观察到数据变化时，接收变化后的数据并对其进行自定义处理。

使用观察者模式的示例如下：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/gopher-notify"
	"time"
)

// TextObserver 实现Observer接口
type TextObserver struct {
	// 观察者名称
	name string
}

// OnUpdate 实现接收到更新后的自定义处理逻辑
func (observer *TextObserver) OnUpdate(data string) {
	fmt.Printf("[%s]接收到数据更新：%s\n", observer.name, data)
}

func main() {
	// 1.创建主题
	subject := gopher_notify.NewSubject[string]()
	// 2.创建观察者实例
	o1, o2 := &TextObserver{"观察者1"}, &TextObserver{"观察者2"}
	// 3.观察主题
	subject.Register(o1, o2)
	// 4.数据变化时发出通知
	subject.UpdateAndNotify("aaa", false)
	time.Sleep(1 * time.Second)
	subject.UpdateAndNotify("bbb", false)
}
```

在实现了观察者接口的自定义数据更新处理逻辑后，即可创建一个主题，并使观察者观察该主题，上述代码中`NewSubject`是主题的构造函数，返回`Subject`对象指针，其泛型`T`代表包含的状态量的类型，主题`Subject`对象有如下方法：

- `Register(observers ...Observer[T])` 添加观察该主题的观察者对象，参数为不定长参数，可一次添加多个观察者
- `Remove(observer Observer[T])` 移除该主题的观察者对象
- `Update(data T)` 更新主题的数据，但是不通知观察者
- `Notify(async bool)` 将当前主题的数据传递并通知全部观察者，参数`async`指定为`true`时，则会异步通知，否则同步通知
- `UpdateAndNotify(data T, async bool)` 更新自身状态，并同时通知全部观察者，参数：
	- `data` 更新的数据
	- `async` 是否异步通知


## 4，发布-订阅模式实现多维度事件处理

在更加复杂的事件处理情况下，例如事件可能存在不同的主题，而不同的订阅者关注不同的主题数据变化，就需要使用**发布-订阅模式**了。

在发布-订阅模式中，组成部分如下：

- `Event` 事件，包含主题`topic`和数据`data`两部分，其中：
	- 主题类似于频道，区分不同关注点的订阅者
	- 数据即为事件包含的内容
- `Publisher` 发布者，产生和发布事件的对象
- `Subscriber` 订阅者，订阅并接受对应主题的变化事件的对象
- `Broker` 事件总线/消息队列，作为发布者和订阅者的中介，处理事件的传递，它接收发布者的事件，并将事件发送给订阅了对应主题的订阅者，在`Broker`中维护了一个不同主题对应的订阅者列表

对于订阅者，提供了`Subscriber`接口定义如下：

```go
// Subscriber 订阅者接口
//
// 泛型：
//   - T 订阅的事件的主题的数据类型
//   - D 订阅的事件包含的内容的数据类型
type Subscriber[T comparable, D any] interface {
	// OnSubscribe 订阅到对应主题的新事件时，该方法被调用
	//
	// 参数：
	//  - e 订阅到的事件对象
	OnSubscribe(e *Event[T, D])
}
```

通过实现该接口的`OnSubscribe`方法，能够自定义订阅者接收到对应主题的事件时的处理逻辑。

使用发布-订阅模式的代码如下：

```go
package main

import (
	"fmt"
	"gitee.com/swsk33/gopher-notify"
)

// MessageSubscriber 实现订阅者接口
type MessageSubscriber struct {
	// 名字
	name string
}

// OnSubscribe 自定义接收到订阅事件的处理逻辑
func (subscriber *MessageSubscriber) OnSubscribe(e *gopher_notify.Event[string, string]) {
	fmt.Printf("[%s]接收到事件，主题：%s，内容：%s\n", subscriber.name, e.GetTopic(), e.GetData())
}

func main() {
	// 1.创建事件总线
	broker := gopher_notify.NewBroker[string, string]()
	// 2.创建发布者
	publisher := gopher_notify.NewBasePublisher[string, string](broker)
	// 3.创建订阅者
	s1, s2 := &MessageSubscriber{"订阅者1"}, &MessageSubscriber{"订阅者2"}
	// 4.通过事件总线，订阅者订阅对应的主题
	const topicOne, topicTwo = "topic-1", "topic-2"
	broker.Subscribe("topic-1", s1)
	broker.Subscribe("topic-2", s2)
	// 5.发布者发布事件
	publisher.Publish(gopher_notify.NewEvent(topicOne, "aaa"), false)
	publisher.Publish(gopher_notify.NewEvent(topicTwo, "bbb"), false)
}
```

在实现了订阅者接口的自定义订阅逻辑后，即可创建对应的事件总线和发布者对象，实现发布-订阅功能，其中：

- `NewEvent`是事件对象`Event`的构造函数，返回一个`Event`对象指针，其中泛型`T`表示事件主题的变量类型，`D`表示事件包含的数据变量类型，此外事件对象有如下方法：
	- `GetTopic` 获取该事件的主题
	- `GetData` 获取该事件包含的数据
- `NewBroker` 是事件总线对象`Broker`的构造函数，返回一个`Broker`对象指针，其泛型的意义和`Event`中的一样，`Broker`对象包含如下方法：
	- `Subscribe(topic T, subscribers ...Subscriber[T, D])` 指定数个订阅者订阅指定主题，参数：
		- `topic` 要订阅的主题，不存在该主题会自动创建
		- `subscribers` 订阅`topic`的订阅者列表，为不定长参数
	- `UnSubscribe(topic T, subscriber Subscriber[T, D])` 指定某个订阅者取消订阅指定主题，参数：
		- `topic` 要取消订阅的主题，若不存在则不会做任何操作
		- `subscriber` 要取消订阅`topic`的订阅者
- `NewBasePublisher`是基本发布者对象`BasePublisher`的构造函数，返回一个`BasePublisher`对象指针，其泛型的意义和`Event`中的一样，该函数传入一个`Broker`对象指定该发布者对应的事件总线对象，`BasePublisher`可被组合至自定义的结构体中进行扩展，此外它包含下列方法：
  - `Publish(event *Event[T, D], async bool)` 发布事件到事件总线，发布后事件总线会通知订阅了该事件主题的全部订阅者，此时订阅者的`OnSubscribe`方法会被调用其参数：
  	- `event` 发布的事件对象
  	- `async` 是否执行异步发布-订阅的事件操作

## 5，异步事件通知

无论是`Subject`对象的`Notify`和`UpdateAndNotify`方法，还是`BasePublisher`的`Publish`方法，都带有一个`bool`类型的参数`async`，当该参数为`true`时，事件的通知和自定义事件的处理逻辑将在一个新的线程中执行，否则全部在当前线程执行。

总体来说：

- 在观察者模式中，`Subject`的`Notify`和`UpdateAndNotify`方法在进行通知时，实质上是调用了全部`Observer`对象的`OnUpdate`方法，实现状态传递以及观察者的自定义事件处理逻辑
- 在发布-订阅模式中，`BasePublisher`的`Publish`方法实际上是调用了对应主题的全部`Subscriber`的`OnSubscribe`方法，实现事件传递以及订阅者的自定义事件处理逻辑

也就是说，如果观察者`Observer`或者订阅者`Subscriber`中自定义处理事件的逻辑耗时较长，在同步的事件发布场景下，`Notify`、`UpdateAndNotify`以及`Publish`方法会被阻塞较长时间，如果观察者/订阅者较多，那么通知每个观察者/订阅者都会耗费较长时间，导致整个事件通知/发布操作非常耗时。

在自定义事件处理逻辑较为复杂的情况下，或者在观察者/订阅者较多的情况下，可指定`async`参数为`true`，此时`Observer`对象的`OnUpdate`方法或`Subscriber`的`OnSubscribe`方法都会在一个新的线程中进行调用，不会使`Notify`、`UpdateAndNotify`以及`Publish`方法在当前线程阻塞。