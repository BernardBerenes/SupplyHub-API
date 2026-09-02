# PRD — Authentication

## 1. Overview

### 1.1 Problem Statement

Aplikasi dikembangkan sebagai aplikasi pribadi untuk membantu pengelolaan produk dan pesanan dari toko.

Karena aplikasi tidak bersifat public dan tidak menggunakan konsep multi-tenant, akses terhadap aplikasi perlu dibatasi hanya kepada pengguna yang memiliki kredensial yang valid.

Authentication dibutuhkan untuk memastikan hanya pengguna yang terdaftar di dalam sistem yang dapat mengakses fitur aplikasi, seperti:

- Pengelolaan produk
- Pengelolaan pesanan
- Pengelolaan transaksi
- Tracking status pesanan
- Fitur lain yang akan dikembangkan di kemudian hari

Pada tahap awal, authentication dibuat sesederhana mungkin dengan hanya menyediakan proses **login** menggunakan username dan password.

### 1.2 Background and Context

Aplikasi tidak menyediakan proses registrasi secara mandiri. User dibuat atau dikelola secara langsung melalui database/application setup.

Authentication menggunakan **JSON Web Token (JWT)** sebagai access token.

Setelah login berhasil, server akan menghasilkan access token yang digunakan oleh client untuk mengakses endpoint yang membutuhkan authentication.

Access token memiliki masa berlaku selama **30 menit**.

Pada MVP tidak digunakan refresh token, authentication session, token blacklist, maupun mekanisme server-side token invalidation.

Dengan demikian, apabila token telah diterbitkan, token tetap dapat digunakan sampai:

1. Token expired.
2. Signature JWT tidak valid.
3. Token tidak memenuhi validasi JWT lainnya.

Logout pada tahap MVP dilakukan dengan menghapus access token dari sisi client.

---

# 2. Goals

## 2.1 Business Goals

### BG-01 — Membatasi akses aplikasi

Sistem harus memastikan bahwa hanya user yang memiliki username dan password yang valid yang dapat mengakses aplikasi.

### BG-02 — Menyediakan authentication yang sederhana

Authentication harus memiliki implementasi yang sederhana sehingga dapat digunakan sebagai fondasi untuk pengembangan fitur bisnis lainnya tanpa menambahkan kompleksitas yang belum dibutuhkan.

### BG-03 — Menjaga keamanan credential

Password user tidak boleh disimpan dalam bentuk plaintext.

Password harus disimpan menggunakan algoritma hashing **bcrypt**.

### BG-04 — Menggunakan authentication berbasis token

Sistem menggunakan JWT sebagai mekanisme authentication untuk endpoint yang membutuhkan akses terautentikasi.

---

## 2.2 User Goals

### UG-01 — User dapat melakukan login

User dapat masuk ke aplikasi menggunakan username dan password yang telah terdaftar.

### UG-02 — User mendapatkan access token

Setelah login berhasil, user mendapatkan JWT access token yang dapat digunakan untuk mengakses protected endpoint.

### UG-03 — User dapat mengakses aplikasi selama token valid

User dapat menggunakan fitur aplikasi selama access token masih valid.

### UG-04 — User dapat logout

User dapat melakukan logout dengan menghapus access token dari sisi client.

---

# 3. Use Cases

## 3.1 User Stories

### UC-01 — Login

**As a user**,

I want to login menggunakan username dan password,

so that I can access the application.

#### Preconditions

- User telah tersedia di database.
- User memiliki username yang valid.
- Password user telah disimpan dalam bentuk bcrypt hash.

#### Main Flow

1. User membuka halaman login.
2. User memasukkan username.
3. User memasukkan password.
4. Client mengirimkan request login ke server.
5. Server mencari user berdasarkan username.
6. Server melakukan verifikasi password menggunakan bcrypt.
7. Jika credential valid, server membuat JWT access token.
8. Server mengembalikan access token kepada client.
9. Client menyimpan access token.
10. User dapat mengakses protected endpoint.

