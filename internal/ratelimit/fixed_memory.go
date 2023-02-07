package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"
)

/*
Простая реализация ограничителя скорости на основе фиксированного окна.
Используется окно, равное n секундам. Каждый входящий запрос увеличивает счётчик для этого окна.
Если счётчик превышает некое пороговое значение, запрос отбрасывается.

Одиночный всплеск трафика вблизи границы окна может привести к удвоению количества обработанных запросов,
поскольку он разрешает запросы как для текущего, так и для следующего окна в течение короткого промежутка времени.
*/

// FixedMemory реализация FixedWin, в качестве хранилища использующая ОЗУ компьютера.
type FixedMemory struct {
	buckets map[string]*bucket
	mu      sync.Mutex
}

func NewFixedMemory() *FixedMemory {
	return &FixedMemory{
		buckets: make(map[string]*bucket),
	}
}

// ExceedLimit проверяет достигнут ли лимит в бакете с кодом bucketCode.
// true - достигнут лимит, запрос отбрасывается.
func (fm *FixedMemory) ExceedLimit(bucketCode string, limits Limits) (bool, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if err := limits.Valid(); err != nil {
		return false, err
	}
	res := fm.getBucket(bucketCode, &limits).takeCount()

	return !res, nil
}

// ResetBucket сбрасывает бакет с кодом bucketCode.
func (fm *FixedMemory) ResetBucket(bucketCode string) (bool, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if item := fm.getBucket(bucketCode, nil); item != nil {
		item.kill()
		<-item.done()
		delete(fm.buckets, bucketCode)

		return true, nil
	}

	return false, nil
}

// Destroy останавливает все бакеты.
func (fm *FixedMemory) Destroy() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	l := len(fm.buckets)
	if l == 0 {
		return nil
	}
	wg := sync.WaitGroup{}
	wg.Add(l)
	for _, item := range fm.buckets {
		go func(item *bucket) {
			<-item.done()
			wg.Done()
		}(item)
	}
	for _, item := range fm.buckets {
		item.kill()
	}
	wg.Wait()
	fm.buckets = make(map[string]*bucket)

	return nil
}

// Capacity количество бакетов.
func (fm *FixedMemory) Capacity() int {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return len(fm.buckets)
}

// BucketNames получить имена бакетов.
func (fm *FixedMemory) BucketNames() []string {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	result := make([]string, 0, len(fm.buckets))
	for bucketCode := range fm.buckets {
		result = append(result, bucketCode)
	}
	return result
}

// getBucket полчить бакет по name, если бакет не найдет и cfg != nil, метод создаст бакет.
func (fm *FixedMemory) getBucket(name string, cfg *Limits) *bucket {
	item, ok := fm.buckets[name]
	if ok || cfg == nil {
		return item
	}
	item = newBucket(fm, name, *cfg)
	fm.buckets[name] = item
	item.makeCounting()

	return item
}

type bucket struct {
	owner *FixedMemory
	// name имя бакета
	name string
	// counter счетчик обращений
	counter int64
	// limits параметры ограничений
	limits Limits
	// touched timestamp последнего
	touched int64

	doneCh chan struct{}
	killCh chan struct{}
	mu     sync.Mutex
}

func newBucket(fm *FixedMemory, name string, cfg Limits) *bucket {
	return &bucket{
		owner:  fm,
		name:   name,
		limits: cfg,
		doneCh: make(chan struct{}),
		killCh: make(chan struct{}),
	}
}

func (b *bucket) done() <-chan struct{} {
	return b.doneCh
}

func (b *bucket) takeCount() bool {
	atomic.AddInt64(&b.counter, 1)
	atomic.StoreInt64(&b.touched, time.Now().UnixNano())

	return atomic.LoadInt64(&b.counter) <= b.limits.Limit
}

// kill принудительная остановка.
func (b *bucket) kill() {
	b.killCh <- struct{}{}
}

// makeCounting вызывает в рутине и в фоне отслеживает свое состояние и при необходимости удаляется.
func (b *bucket) makeCounting() {
	go func() {
		ticker := time.NewTicker(b.limits.Period)
		defer func() {
			ticker.Stop()
			close(b.killCh)
			close(b.doneCh)
		}()
		for {
			select {
			case <-b.killCh:
				return
			case <-ticker.C:
				b.mu.Lock()
				atomic.StoreInt64(&b.counter, 0)
				// если два периода нет новых сигналов в бакете, то удаляем его
				touched := atomic.LoadInt64(&b.touched)
				needKill := time.Now().UnixNano()-touched >= b.limits.Period.Nanoseconds()*2
				b.mu.Unlock()
				if needKill {
					go func() {
						_, _ = b.owner.ResetBucket(b.name)
					}()
				}
			}
		}
	}()
}
