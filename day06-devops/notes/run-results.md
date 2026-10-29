# Day 6 run results

Run at: 2026-10-29 18:45:14  Machine: LAPTOP-MICOTO

## Summary

| Step | Result |
|---|---|
| Tools present | PASS |
| Go: go get pgx + go mod tidy | PASS |
| Go: build with pgx driver (-tags pgx) | PASS |
| Go: vet (-tags pgx) | PASS |
| Go: unit tests (go test ./...) | PASS |
| Docker: build stage image | PASS |
| Docker: final image (distroless) | PASS |
| Docker: final image runs as non-root | PASS |
| Compose: up --build (db healthy, api answers /healthz) | PASS |
| Compose: create + list claim through API and Postgres | PASS |
| Compose: data survives api restart (stored in Postgres) | PASS |
| Compose: graceful shutdown log on stop | PASS |
| kind: create cluster | PASS |
| kind: load image | PASS |
| kubectl apply -k k8s | PASS |
| k8s: all pods Ready | PASS |
| k8s: delete api pod -> new pod created and Ready | PASS |
| k8s: Service reachable via port-forward (/healthz) | PASS |
| k8s: readiness failing on new pods (old pod keeps serving, maxUnavailable 0) | PASS |
| k8s: rollout undo recovers | PASS |
| k8s: rollout restart logs graceful shutdown of old pod | PASS |
| CI: actionlint on GitHub Actions file (static check, pipeline NOT executed) | PASS |

## Tools present  [PASS]

```
> docker version --format "docker {{.Client.Version}} / server {{.Server.Version}}"
docker 29.7.2 / server 29.7.2  (exit 0)

> go version
go version go1.26.1 windows/amd64  (exit 0)

> kind version
kind v0.33.0 go1.26.7 windows/amd64  (exit 0)

> kubectl version --client
Client Version: v1.36.1
Kustomize Version: v5.8.1  (exit 0)
```

## Go: go get pgx + go mod tidy  [PASS]

```


```

## Go: build with pgx driver (-tags pgx)  [PASS]

```

```

## Go: vet (-tags pgx)  [PASS]

```

```

## Go: unit tests (go test ./...)  [PASS]

```
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service	(cached)
?   	github.com/sitthichok-J/self-study-2026/day05-go-service/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day05-go-service/cmd/printsql	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/app	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/config	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/gen	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/handler	(cached)
?   	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/report	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day05-go-service/internal/strcat	(cached)
```

## Docker: build stage image  [PASS]

```
... (39 earlier lines omitted)
#12 exporting to image
#12 exporting layers done
#12 exporting manifest sha256:2eb4a035f4ffa6bf452ddd618c50643346d5468b70a63ab77f0b287a5d155dae done
#12 exporting config sha256:f4f8d0eb02de64207f489fdc3717febf0d6519477e72a609564ec27503528ba1 done
#12 exporting attestation manifest sha256:1a3c44690dadf46a9c60fca60e5bdbc9b46c9788976d4ef84c747466193258ee 0.0s done
#12 exporting manifest list sha256:5862119782a450ecdb9050c59d05d42d9afd15836f2ceac12f7d361647af0d45
#12 exporting manifest list sha256:5862119782a450ecdb9050c59d05d42d9afd15836f2ceac12f7d361647af0d45 0.0s done
#12 naming to docker.io/library/claims-api:build done
#12 unpacking to docker.io/library/claims-api:build done
#12 DONE 0.1s

View build details: docker-desktop://dashboard/build/desktop-linux/desktop-linux/x82dsky2qn3ocl41k5xln71rk
```

## Docker: final image (distroless)  [PASS]

```
... (21 earlier lines omitted)
#6 DONE 0.0s

#7 [stage-1 1/2] FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
#7 resolve gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab 0.0s done
#7 DONE 0.0s

#8 [internal] load build context
#8 transferring context: 1.27kB done
#8 DONE 0.0s

#9 [build 2/6] WORKDIR /src
#9 CACHED

#10 [build 3/6] COPY go.mod go.sum ./
#10 CACHED

#11 [build 6/6] RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -tags pgx -ldflags="-s -w" -o /out/api ./cmd/api
#11 CACHED

#12 [build 4/6] RUN go mod download
#12 CACHED

#13 [build 5/6] COPY . .
#13 CACHED

#14 [stage-1 2/2] COPY --from=build /out/api /api
#14 CACHED

#15 exporting to image
#15 exporting layers done
#15 exporting manifest sha256:e474100064d6a1922a04ccfe931adbc97bffc54111fc26bb986cc3694d72ffea done
#15 exporting config sha256:f57b74b37fd883583e0024eb7a41eab256cd93adba6f3cb4a0f678154917ba80 done
#15 exporting attestation manifest sha256:4c9f9290590215989e89c185d11676a250d46b39830ce16576a5bae692c2b543
#15 exporting attestation manifest sha256:4c9f9290590215989e89c185d11676a250d46b39830ce16576a5bae692c2b543 0.0s done
#15 exporting manifest list sha256:726e87189947b67ed7fe14b23cc110453407c83739a767a4aa12e635f610342c 0.0s done
#15 naming to docker.io/library/claims-api:dev done
#15 unpacking to docker.io/library/claims-api:dev done
#15 DONE 0.1s

View build details: docker-desktop://dashboard/build/desktop-linux/desktop-linux/l191cpc0ga8ga2m3lx56fal15

Image sizes:
claims-api:dev  21.5MB
claims-api:build  1.51GB
```

