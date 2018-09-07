#浅谈 Golang sync 包的相关使用方法

Desc:Go sync 包的使用方法，sync.Mutex，sync.RMutex，sync.Once，sync.Cond，sync.Waitgroup
尽管 Golang 推荐通过 channel 进行通信和同步，但在实际开发中 sync 包用得也非常的多。另外 sync 下还有一个 atomic 包，提供了一些底层的原子操作（这里不做介绍）。本篇文章主要介绍该包下的锁的一些概念及使用方法。

整个包都围绕这 Locker 进行，这是一个 interface：
```golang
type Locker interface {
        Lock()
        Unlock()
}
```
只有两个方法，Lock() 和 Unlock()。

另外该包下的对象，在使用过之后，千万不要复制。

有许多同学不理解锁的概念，下面会一一介绍到：

为什么需要锁？
在并发的情况下，多个线程或协程同时去修改一个变量，可能会出现如下情况：
```golang
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var a = 0

    // 启动 100 个协程，需要足够大
    // var lock sync.Mutex
    for i := 0; i < 100; i++ {
        go func(idx int) {
            // lock.Lock()
            // defer lock.Unlock()
            a += 1
            fmt.Printf("goroutine %d, a=%d\n", idx, a)
        }(i)
    }

    // 等待 1s 结束主程序
    // 确保所有协程执行完
    time.Sleep(time.Second)
}
```
观察打印结果，是否出现 a 的值是相同的情况（未出现则重试或调大协程数），答案：是的。

显然这不是我们想要的结果。出现这种情况的原因是，协程依次执行：从寄存器读取 a 的值 -> 然后做加法运算 -> 最后写会寄存器。试想，此时一个协程取出 a 的值 3，正在做加法运算（还未写回寄存器）。同时另一个协程此时去取，取出了同样的 a 的值 3。最终导致的结果是，两个协程产出的结果相同，a 相当于只增加了 1。

所以，锁的概念就是，我正在处理 a（锁定），你们谁都别和我抢，等我处理完了（解锁），你们再处理。这样就实现了，同时处理 a 的协程只有一个，就实现了同步。

把上面代码里的注释取消掉再试下。

什么是互斥锁 Mutex？
什么是互斥锁？它是锁的一种具体实现，有两个方法：

func (m *Mutex) Lock()
func (m *Mutex) Unlock()
在首次使用后不要复制该互斥锁。对一个未锁定的互斥锁解锁将会产生运行时错误。

一个互斥锁只能同时被一个 goroutine 锁定，其它 goroutine 将阻塞直到互斥锁被解锁（重新争抢对互斥锁的锁定）。如：
```golang
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    ch := make(chan struct{}, 2)

    var l sync.Mutex
    go func() {
        l.Lock()
        defer l.Unlock()
        fmt.Println("goroutine1: 我会锁定大概 2s")
        time.Sleep(time.Second * 2)
        fmt.Println("goroutine1: 我解锁了，你们去抢吧")
        ch <- struct{}{}
    }()

    go func() {
        fmt.Println("groutine2: 等待解锁")
        l.Lock()
        defer l.Unlock()
        fmt.Println("goroutine2: 哈哈，我锁定了")
        ch <- struct{}{}
    }()

    // 等待 goroutine 执行结束
    for i := 0; i < 2; i++ {
        <-ch
    }
}
```
注意，平时所说的锁定，其实就是去锁定互斥锁，而不是说去锁定一段代码。也就是说，当代码执行到有锁的地方时，它获取不到互斥锁的锁定，会阻塞在那里，从而达到控制同步的目的。

什么是读写锁 RWMutex?
那么什么是读写锁呢？它是针对读写操作的互斥锁，读写锁与互斥锁最大的不同就是可以分别对 读、写 进行锁定。一般用在大量读操作、少量写操作的情况：

func (rw *RWMutex) Lock()
func (rw *RWMutex) Unlock()

func (rw *RWMutex) RLock()
func (rw *RWMutex) RUnlock()
由于这里需要区分读写锁定，我们这样定义：