#### Alternative Flow

Jika username tidak ditemukan:

1. Server menolak authentication.
2. Server mengembalikan HTTP `401 Unauthorized`.

Jika password tidak sesuai:

1. Server menolak authentication.
2. Server mengembalikan HTTP `401 Unauthorized`.

---

### UC-02 — Access Protected Resource

**As a logged-in user**,

I want to access protected endpoints using my JWT,

so that I can use application features securely.

#### Main Flow

1. Client memiliki access token.
2. Client mengirim request ke protected endpoint.
3. Client mengirimkan JWT pada HTTP Authorization header.

Contoh:

```
Authorization: Bearer <access_token>
```

1. Server mengambil JWT dari request.
2. Server melakukan validasi JWT.
3. Server memvalidasi signature.
4. Server memvalidasi expiration time.
5. Jika token valid, request diteruskan ke handler/controller.
6. Server mengembalikan response endpoint.

#### Alternative Flow

Jika token tidak tersedia:

```
401 Unauthorized
```

Jika token tidak valid:

```
401 Unauthorized
```

Jika token telah expired:

```
401 Unauthorized
```

---

### UC-03 — Logout

**As a logged-in user**,

I want to logout from the application,

so that my client no longer uses the current authentication token.

#### Main Flow

1. User memilih logout.
2. Client menghapus access token yang tersimpan.
3. User dianggap telah logout pada sisi client.
4. User harus melakukan login kembali untuk mendapatkan token baru.

#### Limitation

Pada MVP, server tidak menyimpan session atau daftar token aktif.

Oleh karena itu, server tidak dapat melakukan invalidasi terhadap JWT yang telah diterbitkan.

Apabila token belum expired, token tersebut secara teknis masih valid apabila digunakan kembali.

Access token akan menjadi tidak valid setelah masa berlaku 30 menit berakhir.

---

# 4. Requirements

## 4.1 Functional Requirements

### FR-01 — User Authentication

System harus menyediakan endpoint untuk melakukan login menggunakan:

- Username
- Password

Contoh request:

```
POST /login
Content-Type: application/json
```

```json
{
  "username": "username",
  "password": "password"
}
```

---

### FR-02 — Validate Username

System harus mencari user berdasarkan username yang diberikan.

Username harus bersifat unique.

Jika username tidak ditemukan, authentication harus gagal.

---

### FR-03 — Validate Password

System harus melakukan validasi password menggunakan bcrypt.

System tidak boleh melakukan perbandingan password menggunakan plaintext comparison.

Secara konseptual:

```
input password
       │
       ▼
bcrypt verify
       │
       ▼
stored bcrypt hash
```

---

### FR-04 — Generate JWT

Setelah username dan password berhasil diverifikasi, system harus menghasilkan JWT access token.

JWT minimal harus memiliki informasi:

- Subject/User ID (`sub`)
- Issued At (`iat`)
- Expiration (`exp`)

Contoh payload:

```json
{
  "sub": 1,
  "iat": 1756700000,
  "exp": 1756701800
}
```

Nilai aktual `iat` dan `exp` mengikuti waktu saat token dibuat.

---

### FR-05 — Access Token Expiration

Access token harus memiliki masa berlaku selama **30 menit** sejak token diterbitkan.

Setelah expiration time tercapai, token harus ditolak oleh server.

Response:

```
401 Unauthorized
```

---

### FR-06 — JWT Signature Validation

Setiap protected endpoint harus melakukan validasi signature JWT sebelum request diproses.

JWT dengan signature yang tidak valid harus ditolak.

---

### FR-07 — Protected Endpoint

Endpoint yang membutuhkan authentication harus menggunakan JWT middleware.

Request harus menyertakan:

```
Authorization: Bearer <access_token>
```

Request tanpa authentication token harus ditolak.

---

### FR-08 — Invalid Token

System harus menolak request apabila:

