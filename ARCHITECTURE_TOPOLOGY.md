# Kap-App Architecture and Topology

Bu doküman, Kap-App projesinin statik yazılım mimarisini, katmanlar arası veri akışını, çalışma zamanı ağ topolojisini ve temel tasarım prensiplerini içermektedir.

## 1. Temel Prensipler

Proje geliştirme süreci boyunca aşağıdaki temel mimari ve güvenlik prensiplerine sıkı sıkıya bağlı kalınacaktır:
- **Clean Architecture & SOLID:** Kod tabanında sorumlulukların ayrılması (Separation of Concerns), bağımlılıkların içe doğru akması ve modüler bir yapı hedeflenmektedir.
- **Multi-Tenant Veri İzolasyonu:** Her bir ailenin (veya kiracının) verileri diğerlerinden tamamen izole edilmiş bir yapıda saklanmalı ve sunulmalıdır.
- **Row-Level Security (RLS):** Veritabanı seviyesinde veri güvenliği sağlanmalı, kullanıcıların sadece yetkili oldukları satırlara erişebilmeleri garanti altına alınmalıdır.

---

## 2. Backend Katmanı (Go)

Backend servisi, Go standart kütüphaneleri ve modüler bir dizin yapısı kullanılarak geliştirilmektedir. Katmanlar arası veri akışı `Handler -> Service -> Repository` hiyerarşisi üzerinden yürütülür.

### 2.1. Yönlendirme (Routing) Akışı
- Yönlendirme işlemi Go standart kütüphanesindeki `http.ServeMux` ile sağlanmaktadır.
- Uygulama başlatılırken `cmd/api/main.go` dosyasında ana bir `ServeMux` örneği oluşturulur.
- Oluşturulan bu `mux` nesnesi, her modülün kendi `RegisterRoutes(mux *http.ServeMux)` metoduna aktarılarak endpoint'lerin modüler olarak bağlanması sağlanır.

### 2.2. Veritabanı ve Bağımlılık Enjeksiyonu (Dependency Injection)
- **Bağlantı Havuzu:** Veritabanı bağlantısı `github.com/jackc/pgx/v5/pgxpool` kullanılarak yönetilir.
- **İlklendirme:** `pgxpool` havuzu `internal/config/database.go` dosyasında uygulama başlangıcında oluşturulur ve yapılandırılır.
- **Enjeksiyon:** Oluşturulan `*pgxpool.Pool` havuzu, `main.go` üzerinden ilgili modüllerin repository katmanlarına (`NewRepository(db *pgxpool.Pool)`) enjekte edilir. Bu sayede servislerin ve handler'ların veritabanı işlemlerinden soyutlanması sağlanır.

---

## 3. Frontend Katmanı (Flutter)

Frontend uygulaması, Flutter ile geliştirilen, REST API ile iletişim kuran ve çoklu dil destekli (i18n) bir yapıdadır.

### 3.1. Ağ İstekleri ve İletişim (Network)
- Backend ile haberleşme, `lib/core/network/api_client.dart` dosyasında yer alan `ApiClient` sınıfı üzerinden merkezileştirilmiştir.
- `ApiClient`, veri katmanındaki (Data/Repository) farklı özellik (feature) modüllerine (örn. `product_repository.dart`, `tenant_repository.dart`) enjekte edilerek API çağrıları yönetilir.

### 3.2. Durum Yönetimi (State Management)
- Uygulamanın reaktif durum yönetimi, modern ve compile-time güvenli olan **Riverpod (`flutter_riverpod`)** mimarisi kullanılarak sağlanmaktadır.
- Modüllerin state yönetimleri `StateNotifierProvider` ve `NotifierProvider` yapıları üzerinden izole edilir. İş kuralları (Business Logic), UI katmanından tamamen bağımsız bir şekilde bu provider'lar içinde yürütülür ve veri değişiklikleri UI bileşenlerine reaktif olarak yansıtılır.


### 3.3. Çoklu Dil Desteği (i18n & Localization)
- Çoklu dil ve yerelleştirme (Localization) ayarları, Flutter'ın standart `flutter_localizations` altyapısı kullanılarak yapılandırılmıştır.
- Dil dosyaları, proje kökündeki `assets/lang/` dizininde `en.json` ve `tr.json` şeklinde tutulmaktadır. 
- Bu JSON dosyaları, uygulama genelinde i18n prensiplerine uygun şekilde metinlerin yönetilmesini sağlar.

---

## Sonuç

Kap-App projesi, sunucu ve istemci arasındaki bağımlılıkları minimize eden, sürdürülebilir, ölçeklenebilir ve güvenli bir iletişim topolojisine sahiptir. Backend tarafında Go'nun sade ve modüler yapısından yararlanılırken, Frontend tarafında Flutter'ın reaktif durum yönetimi ile API etkileşimleri sağlam bir mimari zemin üzerine oturtulmuştur.
