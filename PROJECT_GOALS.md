# 🏠 Kap-App Proje Hedefleri ve Özellikler

> Bu doküman, Kap-App uygulamasının geliştirme sürecinde hedeflenen tüm özellikleri,
> kullanıcı senaryolarını ve gelecek sürüm planlarını içermektedir.

---

## 📋 İçindekiler

1. [Giriş/Kayıt Sistemi](#1-girişkayıt-sistemi)
2. [Konum Tabanlı Özellikler](#2-konum-tabanlı-özellikler)
3. [Kullanıcı Ekleme (Aile/Topluluk)](#3-kullanıcı-ekleme-ailetopluluk)
4. [Kullanıcı Ayarları](#4-kullanıcı-ayarları)
5. [Ekstra Özellikler](#5-ekstra-özellikler)
6. [Gelecek Sürümler (Roadmap)](#6-gelecek-sürümler-roadmap)

---

## 1. Giriş/Kayıt Sistemi

### 1.1. Kullanıcı Adları (Görünen İsim)
- Kayıt olurken herkes kendi ismini alabilir.
- Bu isimler **kullanıcı eklemede kullanılamaz** (sadece görünen amaçlıdır).
- İsimler, diğer kullanıcıların göreceği **düz metin** olarak saklanır.

### 1.2. Rastgele Kullanıcı ID
- Kullanıcılar kayıt olduğunda **rastgele kelimelerden üretilen x haneli bir ID**'ye sahip olur.
- Bu ID, ailelere/topluluklara katılmak için kullanılır.

### 1.3. İki Adımlı Doğrulama
- Kayıt olurken **iki adımlı doğrulama (2FA)** zorunludur.
- E-posta ile doğrulama yapılacaktır (ör. **Resend** API kullanılarak).

---

## 2. Konum Tabanlı Özellikler

### 2.1. Market Yakını Bildirimi (Gelecek Sürüm)
- Kullanıcıdan **konum takip izni** alınır.
- Kullanıcı bir marketin yakınına geldiğinde **bildirim** gönderilir.
- Bildirim, kullanıcıya alınması gereken ürünler olduğunu hatırlatır.

---

## 3. Kullanıcı Ekleme (Aile/Topluluk)

### 3.1. ID ile Kullanıcı Ekleme
- Aileye/topluluğa kullanıcı eklemek, sistemin size verdiği **ID'ler** aracılığıyla yapılır.
- Kullanıcılar kendi ID'lerini **ayarlar kısmından** görüntüleyebilir.

### 3.2. QR Kod ile Kullanıcı Ekleme (Gelecek Sürüm)
- ID'lerin haricinde **QR kodlar** ile de kullanıcı eklenebilecektir.

### 3.3. Çoklu Aile/Topluluk Üyeliği
- Kullanıcı **birden fazla eve/topluluğa** mensup olabilir.
- Mensup olunan aileler **sol üst köşeden** seçilerek değiştirilebilir.

### 3.4. Otomatik Aile Geçişi (Gelecek Sürüm)
- Konum izleme sayesinde, hangi evin/topluluğun yakınında olduğu algılanarak **otomatik** olarak o kısma geçilir.
- Bu özellik **kullanıcının opsiyonundadır**:
  - **Otomatik mod**: Kullanıcı otomatiği seçtiğinde sol üstteki ikon sabit kalır.
  - **Manuel mod**: Kullanıcı el ile seçim yapar.
- İnternet çekmezse veya geçiş hızlıca olmazsa, kullanıcı oradan seçerek alacaklarını daha hızlı görebilir.

---

## 4. Kullanıcı Ayarları

### 4.1. Özel İstekler (Private Requests)
- Kullanıcılar birbirlerine **özel istekte** bulunabilir.
- Bu özel istekler **diğer kullanıcılar tarafından görülemez**.
- Özel istekler, kullanıcı imgesine tıklandığında görünen **geçmiş kısmında da gözükmez**.
- Bu özellik için **özel bir SQL** yazılacaktır (sıradan isteklerden tamamen ayrı).

---

## 5. Ekstra Özellikler

### 5.1. Evde Bulunanlar (Stok Takibi)
- "Evde bulunanlar" kısmı oluşturulacaktır.
- Alınması istenmemiş veya görülünce alınan ürünler buraya eklenerek görüntülenebilir.
- Ürünler **3 miktar seviyesinde** takip edilir:
  - ✅ **Var**
  - ⚠️ **Azaldı**
  - ❌ **Yok**

### 5.2. Tarif Öneri Sistemi (Gelecek Sürüm)
- Uygulama içerisine **tarif kısmı** eklenebilir.
- Evde bulunan ürünlerle yapılabilecek yemekler kullanıcılara sunulur.
- Kullanıcılar **forum benzeri** bir yapıda tarif paylaşabilir.
- Örnek: Kullanıcı "menemen" tarifi yazarken, tarifte soğan kullanılacağını ve evde bulunması gereken miktarı da belirtir.
- Öneriler, evdeki stok durumuna göre yapılır.

---

## 6. Gelecek Sürümler (Roadmap)

| Özellik | Açıklama | Planlanan Sürüm |
|---------|----------|-----------------|
| ✅ Konum bazlı market bildirimi | Market yakınında alışveriş hatırlatması | v2.0 |
| ✅ QR kod ile kullanıcı ekleme | ID yerine QR okutarak katılım | v2.0 |
| ✅ Otomatik aile geçişi | Konuma göre aktif ailenin değişmesi | v2.0 |
| ✅ Tarif öneri sistemi | Evdeki ürünlerle yapılabilecek tarifler | v3.0 |
| ✅ Tarif forumu | Kullanıcıların tarif paylaştığı topluluk alanı | v3.0 |

---

## 7. Production Hazırlık (Teknik Borç)

> **Dış değerlendirme**: "Mimarini, çoklu kiracılık (multi-tenant) altyapını ve Agentic AI vizyonunu çok beğendim.
> Ancak bu yapay zeka ajanlarının projeye dahil olabilmesi ve güvenli büyüme için bazı testler acilen eklenmeli."

### 7.1. Flutter Provider/State Testleri (Yapılacak)
- Provider'lar ayrı ayrı test edilebilir hale getirilecek (dependency injection)
- Her Provider için 3 temel test:
  1. Başarılı state geçişi
  2. Hata state'i (error handling)
  3. Loading state'i

### 7.2. Cross-Tenant Veritabanı İzolasyon Testleri (Yapılacak)
- Tüm SQL sorgularında `WHERE tenant_id = ...` kontrolü test edilecek
- Test senaryoları:
  - Kullanıcı A, Tenant B'nin verilerini görmemeli
  - Tenant üyesi olmayan kullanıcı işlem yapamamalı
  - Farklı tenant verileri karışmamalı

### 7.3. Backend Unit Test İyileştirmeleri
- Mevcut mock'ların kapsamı genişletilecek
- Edge case'ler eklenecek

### 7.4. CI/CD Pipeline Entegrasyonu (Gelecek)
- GitHub Actions ile tüm testlerin otomatik çalıştırılması
- Statik analiz ve güvenlik taraması

---

> 📝 **Not**: Bu doküman, proje geliştirme sürecinde güncellenmeye devam edecektir.
> Her yeni özellik ve karar, ilgili mimari dokümanlara (ARCHITECTURE_TOPOLOGY.md, task.md) da yansıtılacaktır.
> Ayrıca günlük çalışma logu için `SESSION_LOG.md` dosyası tutulmaktadır.
