# Development Progress Track

## Current Phase: Multi-Threaded I/O Architecture

**Status:** IN PROGRESS

### Completed Features
- Basic TCP Server handling connections.
- RESP Parser and Writer implementations.
- Global in-memory dictionary with `sync.RWMutex`.
- ThreadPool base implementation created.
- Implement cross-platform worker pool for processing connections.
- Offload parsing and writing to worker threads.
- Refactor Database execution to run on a single, dedicated goroutine.
- Remove `sync.RWMutex` from Database as thread-safety is guaranteed by the single executor.
- Replace single dictionary map with more advanced data structures.
- Support string, lists, dicts types properly.
- Implement AOF (Append Only File) Persistence.
- Implement RDB Snapshotting.

### Current Tasks
- Epoll / IO Multiplexing core implementation.
- Move away from `net.Listener` blocking and use multiplexing.

### Future Roadmap
- Dispatch socket read events to the ThreadPool.
- ThreadPool parses RESP command and sends to a central Go channel (Command Queue).
- Dedicated executor goroutine pops from Command Queue, executes, and pushes result to response queue.
- ThreadPool workers handle serialization and writing to client sockets.
