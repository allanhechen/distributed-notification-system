package utils

import (
	"context"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func concurrentSetField(l *LogState, key string, value any) {
	l.setField(key, value)
}

func TestLogStateConcurrent(t *testing.T) {
	l := &LogState{}
	var wg sync.WaitGroup

	wg.Go(func() {
		concurrentSetField(l, "key1", "value1")
	})
	wg.Go(func() {
		concurrentSetField(l, "key2", 2)
	})
	wg.Go(func() {
		concurrentSetField(l, "key3", []interface{}{})
	})
	wg.Go(func() {
		concurrentSetField(l, "key4", "value4")
	})
	wg.Go(func() {
		concurrentSetField(l, "key5", 5)
	})
	wg.Go(func() {
		concurrentSetField(l, "key6", []int{1, 2, 3, 4, 5})
	})

	wg.Wait()

	expected := map[string]any{
		"key1": "value1",
		"key2": 2,
		"key3": []interface{}{},
		"key4": "value4",
		"key5": 5,
		"key6": []int{1, 2, 3, 4, 5},
	}
	snapshot := l.Snapshot()

	// set a field after that should not appear in the snapshot
	concurrentSetField(l, "key7", 7)

	if diff := cmp.Diff(snapshot, expected); diff != "" {
		t.Errorf("got %v, wanted %v", snapshot, expected)
	}

}

func TestLogStateAddField(t *testing.T) {
	l := &LogState{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, LoggedState, l)
	AddField(ctx, "key", "value")

	wanted := map[string]any{"key": "value"}
	snapshot := l.Snapshot()
	if diff := cmp.Diff(snapshot, wanted); diff != "" {
		t.Errorf("got %v, wanted %v", snapshot, wanted)
	}

}
