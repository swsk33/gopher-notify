package gopher_notify

import "sync"

// 观察者模式实现

// 自定义空类型
type void struct{}

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

// Subject 观察主题（被观察对象）
//
// 泛型：
//   - T 主题的（状态）数据类型
type Subject[T any] struct {
	// 主题数据（状态），变化时将通知观察者
	data T
	// 观察者该主题的观察者列表
	observers sync.Map
}

// Register 注册观察者
//
//   - observers 要注册的观察者，不定长参数
func (subject *Subject[T]) Register(observers ...Observer[T]) {
	// 执行注册
	for _, observer := range observers {
		subject.observers.Store(observer, void{})
	}
}

// Remove 移除观察者
//
//   - observer 要移除的观察者
func (subject *Subject[T]) Remove(observer Observer[T]) {
	// 执行移除
	subject.observers.Delete(observer)
}

// Update 更新数据，但是不通知观察者
//
//   - data 更新的数据值
func (subject *Subject[T]) Update(data T) {
	// 更新数据
	subject.data = data
}

// Notify 将当前主题的数据传递并通知全部观察者
//
//   - async 是否异步通知
func (subject *Subject[T]) Notify(async bool) {
	if async {
		subject.observers.Range(func(key, value any) bool {
			go key.(Observer[T]).OnUpdate(subject.data)
			return true
		})
	} else {
		subject.observers.Range(func(key, value any) bool {
			key.(Observer[T]).OnUpdate(subject.data)
			return true
		})
	}
}

// UpdateAndNotify 更新自身状态，并通知全部观察者
//
//   - data 传入更新的数据
//   - async 是否异步通知
func (subject *Subject[T]) UpdateAndNotify(data T, async bool) {
	subject.Update(data)
	subject.Notify(async)
}

// NewSubject 创建一个被观察主题
func NewSubject[T any]() *Subject[T] {
	return &Subject[T]{
		observers: sync.Map{},
	}
}