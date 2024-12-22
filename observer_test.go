package gopher_notify

import (
	"fmt"
	"testing"
	"time"
)

// 实现Observer接口
type TextObserver struct {
	// 观察者名称
	name string
}

// 实现接收到更新后的自定义处理逻辑
func (observer *TextObserver) OnUpdate(data string) {
	fmt.Printf("[%s]接收到数据更新：%s\n", observer.name, data)
}

// 测试观察者-更新并通知
func TestObserver_UpdateAndNotify(t *testing.T) {
	// 1.创建主题
	subject := NewSubject[string]()
	// 2.创建观察者实例
	o1, o2 := &TextObserver{"观察者1"}, &TextObserver{"观察者2"}
	// 3.订阅主题
	subject.Register(o1, o2)
	// 4.数据变化时发出通知
	subject.UpdateAndNotify("aaa", false)
	time.Sleep(1 * time.Second)
	subject.UpdateAndNotify("bbb", false)
}

// 测试观察者-手动通知
func TestObserver_ManuallyNotify(t *testing.T) {
	// 1.创建主题
	subject := NewSubject[string]()
	// 2.创建观察者实例
	o1, o2 := &TextObserver{"观察者1"}, &TextObserver{"观察者2"}
	// 3.订阅主题
	subject.Register(o1, o2)
	// 4.更新数据但是不通知
	subject.Update("aaa")
	subject.Update("bbb")
	// 5.需要的时候手动通知
	time.Sleep(1 * time.Second)
	subject.Notify(false)
}

// 测试观察者-更新并异步通知
func TestObserver_UpdateAndNotifyAsync(t *testing.T) {
	// 1.创建主题
	subject := NewSubject[string]()
	// 2.创建观察者实例
	o1, o2 := &TextObserver{"观察者1"}, &TextObserver{"观察者2"}
	// 3.订阅主题
	subject.Register(o1, o2)
	// 4.数据变化时发出通知
	subject.UpdateAndNotify("aaa", true)
	subject.UpdateAndNotify("bbb", true)
	time.Sleep(1 * time.Second)
}