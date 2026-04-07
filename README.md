# API Crawler Website


## Persyaratan Sistem

- [Go](https://go.dev/dl/) versi 1.22 atau lebih baru.
- Koneksi internet untuk mengunduh *binary* Chromium pada saat pertama kali dijalankan.

## Instalasi & Cara Menjalankan

1. Clone repositori ini (atau salin file-nya).
2. Instal dependensi:
   ```bash
   go mod tidy
   ```
3. Jalankan aplikasi:
   ```bash
   go run main.go
   ```
 *(Catatan: Pada saat pertama dijalankan, aplikasi akan otomatis mengunduh binary browser Playwright yang dibutuhkan.)*

## Penggunaan API

Secara *default*, server berjalan pada `http://localhost:5000`.

### 1. Crawl URL

**Endpoint:** `POST /api/crawl`

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "url": [
    "https://cmlabs.co",
    "https://sequence.day",
    "https://go.dev"
  ]
}
```

**Contoh cURL:**
```bash
curl -X POST http://localhost:5000/api/crawl \
-H "Content-Type: application/json" \
-d '{
  "url": [
    "https://cmlabs.co",
    "https://sequence.day"
  ]
}'
```

**Contoh Respons Data:**
```json
[
  {
    "url": "https://cmlabs.co",
    "file_url": "http://localhost:5000/outputs/cmlabs.co_1775547348.html",
    "status": "Success"
  },
  {
    "url": "https://sequence.day",
    "file_url": "http://localhost:5000/outputs/sequence.day_1775547345.html",
    "status": "Success"
  }
]
```

### 2. Melihat File HTML Hasil Crawl

Setelah API selesai memproses, file HTML akan disimpan di dalam direktori `/outputs`.
Anda bisa melihat HTML mentahnya langsung lewat browser menggunakan URL yang dikembalikan pada *field* `file_url` di respons JSON:
**Contoh:** `http://localhost:5000/outputs/cmlabs.co_1775547348.html`

## Cara Kerja

1. **Setup Fiber:** Aplikasi membuat *endpoint* `POST /api/crawl` untuk menerima *array* berisi sekelompok URL.
2. **Goroutine:** Untuk setiap URL di dalam *request payload*, sebuah Goroutine baru dibuat untuk meneruskan pekerjaan secara kongkuren (paralel).
3. **Eksekusi Playwright:** Setiap Goroutine membuka sebuah tab Chromium *headless* baru.
4. **Network Idle:** Bot menunggu status `WaitUntilStateNetworkidle` (yaitu ketika tidak ada aktivitas koneksi *network* dalam 500ms terakhir). Ini menjamin bahwa semua panggilan API dari JavaScript selesai dan DOM HTML dimuat secara menyeluruh.
5. **Koleksi Data:** `sync.WaitGroup` bertugas untuk menunggu hingga seluruh kumpulan program Goroutine agar mengakhiri prosesnya.
6. **Hasil:** Sistem menyimpan data dari `htmlContent` ke dalam *folder* statik `/outputs` lalu menyajikan respons lengkapnya kepada User.
