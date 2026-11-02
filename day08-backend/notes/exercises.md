# Day 8 exercises

1. Kill the worker with `docker compose kill worker`, post 5 claims, start it again. Do all 5 get decided? Why no duplicates?
2. `docker compose exec valkey valkey-cli monitor` then GET the same claim twice. Which one hits the cache?
3. Change `OverLimit`, see which tests and the baseline fail, update the baseline on purpose.
4. Replay a DLQ message by hand: read it with kafka-console-consumer and produce the value back to claims.submitted. What happens to the claim?
5. Replace limit/offset with keyset pagination. Which tests need to change?
6. Add a FAILED status for poison claims (migration + domain + tests + baseline).
