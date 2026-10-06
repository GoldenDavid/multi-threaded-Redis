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
- Epoll / IO Multiplexing core implementation.
- Move away from `net.Listener` blocking and use multiplexing.

- Dispatch socket read events to the ThreadPool.
- ThreadPool parses RESP command and sends to a central Go channel (Command Queue).
- Dedicated executor goroutine pops from Command Queue, executes, and pushes result to response queue.
- ThreadPool workers handle serialization and writing to client sockets.
- Unify the Thread Pool Framework: Refactor the implicit worker goroutines into a polished, generic ThreadPool system.
- Implement Time-To-Live (TTL) & Background Expiration Engine for keys.

### Current Tasks
- Implement non-blocking writes (`OpWrite`) using the IO multiplexer to handle socket backpressure.

### Future Roadmap
- Expand core data structures and commands (e.g., Sets `SADD`, `SMEMBERS` and Sorted Sets `ZADD`, `ZRANGE`).