读锁定（RLock），对读操作进行锁定
读解锁（RUnlock），对读锁定进行解锁
写锁定（Lock），对写操作进行锁定
写解锁（Unlock），对写锁定进行解锁
在首次使用之后，不要复制该读写锁。不要混用锁定和解锁，如：Lock 和 RUnlock、RLock 和 Unlock。因为对未读锁定的读写锁进行读解锁或对未写锁定的读写锁进行写解锁将会引起运行时错误。

如何理解读写锁呢？

同时只能有一个 goroutine 能够获得写锁定。
同时可以有任意多个 gorouinte 获得读锁定。
同时只能存在写锁定或读锁定（读和写互斥）。
也就是说，当有一个 goroutine 获得写锁定，其它无论是读锁定还是写锁定都将阻塞直到写解锁；当有一个 goroutine 获得读锁定，其它读锁定任然可以继续；当有一个或任意多个读锁定，写锁定将等待所有读锁定解锁之后才能够进行写锁定。所以说这里的读锁定（RLock）目的其实是告诉写锁定：有很多人正在读取数据，你给我站一边去，等它们读（读解锁）完你再来写（写锁定）。

使用例子：
```golang
package main

import (
    "fmt"
    "math/rand"
    "sync"
)

var count int
var rw sync.RWMutex

func main() {
    ch := make(chan struct{}, 10)
    for i := 0; i < 5; i++ {
        go read(i, ch)
    }
    for i := 0; i < 5; i++ {
        go write(i, ch)
    }

    for i := 0; i < 10; i++ {
        <-ch
    }
}

func read(n int, ch chan struct{}) {
    rw.RLock()
    fmt.Printf("goroutine %d 进入读操作...\n", n)
    v := count
    fmt.Printf("goroutine %d 读取结束，值为：%d\n", n, v)
    rw.RUnlock()
    ch <- struct{}{}
}

func write(n int, ch chan struct{}) {
    rw.Lock()
    fmt.Printf("goroutine %d 进入写操作...\n", n)
    v := rand.Intn(1000)
    count = v
    fmt.Printf("goroutine %d 写入结束，新值为：%d\n", n, v)
    rw.Unlock()
    ch <- struct{}{}
}
```
WaitGroup 例子
WaitGroup 用于等待一组 goroutine 结束，用法很简单。它有三个方法：

func (wg *WaitGroup) Add(delta int)
func (wg *WaitGroup) Done()
func (wg *WaitGroup) Wait()
Add 用来添加 goroutine 的个数。Done 执行一次数量减 1。Wait 用来等待结束：
```golang
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    for i, s := range seconds {
        // 计数加 1
        wg.Add(1)
        go func(i, s int) {
            // 计数减 1
            defer wg.Done()
            fmt.Printf("goroutine%d 结束\n", i)
        }(i, s)
    }
    
    // 等待执行结束
    wg.Wait()
    fmt.Println("所有 goroutine 执行结束")
}
注意，wg.Add() 方法一定要在 goroutine 开始前执行哦。
```
Cond 条件变量
Cond 实现一个条件变量，即等待或宣布事件发生的 goroutines 的会合点。

type Cond struct {
    noCopy noCopy
  
    // L is held while observing or changing the condition
    L Locker
  
    notify  notifyList
    checker copyChecker
}
它会保存一个通知列表。

func NewCond(l Locker) *Cond
func (c *Cond) Broadcast()
func (c *Cond) Signal()
func (c *Cond) Wait()
Wait 方法、Signal 方法和 Broadcast 方法。它们分别代表了等待通知、单发通知和广播通知的操作。

我们来看一下 Wait 方法：

func (c *Cond) Wait() {
    c.checker.check()
    t := runtime_notifyListAdd(&c.notify)
    c.L.Unlock()
    runtime_notifyListWait(&c.notify, t)
    c.L.Lock()
}
它的操作为：加入到通知列表 -> 解锁 L -> 等待通知 -> 锁定 L。其使用方法是：