## Docker: final image runs as non-root  [PASS]

```
user=nonroot
```

## Compose: up --build (db healthy, api answers /healthz)  [PASS]

```
... (79 earlier lines omitted)
 Container claims-db-1 Creating 
 Container claims-db-1 Created 
 Container claims-api-1 Creating 
 Container claims-api-1 Created 
 Container claims-db-1 Starting 
 Container claims-db-1 Started 
 Container claims-db-1 Waiting 
 Container claims-db-1 Healthy 
 Container claims-api-1 Starting 
 Container claims-api-1 Started

NAME           IMAGE            COMMAND                  SERVICE   CREATED         STATUS                   PORTS
claims-api-1   claims-api:dev   "/api"                   api       5 seconds ago   Up Less than a second    0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
claims-db-1    postgres:16      "docker-entrypoint.s…"   db        5 seconds ago   Up 4 seconds (healthy)   0.0.0.0:5433->5432/tcp, [::]:5433->5432/tcp

/healthz -> 200 {"status":"ok"}

```

## Compose: create + list claim through API and Postgres  [PASS]

```
POST /claims -> 201 {"id":1,"patient_ref":"P001","hospital":"Test Hospital","amount_satang":150000,"status":"PENDING","created_at":"2026-10-29T11:45:34.23071611Z","updated_at":"2026-10-29T11:45:34.23071611Z"}

GET /claims -> 200 {"count":1,"items":[{"id":1,"patient_ref":"P001","hospital":"Test Hospital","amount_satang":150000,"status":"PENDING","created_at":"2026-10-29T11:45:34.230716Z","updated_at":"2026-10-29T11:45:34.230716Z"}]}

```

## Compose: data survives api restart (stored in Postgres)  [PASS]

```
GET /claims after restart -> 200 {"count":1,"items":[{"id":1,"patient_ref":"P001","hospital":"Test Hospital","amount_satang":150000,"status":"PENDING","created_at":"2026-10-29T11:45:34.230716Z","updated_at":"2026-10-29T11:45:34.230716Z"}]}

```

## Compose: graceful shutdown log on stop  [PASS]

```
{"time":"2026-10-29T11:45:33.908234467Z","level":"INFO","msg":"listening","addr":"[::]:8080","db":true}
{"time":"2026-10-29T11:45:33.961892977Z","level":"INFO","msg":"request","request_id":"2d7b9addf50702b6","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.04}
{"time":"2026-10-29T11:45:34.234214236Z","level":"INFO","msg":"request","request_id":"ec408c247ae88dd7","method":"POST","path":"/claims","status":201,"bytes":189,"duration_ms":3.705}
{"time":"2026-10-29T11:45:34.250034141Z","level":"INFO","msg":"request","request_id":"e1513aa6aed51928","method":"GET","path":"/claims","status":200,"bytes":207,"duration_ms":2.49}
{"time":"2026-10-29T11:45:34.472759747Z","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-29T11:45:34.472969256Z","level":"INFO","msg":"shutdown complete"}
{"time":"2026-10-29T11:45:35.24759576Z","level":"INFO","msg":"listening","addr":"[::]:8080","db":true}
{"time":"2026-10-29T11:45:41.293473499Z","level":"INFO","msg":"request","request_id":"63e057fa31831073","method":"GET","path":"/claims","status":200,"bytes":207,"duration_ms":2.606}
{"time":"2026-10-29T11:45:41.542206742Z","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-29T11:45:41.54244257Z","level":"INFO","msg":"shutdown complete"}
```

## kind: create cluster  [PASS]

```
... (8 earlier lines omitted)
 • Starting control-plane 🕹️  ...
 ✓ Starting control-plane 🕹️
 • Installing CNI 🔌  ...
 ✓ Installing CNI 🔌
 • Installing StorageClass 💾  ...
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-claims"
You can now use your cluster with:

kubectl cluster-info --context kind-claims

Thanks for using kind! 😊
```

## kind: load image  [PASS]

```
Image: "claims-api:dev" with ID "sha256:df24a7c8460447e554d221c81874e55f75ef2746e1289c14e5d902c02d06f0e2" not yet present on node "claims-control-plane", loading...
```

## kubectl apply -k k8s  [PASS]

```
namespace/claims created
configmap/claims-api-config created
configmap/postgres-initdb created
secret/claims-db created
service/claims-api created
service/postgres created
deployment.apps/claims-api created
statefulset.apps/postgres created
```

## k8s: all pods Ready  [PASS]

