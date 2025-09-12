package genericmap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextMap(t *testing.T) {
	cm := NewGenericMap[string, context.Context]()

	// Test initial state
	assert.Equal(t, 0, cm.Size())

	// Test Store and Load
	ctx1 := context.WithValue(context.Background(), "key1", "value1")
	cm.Store("msg1", ctx1)

	assert.Equal(t, 1, cm.Size())

	loadedCtx, exists := cm.Load("msg1")
	require.True(t, exists)
	assert.Equal(t, "value1", loadedCtx.Value("key1"))

	// Test Load non-existent
	_, exists = cm.Load("nonexistent")
	assert.False(t, exists)

	// Test Store multiple
	ctx2 := context.WithValue(context.Background(), "key2", "value2")
	cm.Store("msg2", ctx2)

	assert.Equal(t, 2, cm.Size())

	// Test LoadAndDelete
	loadedCtx, exists = cm.LoadAndDelete("msg1")
	require.True(t, exists)
	assert.Equal(t, "value1", loadedCtx.Value("key1"))
	assert.Equal(t, 1, cm.Size())

	// Test LoadAndDelete non-existent
	_, exists = cm.LoadAndDelete("nonexistent")
	assert.False(t, exists)

	// Test Delete
	cm.Delete("msg2")
	assert.Equal(t, 0, cm.Size())

	// Test Clear
	cm.Store("msg3", context.Background())
	cm.Store("msg4", context.Background())
	assert.Equal(t, 2, cm.Size())

	cm.Clear()
	assert.Equal(t, 0, cm.Size())
}

func TestContextMapConcurrency(t *testing.T) {
	cm := NewGenericMap[string, context.Context]()

	// Test concurrent access
	done := make(chan bool, 10)

	// Start 5 goroutines storing m
	for i := 0; i < 5; i++ {
		go func(id int) {
			ctx := context.WithValue(context.Background(), "id", id)
			cm.Store(string(rune('a'+id)), ctx)
			done <- true
		}(i)
	}

	// Start 5 goroutines loading m
	for i := 0; i < 5; i++ {
		go func(id int) {
			cm.Load(string(rune('a' + id)))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	assert.Equal(t, 5, cm.Size())
}
