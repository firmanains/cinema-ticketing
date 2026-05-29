# A.2 constant Solusi Sistem Pemilihan Tempat Duduk

## 1. Sistem pemilihan tempat duduk (high concurrency)

Tantangan utama adalah banyak user bisa memilih kursi yang sama secara bersamaan. Solusi menggunakan dua layer:

**Layer 1 constant Redis sebagai fast lock (sebelum masuk queue)**

Saat user memilih kursi, Booking Service langsung melakukan atomic lock ke Redis:

```
SETNX seat:{showtime_id}:{seat_id} {user_id} EX 300
```

Jika kursi sudah dikunci user lain, Redis langsung mengembalikan gagal dan sistem merespons `409 Conflict` tanpa perlu menyentuh database. Ini memastikan sistem tetap responsif meski diakses ribuan user sekaligus, karena Redis beroperasi in-memory dengan latensi sub-millisecond.

Jika berhasil, kursi terkunci selama 5 menit (sesuai TTL), kemudian event dikirim ke Kafka dan user mendapat respons `202 Accepted constant booking pending`.

**Layer 2 constant PostgreSQL dengan SELECT FOR UPDATE (source of truth)**

Consumer yang memproses event dari Kafka tetap melakukan validasi akhir ke database menggunakan `SELECT FOR UPDATE` sebelum commit. Ini memastikan konsistensi data meskipun ada edge case di layer Redis.

Pola UX yang digunakan adalah *eventual consistency* constant, jadi user tidak langsung mendapat "booking sukses", melainkan status *pending* terlebih dahulu. Konfirmasi akhir dikirim via WebSocket setelah consumer selesai memproses.

---

## 2. Sistem restok tiket yang telah terjual

Ada dua mekanisme restok yang berjalan otomatis:

**a. User tidak jadi memesan (timeout)**

Redis TTL pada seat lock akan habis otomatis setelah 5 menit jika user tidak menyelesaikan pembayaran. Kursi kembali tersedia tanpa perlu intervensi manual. Di sisi database, Consumer secara berkala melakukan cleanup:

```sql
UPDATE booking_seats SET status = 'available'
WHERE booking_id IN (
  SELECT id FROM bookings
  WHERE status = 'pending' AND expires_at < NOW()
)
```

**b. User membatalkan pesanan**

User menerbitkan event `booking.cancelled` yang dikonsumsi oleh Consumer. Consumer kemudian mengupdate status booking di database dan menghapus Redis key untuk kursi tersebut, sehingga kursi langsung tersedia kembali bagi user lain.

---

## 3. Refund dan pembatalan dari pihak bioskop

Ketika bioskop membatalkan satu jadwal tayang, dampaknya bisa mengenai ratusan booking sekaligus. Untuk menghindari bottleneck, proses ini dilakukan secara batch melalui Kafka:

```
Admin batalkan showtime_id: X
  → Publish event showtime.cancelled ke Kafka
    → Consumer batch process:
        1. UPDATE semua kursi showtime tersebut → AVAILABLE
        2. DEL semua Redis key terkait showtime
        3. INSERT bulk refund_records
        4. Publish event refund.initiated per booking
        5. Notifikasi user via WebSocket
```

Tabel `refunds` menyimpan kolom `initiated_by` yang membedakan apakah pembatalan berasal dari `user` atau `cinema`, sehingga tidak perlu tabel terpisah untuk dua jenis pembatalan ini.

Rekonsiliasi dengan payment gateway (Midtrans / Xendit) dilakukan menggunakan `payments.gateway_ref` yang tersimpan di tabel payments.
