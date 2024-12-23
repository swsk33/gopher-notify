package gopher_notify

import (
	"sync"
	"time"
)

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
	// 防抖的时间间隔，设为0表示不使用防抖
	debounceDuration time.Duration
	// 是否正在防抖间隔冷却时间内
	// 若为true，则主题变化且Notify调用时，也不会通知观察者
	debounceFlag bool
	// 互斥锁，保证数据和定时器的安全操作
	lock sync.Mutex
}

// NewSubject 创建一个被观察主题
func NewSubject[T any]() *Subject[T] {
	return &Subject[T]{
		observers:        sync.Map{},
		debounceDuration: 0,
		debounceFlag:     false,
		lock:             sync.Mutex{},
	}
}

// NewSubjectWithDebounce 创建一个被观察主题，带有防抖机制
//
//   - duration 防抖间隔，0表示不使用防抖
//     若主题高频变化，就可能导致观察者被高频调用，出现资源浪费，可设定一个防抖间隔，在防抖时间间隔内出现的变化不会通知给观察者
//     例如设为 1*time.Second 观察者会在防抖时间1秒后收到通知，即使在1秒内主题多次更新状态
func NewSubjectWithDebounce[T any](duration time.Duration) *Subject[T] {
	return &Subject[T]{
		observers:        sync.Map{},
		debounceDuration: duration,
		debounceFlag:     false,
		lock:             sync.Mutex{},
	}
}

// 通知全部观察者的逻辑
//
//   - async 是否异步通知
func (subject *Subject[T]) notifyObserver(async bool) {
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
	// 上锁
	subject.lock.Lock()
	defer subject.lock.Unlock()
	// 若处于防抖冷却时间内，则不进行通知
	if subject.debounceFlag {
		return
	}
	// 否则，执行通知
	subject.notifyObserver(async)
	// 设定防抖冷却，进入防抖状态
	if subject.debounceDuration > 0 {
		subject.debounceFlag = true
		go func() {
			time.Sleep(subject.debounceDuration)
			subject.debounceFlag = false
		}()
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