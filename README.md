# trade-app-backend锛堣璁剧簿绠€鐗堬級

鏈粨搴撳疄鐜颁竴涓簩鎵嬩氦鏄撳钩鍙板悗绔殑璇捐绮剧畝鐗?MVP锛屽綋鍓嶅凡缁忓叿澶囦粠娉ㄥ唽鐧诲綍鍒板晢鍝佸彂甯冦€佷笅鍗曘€佸畬鎴愯鍗曠殑鏈€灏忎氦鏄撻棴鐜€?
## 褰撳墠浜や粯鑼冨洿

宸插畬鎴愰樁娈碉細
- 闃舵 1锛氳剼鎵嬫灦涓庢湰鍦板彲杩愯
- 闃舵 2锛氶厤缃€佹棩蹇椼€侀敊璇拰涓棿浠?- 闃舵 3锛歅ostgreSQL 涓庤縼绉讳綋绯?- 闃舵 4锛氳璇佷笌鐢ㄦ埛璧勬枡
- 闃舵 5锛氬晢鍝佷笌鍒嗙被
- 闃舵 6锛氭渶灏忚鍗曢棴鐜?- 闃舵 7锛歄penAPI銆丷EADME銆丏ocker 鏀跺熬

褰撳墠瀹炵幇鐨勪笟鍔¤兘鍔涳細
- 鐢ㄦ埛娉ㄥ唽銆佺櫥褰曘€丣WT Access Token
- `GET/PATCH /api/v1/users/me`
- 鍒嗙被鍒楄〃
- 鍟嗗搧鍒涘缓銆佸垪琛ㄣ€佽鎯呫€佺紪杈戙€佹垜鐨勫晢鍝?- 璁㈠崟鍒涘缓銆佸彇娑堛€佸畬鎴愩€佹垜鐨勮鍗?
## 浠庨浂鍚姩

### 1. 鍑嗗鐜鍙橀噺

澶嶅埗鏍蜂緥鏂囦欢锛?
```powershell
Copy-Item .env.example .env
```

绋嬪簭浼氶粯璁ゅ皾璇曞姞杞介」鐩牴鐩綍涓嬬殑 `.env` 鏂囦欢銆?
鑷冲皯纭杩欎簺鍙橀噺鍙敤锛?- `JWT_SECRET`
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`

### 2. 鍚姩 PostgreSQL

```powershell
docker compose up -d
```

### 3. 鎵ц鏁版嵁搴撹縼绉?
```powershell
go run ./cmd/migrate up
go run ./cmd/migrate version
```

棰勬湡鐗堟湰锛歚version=4 dirty=false`

### 4. 鍚姩 API

```powershell
go run ./cmd/api
```

### 5. 楠岃瘉鍋ュ悍妫€鏌?
```powershell
Invoke-WebRequest http://127.0.0.1:8080/healthz
Invoke-WebRequest http://127.0.0.1:8080/readyz
```

## Docker 鍚姩鏂瑰紡

鏋勫缓闀滃儚锛?
```powershell
docker build -t trade-app-backend .
```

杩愯瀹瑰櫒鏃惰鑷娉ㄥ叆鐜鍙橀噺锛屼緥濡傦細

```powershell
docker run --rm -p 8080:8080 --env-file .env trade-app-backend
```

## 甯哥敤鍛戒护

```bash
make fmt
make test
make run
make migrate-up
make migrate-down
make migrate-version
```

## 鏍稿績鎺ュ彛

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`
- `GET /api/v1/categories`
- `POST /api/v1/listings`
- `GET /api/v1/listings`
- `GET /api/v1/listings/{id}`
- `PATCH /api/v1/listings/{id}`
- `GET /api/v1/users/me/listings`
- `POST /api/v1/orders`
- `POST /api/v1/orders/{id}/cancel`
- `POST /api/v1/orders/{id}/pay`n- `POST /api/v1/orders/{id}/ship`n- `POST /api/v1/orders/{id}/receive``
- `GET /api/v1/users/me/orders`

瀹屾暣濂戠害瑙侊細`api/openapi/openapi.yaml`

## 鏈€灏忚仈璋冪ず渚?
### 1. 娉ㄥ唽鍗栧涓庝拱瀹?
```json
POST /api/v1/auth/register
{
  "email": "seller@example.com",
  "password": "secret123",
  "nickname": "seller"
}
```

```json
POST /api/v1/auth/register
{
  "email": "buyer@example.com",
  "password": "secret123",
  "nickname": "buyer"
}
```

### 2. 鍗栧鍒涘缓鍟嗗搧

璇锋眰澶达細`Authorization: Bearer <seller_token>`

```json
POST /api/v1/listings
{
  "category_id": 1,
  "title": "ThinkPad X1 Carbon",
  "description": "Lightly used business laptop",
  "price_cents": 450000,\n  "image_urls": [\n    "https://example.com/image-1.jpg",\n    "https://example.com/image-2.jpg"\n  ],\n  "publish": true
}
```

### 3. 涔板涓嬪崟

璇锋眰澶达細`Authorization: Bearer <buyer_token>`

```json
POST /api/v1/orders
{
  "listing_id": 1
}
```

### 4. 涔板瀹屾垚璁㈠崟

璇锋眰澶达細`Authorization: Bearer <buyer_token>`

```text
POST /api/v1/orders/1/complete
```

### 5. 鏌ヨ鎴戠殑璁㈠崟

璇锋眰澶达細`Authorization: Bearer <buyer_token>`

```text
GET /api/v1/users/me/orders?page=1&page_size=20
```

## 鏂囨。涓庨獙璇佽鏄?
- OpenAPI 宸插缓绔嬶細`api/openapi/openapi.yaml`
- `openapi lint` / `openapi validate`锛氬綋鍓嶄粨搴撴湭寮曞叆瀵瑰簲宸ュ叿锛岄樁娈?7 鎸夎姹傛槑纭烦杩?- 鎵€鏈夋帴鍙ｆ枃妗ｄ粎瑕嗙洊褰撳墠宸茬粡瀹炵幇鐨勮兘鍔涳紝娌℃湁棰濆澹版槑鏈潵闃舵鎺ュ彛