c.L.Lock()
for !condition() {
    c.Wait()
}
... make use of condition ...
c.L.Unlock()
举个例子：
```golang
// Package main provides ...
package main

import (
    "fmt"
    "sync"
    "time"
)

var count int = 4

func main() {
    ch := make(chan struct{}, 5)

    // 新建 cond
    var l sync.Mutex
    cond := sync.NewCond(&l)

    for i := 0; i < 5; i++ {
        go func(i int) {
            // 争抢互斥锁的锁定
            cond.L.Lock()
            defer func() {
                cond.L.Unlock()
                ch <- struct{}{}
            }()

            // 条件是否达成
            for count > i {
                cond.Wait()
                fmt.Printf("收到一个通知 goroutine%d\n", i)
            }
            
            fmt.Printf("goroutine%d 执行结束\n", i)
        }(i)
    }

    // 确保所有 goroutine 启动完成
    time.Sleep(time.Millisecond * 20)
    // 锁定一下，我要改变 count 的值
    fmt.Println("broadcast...")
    cond.L.Lock()
    count -= 1
    cond.Broadcast()
    cond.L.Unlock()

    time.Sleep(time.Second)
    fmt.Println("signal...")
    cond.L.Lock()
    count -= 2
    cond.Signal()
    cond.L.Unlock()

    time.Sleep(time.Second)
    fmt.Println("broadcast...")
    cond.L.Lock()
    count -= 1
    cond.Broadcast()
    cond.L.Unlock()

    for i := 0; i < 5; i++ {
        <-ch
    }
}
Pool 临时对象池
sync.Pool 可以作为临时对象的保存和复用的集合。其结构为：

type Pool struct {
    noCopy noCopy

    local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
    localSize uintptr        // size of the local array

    // New optionally specifies a function to generate
    // a value when Get would otherwise return nil.
    // It may not be changed concurrently with calls to Get.
    New func() interface{}
}
```
func (p *Pool) Get() interface{}
func (p *Pool) Put(x interface{})
新键 Pool 需要提供一个 New 方法，目的是当获取不到临时对象时自动创建一个（不会主动加入到 Pool 中），Get 和 Put 方法都很好理解。

深入了解过 Go 的同学应该知道，Go 的重要组成结构为 M、P、G。Pool 实际上会为每一个操作它的 goroutine 相关联的 P 都生成一个本地池。如果从本地池 Get 对象的时候，本地池没有，则会从其它的 P 本地池获取。因此，Pool 的一个特点就是：可以把由其中的对象值产生的存储压力进行分摊。

它有着以下特点：

Pool 中的对象在仅有 Pool 有着唯一索引的情况下可能会被自动删除（取决于下一次 GC 执行的时间）。
goroutines 协程安全，可以同时被多个协程使用。
GC 的执行一般会使 Pool 中的对象全部移除。

那么 Pool 都适用于什么场景呢？从它的特点来说，适用与无状态的对象的复用，而不适用与如连接池之类的。在 fmt 包中有一个很好的使用池的例子，它维护一个动态大小的临时输出缓冲区。

官方例子：
```golang
package main

import (
    "bytes"
    "io"
    "os"
    "sync"
    "time"
)

var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func timeNow() time.Time {
    return time.Unix(1136214245, 0)
}

func Log(w io.Writer, key, val string) {
    // 获取临时对象，没有的话会自动创建
    b := bufPool.Get().(*bytes.Buffer)
    b.Reset()
    b.WriteString(timeNow().UTC().Format(time.RFC3339))
    b.WriteByte(' ')
    b.WriteString(key)
    b.WriteByte('=')
    b.WriteString(val)
    w.Write(b.Bytes())
    // 将临时对象放回到 Pool 中
    bufPool.Put(b)
}

func main() {
    Log(os.Stdout, "path", "/search?q=flowers")
}

打印结果：
2006-01-02T15:04:05Z path=/search?q=flowers
Once 执行一次
使用 sync.Once 对象可以使得函数多次调用只执行一次。其结构为：

type Once struct {
    m    Mutex
    done uint32
}

func (o *Once) Do(f func())
用 done 来记录执行次数，用 m 来保证保证仅被执行一次。只有一个 Do 方法，调用执行。

package main

import (
    "fmt"
    "sync"
)

func main() {
    var once sync.Once
    onceBody := func() {
        fmt.Println("Only once")
    }
    done := make(chan bool)
    for i := 0; i < 10; i++ {
        go func() {
            once.Do(onceBody)
            done <- true
        }()
    }
    for i := 0; i < 10; i++ {
        <-done
    }
}

# 打印结果
Only once
```