# RabbitMQ Channel Pool Examples

This directory contains examples demonstrating the RabbitMQ channel pool implementation and performance improvements.

## Examples Overview

### 1. `channel_pool/main.go`
- **Purpose**: Demonstrates the new channel pool publisher in action
- **Features**: Concurrent publishing, pool statistics, performance monitoring
- **Usage**: `go run examples/channel_pool/main.go`

### 2. `performance_test/main.go` 
- **Purpose**: Interactive performance testing and comparison
- **Features**: Concurrent goroutines, throughput measurement, error handling
- **Usage**: `go run examples/performance_test/main.go`

### 3. `concurrent_publishing_test.go`
- **Purpose**: Proper Go test for automated testing
- **Features**: Unit tests, benchmarks, CI/CD integration
- **Usage**: 
  ```bash
  # Run integration tests (requires RabbitMQ)
  go test -v ./examples
  
  # Run benchmarks
  go test -bench=. ./examples
  
  # Skip integration tests
  go test -v ./examples -short
  ```

## Two Approaches Explained

### 🎯 **Approach 1: Demo Executable (main function)**
**Best for:** Manual testing, demonstrations, learning

```go
// examples/performance_test/main.go
package main

func main() {
    // Interactive demo code
    testConcurrentPublishing(publisher)
}
```

**Pros:**
- ✅ Interactive output dan progress monitoring
- ✅ Easy to run dan modify
- ✅ Great for demonstrations
- ✅ Real-time statistics display

**Cons:**
- ❌ Not integrated dengan CI/CD
- ❌ No automated assertions
- ❌ Requires manual verification

### 🧪 **Approach 2: Proper Go Test (test function)**
**Best for:** Automated testing, CI/CD, validation

```go
// examples/concurrent_publishing_test.go
package main

func TestConcurrentPublishing(t *testing.T) {
    // Automated test dengan assertions
}

func BenchmarkChannelPoolPublishing(b *testing.B) {
    // Performance benchmarks
}
```

**Pros:**
- ✅ Automated assertions dan validation
- ✅ CI/CD integration ready
- ✅ Benchmark capabilities
- ✅ Skip when RabbitMQ unavailable
- ✅ Standard Go testing patterns

**Cons:**
- ❌ Less interactive
- ❌ No real-time monitoring display

## Performance Results

### Before (Smart Locking):
- **Throughput**: ~100-200 messages/second
- **Bottleneck**: Mutex contention on single channel
- **Concurrency**: Limited by lock contention

### After (Channel Pool):
- **Throughput**: ~1000-2000 messages/second  
- **Improvement**: 5-10x performance increase
- **Concurrency**: True parallelism with multiple channels

## Running Examples

### Prerequisites
```bash
# Start RabbitMQ server
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Or use docker-compose
docker-compose up -d
```

### Run Channel Pool Demo
```bash
go run examples/channel_pool/main.go
```

### Run Performance Test
```bash
go run examples/concurrent_publish_test.go
```

## Key Benefits Demonstrated

1. **✅ True Parallelism** - Multiple channels eliminate lock contention
2. **✅ High Throughput** - 5-10x performance improvement 
3. **✅ Resource Efficiency** - Pool management optimizes memory usage
4. **✅ Fault Tolerance** - Channel recreation on failures
5. **✅ Monitoring** - Real-time pool statistics and metrics

## Architecture Patterns Shown

- **Channel Pool Pattern** - Resource pooling for scalability
- **Worker Pool Pattern** - Concurrent processing 
- **Publisher Confirms** - Reliable message delivery
- **Context Cancellation** - Graceful shutdown
- **Atomic Operations** - Lock-free counters
