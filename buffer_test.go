package pubsub

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func TestBufferWrite(t *testing.T) {
	buffer := NewBuffer(3, 1)

	want := []interface{}{}
	if got := buffer.Read(); !reflect.DeepEqual(want, got) {
		t.Errorf("want empty buffer read %v, got %v", want, got)
	}

	buffer.Write("one")
	want = append(want, "one")
	if got := buffer.Read(); !reflect.DeepEqual(want, got) {
		t.Errorf("want sparse buffer read %v, got %v", want, got)
	}

	buffer.Write("two")
	buffer.Write("three")
	want = append(want, "two", "three")
	if got := buffer.Read(); !reflect.DeepEqual(want, got) {
		t.Errorf("want full buffer read %v, got %v", want, got)
	}

	buffer.Write("four")
	want = append(want[1:], "four")
	if got := buffer.Read(); !reflect.DeepEqual(want, got) {
		t.Errorf("want wrapped buffer read %v, got %v", want, got)
	}
}

func TestBufferReadTo(t *testing.T) {
	buffer := NewBuffer(3, 1)
	donec := make(chan struct{})
	want := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}

	got := []string{}
	var rfn ReaderFunc = func(v interface{}) bool {
		got = append(got, v.(string))
		if len(got) == len(want) {
			close(donec)
			return false
		}
		return true
	}
	buffer.ReadTo(rfn)

	for _, v := range want {
		buffer.Write(v)
	}

	<-donec
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want buffer read to %v, got %v", want, got)
	}
}

func TestBufferWriteSlice(t *testing.T) {
	buffer := NewBuffer(3, 1)
	donec := make(chan struct{})
	want := []interface{}{"A", "B", "C", "D", "E"}

	got := []interface{}{}
	var rfn ReaderFunc = func(v interface{}) bool {
		got = append(got, v)
		if len(got) == len(want) {
			close(donec)
			return false
		}
		return true
	}
	buffer.ReadTo(rfn)

	buffer.WriteSlice(want)

	<-donec
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want buffer read to %v, got %v", want, got)
	}

	got = buffer.Read()
	if !reflect.DeepEqual(want[2:], got) {
		t.Errorf("want buffer read %v, got %v", want[2:], got)
	}
}

func TestBufferFullReadTo(t *testing.T) {
	buffer := NewBuffer(3, 1)
	donec := make(chan struct{})
	data := []interface{}{"A", "B", "C", "D", "E", "F", "G", "H", "I"}

	buffer.WriteSlice(data[:5])

	got := []interface{}{}
	var rfn ReaderFunc = func(v interface{}) bool {
		got = append(got, v)
		if len(got) == len(data[5:]) {
			close(donec)
			return false
		}
		return true
	}

	want := data[2:5]
	if s := buffer.FullReadTo(rfn); !reflect.DeepEqual(want, s) {
		t.Errorf("want full read %v, got %v", want, s)
	}

	buffer.WriteSlice(data[5:])

	<-donec
	want = data[5:]
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want full read func %v, got %v", want, got)
	}

}

func TestConcurrentReadTo(t *testing.T) {
	n, m := 1024, runtime.NumCPU()-1
	buffer := NewBuffer(n, m)

	data := make([]interface{}, n*m)
	for i := range data {
		data[i] = i
	}

	gots := make([][]interface{}, m)

	readywg := sync.WaitGroup{}
	readywg.Add(m)

	donewg := sync.WaitGroup{}
	donewg.Add(m)

	for i := range gots {
		go func(i int) {
			got := make([]interface{}, 0, len(data))
			var rfn ReaderFunc = func(v interface{}) bool {
				got = append(got, v)
				if len(got) == len(data) {
					gots[i] = got
					donewg.Done()
					return false
				}
				return true
			}

			buffer.ReadTo(rfn)
			readywg.Done()
		}(i)
	}

	readywg.Wait()
	buffer.WriteSlice(data)
	donewg.Wait()

	for _, got := range gots {
		if !reflect.DeepEqual(data, got) {
			t.Errorf("want buffer read to %v, got %v", data, got)
		}
	}
}

var buffr *Buffer

func BenchmarkSmallBuffer(b *testing.B) {
	n := (b.N + 1) / 2
	if n < 2 {
		n = 2
	}

	m := runtime.GOMAXPROCS(-1)
	buffer := NewBuffer(n, m)
	buffr = buffer

	data := make([]interface{}, b.N)
	for i := range data {
		data[i] = i
	}

	donewg := sync.WaitGroup{}
	donewg.Add(m)

	for i := 0; i < m; i++ {
		go func() {
			defer donewg.Done()
			count := 0
			buffer.ReadTo(func(v interface{}) bool {
				count++
				return count == len(data)
			})
		}()
	}

	b.SetBytes(int64(m))
	b.ReportAllocs()
	b.ResetTimer()

	buffer.WriteSlice(data)
	donewg.Wait()
}

func BenchmarkLargeBuffer(b *testing.B) {
	m := runtime.GOMAXPROCS(-1)
	buffer := NewBuffer(b.N*2, m)

	data := make([]interface{}, b.N)
	for i := range data {
		data[i] = i
	}

	donewg := sync.WaitGroup{}
	donewg.Add(m)

	for i := 0; i < m; i++ {
		go func() {
			defer donewg.Done()
			count := 0
			buffer.ReadTo(func(v interface{}) bool {
				count++
				return count == len(data)
			})
		}()
	}

	b.SetBytes(int64(m))
	b.ReportAllocs()
	b.ResetTimer()

	buffer.WriteSlice(data)
	donewg.Wait()
}