```
pod/claims-api-56dcbb6748-fshgq condition met
pod/claims-api-56dcbb6748-gjk8j condition met
pod/postgres-0 condition met

NAME                          READY   STATUS    RESTARTS      AGE
claims-api-56dcbb6748-fshgq   1/1     Running   3 (42s ago)   78s
claims-api-56dcbb6748-gjk8j   1/1     Running   3 (44s ago)   78s
postgres-0                    1/1     Running   0             78s


```

## k8s: delete api pod -> new pod created and Ready  [PASS]

```
old pod: claims-api-56dcbb6748-fshgq
seconds until a new Ready pod: 4.4

NAME                          READY   STATUS    RESTARTS   AGE
claims-api-56dcbb6748-kbg7h   1/1     Running   0          5s
claims-api-56dcbb6748-ldh6g   1/1     Running   0          5s
postgres-0                    1/1     Running   0          83s
```

## k8s: Service reachable via port-forward (/healthz)  [PASS]

```
-> 200 {"status":"ok"}

```

## k8s: readiness failing on new pods (old pod keeps serving, maxUnavailable 0)  [PASS]

```
patch: deployment.apps/claims-api patched

NAME                          READY   STATUS    RESTARTS   AGE
claims-api-56dcbb6748-kbg7h   1/1     Running   0          51s
claims-api-56dcbb6748-ldh6g   1/1     Running   0          51s
claims-api-694d8f494d-sn5j6   0/1     Running   0          40s
postgres-0                    1/1     Running   0          2m9s

Warning: v1 Endpoints is deprecated in v1.33+; use discovery.k8s.io/v1 EndpointSlice
NAME         ENDPOINTS                          AGE
claims-api   10.244.0.10:8080,10.244.0.9:8080   2m9s

endpoints still has an address: True
```

## k8s: rollout undo recovers  [PASS]

```
Warning: resource deployments/claims-api was previously managed with 'kubectl apply'. Rolling back will not update the kubectl.kubernetes.io/last-applied-configuration annotation, which may cause unexpected behavior on future 'kubectl apply' operations. Consider using 'kubectl apply' with your previous configuration file instead.
deployment.apps/claims-api rolled back
deployment "claims-api" successfully rolled out
```

## k8s: rollout restart logs graceful shutdown of old pod  [PASS]

```
Waiting for deployment "claims-api" rollout to finish: 1 out of 2 new replicas have been updated...
Waiting for deployment "claims-api" rollout to finish: 1 out of 2 new replicas have been updated...
Waiting for deployment "claims-api" rollout to finish: 1 out of 2 new replicas have been updated...
Waiting for deployment "claims-api" rollout to finish: 1 old replicas are pending termination...
Waiting for deployment "claims-api" rollout to finish: 1 old replicas are pending termination...
deployment "claims-api" successfully rolled out

last lines of followed log:
... (7 earlier lines omitted)
{"time":"2026-10-29T11:47:46.202101664Z","level":"INFO","msg":"request","request_id":"e009fd8cfab24482","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.033}
{"time":"2026-10-29T11:47:48.130536471Z","level":"INFO","msg":"request","request_id":"e0a49fe522c5cf4e","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.029}
{"time":"2026-10-29T11:47:51.201639862Z","level":"INFO","msg":"request","request_id":"eba4b46fcd82d07e","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.016}
{"time":"2026-10-29T11:47:56.202272731Z","level":"INFO","msg":"request","request_id":"0d0634dd0cdb46f7","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.016}
{"time":"2026-10-29T11:47:58.130275195Z","level":"INFO","msg":"request","request_id":"8725a793f9d65b5a","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.025}
{"time":"2026-10-29T11:48:01.202490622Z","level":"INFO","msg":"request","request_id":"1c4a2861f66a9dcb","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.03}
{"time":"2026-10-29T11:48:06.200320145Z","level":"INFO","msg":"request","request_id":"d9c030cbfb00bc1c","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.02}
{"time":"2026-10-29T11:48:08.128922847Z","level":"INFO","msg":"request","request_id":"cc9f5cd9039b1941","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.021}
{"time":"2026-10-29T11:48:11.200109902Z","level":"INFO","msg":"request","request_id":"2af38f97accabd5d","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.022}
{"time":"2026-10-29T11:48:16.201151036Z","level":"INFO","msg":"request","request_id":"b08ff3f903d851f7","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.028}
{"time":"2026-10-29T11:48:18.128769148Z","level":"INFO","msg":"request","request_id":"c2b5903846303585","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.045}
{"time":"2026-10-29T11:48:21.200898701Z","level":"INFO","msg":"request","request_id":"69ae6de2cbcbddc9","method":"GET","path":"/healthz","status":200,"bytes":16,"duration_ms":0.014}
{"time":"2026-10-29T11:48:25.380341062Z","level":"INFO","msg":"shutting down","timeout":"10s"}
{"time":"2026-10-29T11:48:25.380470138Z","level":"INFO","msg":"shutdown complete"}

```

## CI: actionlint on GitHub Actions file (static check, pipeline NOT executed)  [PASS]

```

(exit 0)
```