- JWT tidak memiliki format yang valid.
- JWT signature tidak valid.
- JWT telah expired.
- JWT tidak memiliki claim yang dibutuhkan.
- JWT menggunakan algoritma yang tidak diperbolehkan.

Response:

```
401 Unauthorized
```

---

### FR-09 — Logout

System harus menyediakan mekanisme logout pada sisi client.

Pada MVP, logout dilakukan dengan menghapus access token yang tersimpan di client.

Server tidak perlu menyimpan informasi session atau token.

---

### FR-10 — User Registration

System **tidak menyediakan endpoint registration**.

Tidak diperbolehkan melakukan:

```
POST /register
```

User dibuat melalui database atau mekanisme internal aplikasi.

---

### FR-11 — Refresh Token

System **tidak menggunakan refresh token**.

Setelah access token expired, user harus melakukan login kembali.

Flow:

```
Access Token
    │
    ▼
30 minutes
    │
    ▼
Expired
    │
    ▼
Login kembali
    │
    ▼
New Access Token
```

---

### FR-12 — Password Storage

Password harus disimpan dalam database dalam bentuk bcrypt hash.

Contoh:

```
password
↓
bcrypt
↓
$2y$...hashed-password...
```

Plaintext password tidak boleh disimpan dalam database.

---

## 4.2 Database Requirements

Authentication MVP hanya membutuhkan satu tabel utama:

```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL
);
```

### Field Definition

| Field | Type | Constraint | Description |
| --- | --- | --- | --- |
| `id` | BIGINT | PK | Unique identifier user |
| `username` | VARCHAR(100) | NOT NULL, UNIQUE | Username untuk login |
| `password` | VARCHAR(255) | NOT NULL | Bcrypt password hash |
| `created_at` | TIMESTAMP | NULL | Waktu user dibuat |
| `updated_at` | TIMESTAMP | NULL | Waktu user terakhir diperbarui |

Tidak diperlukan tabel:

- `auth_sessions`
- `refresh_tokens`
- `token_blacklist`
- `password_reset_tokens`
- `user_roles`

untuk MVP authentication.

---

## 4.3 API Requirements

### Login

```
POST /login
```

Request:

```json
{
  "username": "bernard",
  "password": "secret"
}
```

Success:

```
200 OK
```

```json
{
  "message": "Login successful",
  "data": {
    "access_token": "<jwt>"
  }
}
```

### Failed Login

```
401 Unauthorized
```

Response dapat menggunakan format:

```json
{
  "message": "Invalid username or password"
}
```

Pesan error tidak perlu membedakan apakah username tidak ditemukan atau password salah untuk menghindari user enumeration.

---

# 5. Non-Functional Requirements

## 5.1 Security

### NFR-01 — Password Hashing

Password harus menggunakan bcrypt.

Plaintext password tidak boleh disimpan di database maupun application log.

### NFR-02 — HTTPS

Authentication endpoint harus digunakan melalui HTTPS pada environment production.

Credential dan JWT tidak boleh dikirim melalui koneksi HTTP plaintext pada production.

### NFR-03 — JWT Secret

JWT signing secret tidak boleh di-hardcode di source code.

Secret harus disimpan melalui environment/configuration management.

Contoh:

```
JWT_SECRET=...
```

### NFR-04 — JWT Algorithm

System harus menggunakan algoritma JWT yang telah ditentukan secara eksplisit.

Server tidak boleh menerima algoritma JWT secara sembarangan dari token client.

### NFR-05 — Sensitive Data Logging

System tidak boleh mencatat:

- Password
- JWT access token
- JWT secret

ke application log.

---

## 5.2 Performance

### NFR-06 — Authentication Latency

Proses authentication harus memiliki response time yang cukup rendah untuk penggunaan aplikasi normal.

Bcrypt memang membutuhkan computational cost untuk meningkatkan keamanan password, sehingga cost factor harus dipilih dengan mempertimbangkan security dan performance.

---

## 5.3 Availability

### NFR-07 — Authentication Availability

