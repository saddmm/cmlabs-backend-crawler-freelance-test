# API Crawler Website

## Persyaratan Sistem

- [Go](https://go.dev/dl/) versi 1.22 atau lebih baru.
- Koneksi internet untuk mengunduh _binary_ Chromium pada saat pertama kali dijalankan.

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
   _(Catatan: Pada saat pertama dijalankan, aplikasi akan otomatis mengunduh binary browser Playwright yang dibutuhkan.)_

## Penggunaan API

Secara _default_, server berjalan pada `http://localhost:5000`.

### 1. Crawl URL

**Endpoint:** `POST /api/crawl`

**Headers:**

- `Content-Type: application/json`

**Request Body:**

```json
{
  "urls": ["https://cmlabs.co", "https://sequence.day", "https://go.dev"]
}
```

**Contoh cURL:**

```bash
curl -X POST http://localhost:5000/api/crawl \
-H "Content-Type: application/json" \
-d '{
  "urls": [
    "https://cmlabs.co",
    "https://sequence.day"
  ]
}'
```

**Contoh Respons Data:**

```json
{
  "code":    200,
			"status":  "OK",
			"message": "Crawling process completed",
			"data": [
          {
            "url": "https://cmlabs.co",
            "file_url": "http://localhost:5000/outputs/cmlabs.co_1775547348.html",
            "status": "Success"
          },
          {
            "url": "https://sequence.day",
            "file_url": "http://localhost:5000/outputs/sequence.day_1775547345.html",
            "status": "Success"
          },
          ...
      ],
			"meta": {
				"total_urls":     3,
				"success_count":  3,
				"failed_count":   0,
				"execution_time": "4.476348s",
			},
}

```

### 2. Melihat File HTML Hasil Crawl

Setelah API selesai memproses, file HTML akan disimpan di dalam direktori `/outputs`.
Anda bisa melihat HTML mentahnya langsung lewat browser menggunakan URL yang dikembalikan pada _field_ `file_url` di respons JSON:
**Contoh:** `http://localhost:5000/outputs/cmlabs.co_1775547348.html`

## Cara Kerja

1. **Setup Fiber:** Aplikasi membuat _endpoint_ `POST /api/crawl` untuk menerima _array_ berisi sekelompok URL.
2. **Goroutine:** Untuk setiap URL di dalam _request payload_, sebuah Goroutine baru dibuat untuk meneruskan pekerjaan secara kongkuren (paralel).
3. **Eksekusi Playwright:** Setiap Goroutine membuka sebuah tab Chromium _headless_ baru.
4. **Network Idle:** Bot menunggu status `WaitUntilStateNetworkidle` (yaitu ketika tidak ada aktivitas koneksi _network_ dalam 500ms terakhir). Ini menjamin bahwa semua panggilan API dari JavaScript selesai dan DOM HTML dimuat secara menyeluruh.
5. **Koleksi Data:** `sync.WaitGroup` bertugas untuk menunggu hingga seluruh kumpulan program Goroutine agar mengakhiri prosesnya.
6. **Hasil:** Sistem menyimpan data dari `htmlContent` ke dalam _folder_ statik `/outputs` lalu menyajikan respons lengkapnya kepada User.
