# Kong DB-less, เทมเพลต: scripts\render-gateway.ps1 แทนช่วงระหว่าง marker BEGIN/END ใต้ jwt_secrets ด้วยรายการ public key ทุกตัวใน keys\*.pub (หนึ่งรายการต่อ kid)
# แล้วเขียนไฟล์จริงไว้ที่ gateway/.rendered/kong.yml (git-ignore) Kong ได้แค่ public key ไม่มี private key
_format_version: "3.0"

consumers:
  - username: claims-web
    # ปลั๊กอิน jwt หา key จาก kid ใน header ของ token
    jwt_secrets:
      # __JWT_KEYS_BEGIN__
      - key: example-kid # รายการตัวอย่างนี้ถูกแทนที่ตอน render
        algorithm: RS256
        rsa_public_key: __PUBLIC_KEY__
      # __JWT_KEYS_END__

services:
  # login / refresh / logout: ไม่มี JWT ที่ทางเข้า (ก็คือจุดที่ขอ token) แต่จำกัดอัตราแน่นกว่า
  - name: auth-api
    url: http://api:8080/auth
    connect_timeout: 3000
    read_timeout: 10000
    write_timeout: 10000
    retries: 0
    routes:
      - name: auth-api
        paths: ["/api/auth"]
        strip_path: true # /api/auth/login -> /auth/login
        methods: [POST, OPTIONS]
        preserve_host: false
    plugins:
      - name: cors
        config:
          origins: ["http://localhost:3000"]
          methods: [POST, OPTIONS]
          headers: [Accept, Authorization, Content-Type, X-Requested-With] # Authorization: logout ส่ง access token ไปให้ API เพิกถอน
          exposed_headers: [Retry-After, X-RateLimit-Remaining-Minute] # ให้ UI อ่านเวลารอตอนโดน 429
          credentials: true # จำเป็นให้ browser ส่ง/รับ cookie refresh ข้าม origin (localhost:3000 -> localhost:8000)
          max_age: 3600
          preflight_continue: false
      # 20 ครั้ง/นาที/IP: กันเดารหัสผ่าน (UI ปกติ login/refresh น้อยกว่านี้มาก)
      - name: rate-limiting
        config:
          minute: 20
          limit_by: ip
          policy: local
          fault_tolerant: true
          hide_client_headers: false
          error_code: 429
          error_message: Too many authentication attempts
      - name: request-size-limiting
        config:
          allowed_payload_size: 8
          size_unit: kilobytes

  - name: claims-api
    url: http://api:8080
    connect_timeout: 3000
    read_timeout: 10000
    write_timeout: 10000
    retries: 0 # POST ที่ retry เองหลัง timeout อาจซ้ำ ปล่อยให้ client ส่งซ้ำพร้อม Idempotency-Key
    routes:
      - name: claims-api
        paths: ["/api"]
        strip_path: true # /api/claims -> /claims
        methods: [GET, POST, OPTIONS]
        preserve_host: false
    plugins:
      # ตรวจลายเซ็น + exp ที่ทางเข้า (ไม่ผ่านไม่ถึง API) API ตรวจซ้ำและยังเป็นคนตัดสินเจ้าของข้อมูล
      - name: jwt
        config:
          key_claim_name: kid # เลือก public key จาก kid ใน header: เก็บ key เก่าไว้คู่กับ key ใหม่ได้ระหว่างหมุน key
          claims_to_verify: [exp]
          maximum_expiration: 3600
          run_on_preflight: false # OPTIONS (CORS preflight) ไม่มี Authorization
          header_names: [authorization]
          cookie_names: []
          uri_param_names: [] # ไม่รับ token ใน query string (หลุดลง log ง่าย)
      - name: cors
        config:
          origins: ["http://localhost:3000"]
          methods: [GET, POST, OPTIONS]
          headers: [Accept, Authorization, Content-Type, Idempotency-Key]
          exposed_headers: [Idempotent-Replayed, Retry-After]
          credentials: false
          max_age: 3600
          preflight_continue: false
      # 120 ครั้ง/นาที/IP (หน้า detail poll ทุก 2 วินาที = 30 ครั้ง/นาที)
      - name: rate-limiting
        config:
          minute: 120
          limit_by: ip
          policy: local
          fault_tolerant: true
          hide_client_headers: false
          error_code: 429
          error_message: API rate limit exceeded
      - name: request-size-limiting
        config:
          allowed_payload_size: 1
          size_unit: megabytes
