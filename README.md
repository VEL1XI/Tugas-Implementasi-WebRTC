# WebRTC Video Call Application

Aplikasi sederhana implementasi WebRTC menggunakan **Pion WebRTC** (Golang).

## 📋 Fitur

- ✅ Video Call multi-user (room-based)
- ✅ Signaling Server menggunakan WebSocket
- ✅ Data Channel untuk pengiriman pesan
- ✅ Toggle Audio/Video
- ✅ UI responsif dan modern

## 🏗️ Arsitektur

```
┌─────────────────┐     WebSocket      ┌──────────────────┐
│   Browser A     │◄──────────────────►│                  │
│   (WebRTC)      │     Signaling      │   Go Server      │
└─────────────────┘                    │   (Pion WebRTC)  │
                                       │                  │
┌─────────────────┐     WebSocket      │  - Signaling     │
│   Browser B     │◄──────────────────►│  - ICE Server    │
│   (WebRTC)      │     Signaling      │  - Data Channel  │
└─────────────────┘                    └──────────────────┘
         ▲                                      
         │         Peer-to-Peer                
         └──────── Media Stream ───────────────┘
```

## 🚀 Cara Menjalankan

### Prasyarat

- Go 1.21 atau lebih baru
- Browser modern (Chrome, Firefox, Edge)

### Langkah-langkah

1. **Clone atau masuk ke direktori project**

```bash
cd "d:\PENJAR\Tugas Implementasi WebRTC"
```

2. **Download dependencies**

```bash
go mod tidy
```

3. **Jalankan server**

```bash
go run main.go
```

4. **Buka browser**

- Video Call: http://localhost:8080
- Data Channel Demo: http://localhost:8080/datachannel.html

5. **Untuk testing video call**:
   - Buka 2 tab browser
   - Masukkan Room ID yang sama di kedua tab
   - Masukkan nama yang berbeda
   - Klik "Gabung Room"

## 📁 Struktur Project

```
webrtc-app/
├── main.go              # Server utama (Signaling + Pion WebRTC)
├── go.mod               # Go module
├── go.sum               # Dependencies checksum
├── README.md            # Dokumentasi
└── static/
    ├── index.html       # UI Video Call
    └── datachannel.html # Demo Data Channel
```

## 🔧 Teknologi yang Digunakan

### Backend (Go)
- **Pion WebRTC** - Library WebRTC untuk Go
- **Gorilla WebSocket** - WebSocket server untuk signaling

### Frontend
- **Vanilla JavaScript** - WebRTC API browser
- **HTML5** - getUserMedia, Video elements
- **CSS3** - Styling modern

## 📡 Alur Signaling

1. **User A** bergabung ke room
2. **User B** bergabung ke room yang sama
3. Server memberitahu User A tentang User B
4. **User A** membuat SDP Offer → dikirim ke server → diteruskan ke User B
5. **User B** membuat SDP Answer → dikirim ke server → diteruskan ke User A
6. Kedua peer bertukar ICE Candidates melalui server
7. Koneksi peer-to-peer terbentuk untuk media stream

## 🎯 Konsep WebRTC yang Diimplementasikan

### 1. Signaling
- Pertukaran SDP (Session Description Protocol)
- Pertukaran ICE Candidates
- Room management

### 2. Peer Connection
- RTCPeerConnection API
- Media tracks (audio/video)
- Connection state management

### 3. ICE (Interactive Connectivity Establishment)
- STUN servers (Google's public STUN)
- ICE candidate gathering
- NAT traversal

### 4. Data Channel
- RTCDataChannel untuk pesan teks
- Komunikasi bidirectional dengan server Pion

## 🔒 Catatan Keamanan

Ini adalah aplikasi demo. Untuk production, pertimbangkan:

- Menggunakan HTTPS/WSS
- Implementasi autentikasi
- Rate limiting
- TURN server untuk NAT yang lebih kompleks
- Validasi input yang lebih ketat

## 📚 Referensi

- [Pion WebRTC](https://github.com/pion/webrtc)
- [WebRTC API (MDN)](https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API)
- [WebRTC Samples](https://webrtc.github.io/samples/)

## 📝 Lisensi

MIT License - Bebas digunakan untuk pembelajaran.
