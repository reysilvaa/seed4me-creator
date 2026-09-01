# Seed4Me Auto Account Creator (Go Edition)

Tool automasi pembuatan akun Seed4Me instan menggunakan **Go Standard Library** (zero external dependencies), **CatchMail API** (`https://api.catchmail.io/api/v1`), **Auto Proxy Rotator**, dan konfigurasi terpusat **`config.json`**.

---

## Fitur

1. **1-Click Auto Create**: Registrasi otomatis, polling email CatchMail API, dan aktivasi 7 hari free trial.
2. **CatchMail & Custom Domain**: Mendukung `@catchmail.io` bawaan serta custom domain tanpa batas (cukup arahkan MX DNS ke `smtp.catchmail.io`).
3. **Multi-Worker Concurrency**: Buat banyak akun secara paralel menggunakan opsi worker pool (`-c`).
4. **Auto-Rotate IP & Tor Support**: Rotasi otomatis proxy publik dan deteksi otomatis Tor SOCKS5 lokal (`127.0.0.1:9050`).
5. **Output Bersih & Thread-Safe**: Akun tersimpan rapi dan sinkron di `accounts.json` & `accounts.txt`.

---

## Cara Pakai

```bash
# Buat 1 akun instan:
./run.sh    # (atau run.bat di Windows)

# Buat beberapa akun sekaligus (misal 5 akun):
./run.sh -n 5

# Buat akun secara paralel (misal 5 akun dengan 3 worker paralel):
./run.sh -n 5 -c 3

# Gunakan file konfigurasi kustom:
./run.sh -config "my_config.json"

# Pakai proxy manual:
./run.sh -proxy "http://ip:port"

# Gunakan promo code tertentu:
./run.sh -promo "KODEPROMO"
```

---

## Konfigurasi (`config.json`)

```json
{
  "count": 1,
  "concurrency": 1,
  "promo_code": "",
  "proxy": "",
  "tor_socks": "127.0.0.1:9050",
  "email_domain": "catchmail.io",
  "json_file": "accounts.json",
  "txt_file": "accounts.txt"
}
```

### Setup Custom Domain (Bebas Blokir Anti-Bot)
Agar akun tidak terblokir filter temporary email Seed4Me:
1. Masuk ke DNS Manager domain Anda (Cloudflare / Namecheap / dll).
2. Tambahkan DNS Record:
   - **Type**: `MX`
   - **Host**: `@` (atau subdomain)
   - **Value**: `smtp.catchmail.io`
   - **Priority**: `10`
3. Masukkan domain Anda ke `config.json`: `"email_domain": "domainanda.com"`.
4. Semua email masuk ke `*@domainanda.com` otomatis terbaca oleh CatchMail API.
