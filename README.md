# Seed4Me Auto Account Creator (Go Edition)

Tool automasi pembuatan akun Seed4Me instan menggunakan **Go Standard Library** (zero dependencies), **CatchMail.io API**, dan **Auto Proxy Rotator**.

---

## Fitur

1. **1-Click Auto Create**: Registrasi otomatis, polling email CatchMail.io, dan aktivasi 7 hari trial.
2. **Auto-Rotate IP**: Menghindari rate limit Seed4Me dengan proxy publik aktif atau Tor lokal.
3. **Akun Tersimpan**: Disimpan otomatis di `accounts.txt` & `accounts.json`.

---

## Cara Pakai

```bash
# Buat 1 akun instan:
./run.sh

# Buat beberapa akun sekaligus (misal 3 akun):
./run.sh -n 3

# Pakai proxy sendiri jika punya:
./run.sh -proxy "http://ip:port"
```