Login harus tersedia selama application service dan database dapat diakses.

Apabila database tidak tersedia, server harus mengembalikan error yang sesuai tanpa mengekspos detail internal database.

---

## 5.4 Maintainability

### NFR-08 — Authentication Middleware

Validasi JWT harus diimplementasikan dalam middleware sehingga authentication logic tidak perlu diulang pada setiap endpoint.

Konsep:

```
Request
   │
   ▼
JWT Middleware
   │
   ├── Invalid → 401
   │
   ▼
Controller / Handler
   │
   ▼
Business Logic
```

---

# 6. Out of Scope

Fitur berikut tidak termasuk dalam MVP authentication:

- User registration
- Refresh token
- Server-side token invalidation
- Token blacklist
- Authentication session management
- Multi-factor authentication (MFA)
- Forgot password
- Reset password
- Social login
- OAuth
- Role-based access control (RBAC)
- Permission management
- Multi-tenancy
- Multiple organization/tenant
- Login menggunakan email
- Login menggunakan nomor telepon

Fitur-fitur tersebut dapat dipertimbangkan pada fase pengembangan berikutnya apabila kebutuhan aplikasi berkembang.

---

# 7. Authentication Flow

## Login

```
┌─────────────┐
│    Client   │
└──────┬──────┘
       │
       │ username + password
       ▼
┌─────────────┐
│ POST /login │
└──────┬──────┘
       │
       ▼
┌────────────────────┐
│ Find user by       │
│ username           │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ bcrypt verify      │
│ password           │
└─────────┬──────────┘
          │
          │ valid
          ▼
┌────────────────────┐
│ Generate JWT       │
│ exp = 30 minutes   │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ Return Access      │
│ Token              │
└────────────────────┘
```

## Access Protected Endpoint

```
Client
  │
  │ Authorization: Bearer JWT
  ▼
JWT Middleware
  │
  ├── Parse JWT
  ├── Verify signature
  ├── Verify expiration
  └── Validate claims
        │
        ├── Invalid → 401
        │
        ▼
   Controller
        │
        ▼
   Business Logic
```

## Token Expiration

```
Login
  │
  ▼
JWT issued
  │
  ├────────────── 30 minutes ──────────────┐
  │                                        │
  ▼                                        ▼
Valid                                  Expired
                                           │
                                           ▼
                                     401 Unauthorized
                                           │
                                           ▼
                                      Login again
```

---

# 8. Acceptance Criteria

### AC-01 — Successful Login

**Given** user dengan username dan password yang valid tersedia di database,

**When** user melakukan login,

**Then** server harus mengembalikan JWT access token dengan expiration 30 menit.

---

### AC-02 — Invalid Password

**Given** username tersedia tetapi password salah,

**When** user melakukan login,

**Then** server harus mengembalikan `401 Unauthorized`.

---

### AC-03 — Unknown Username

**Given** username tidak tersedia,

**When** user melakukan login,

**Then** server harus mengembalikan `401 Unauthorized`.

---

### AC-04 — Valid JWT

**Given** client memiliki JWT yang valid dan belum expired,

**When** client mengakses protected endpoint,

**Then** request harus dapat diproses oleh endpoint tersebut.

---

### AC-05 — Expired JWT

**Given** JWT telah melewati expiration time,

**When** client mengakses protected endpoint,

**Then** server harus mengembalikan `401 Unauthorized`.

---

### AC-06 — Invalid JWT

**Given** JWT memiliki signature yang tidak valid,

**When** client mengakses protected endpoint,

**Then** server harus mengembalikan `401 Unauthorized`.

---

### AC-07 — No Authentication

**Given** endpoint membutuhkan authentication,

**When** client tidak mengirimkan Authorization header,

**Then** server harus mengembalikan `401 Unauthorized`.

---

### AC-08 — Logout

**Given** user memiliki access token,

**When** user melakukan logout,

**Then** client harus menghapus access token sehingga request berikutnya tidak lagi menggunakan token tersebut.

---