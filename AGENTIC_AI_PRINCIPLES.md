# 🤖 Agentic AI Teknik Prensipleri

> Bu doküman, Kap-App projesinde AI iş ortağımızla (agentic AI) çalışırken
> uyulması gereken temel prensipleri, sınırları ve kalite standartlarını tanımlar.
>
> **Özet**: Agentic AI bir teknik iş ortağıdır — mimari kararlar, kod sahipliği
> ve kalite kontrolü her zaman insanda kalır.

---

## Kapsam

Kod üretimi, dosya ve sistem erişimi, CI/CD entegrasyonu, önyüz mimarisi,
SOLID, Clean Architecture, i18n ve kütüphane standartları.

---

## 1. Niyeti ve Bağlamı Önceden Tanımla

> Agentic araçlara ne yapması gerektiğini değil, ne başarması gerektiğini söyle.

- `"Şu fonksiyonu yaz"` yerine `"Şu problemi çöz, şu kısıtlar dahilinde, şu pattern'ı kullanarak"` de.
- Implementation detayını AI yönetir; insan intent (niyet) verir.

---

## 2. Görevi Atomik Birimlere Böl

> Tek bir agentic çağrı = tek bir iyi tanımlı sorumluluk.

- Büyük ve belirsiz görevler sapma ve hata üretir.
- `"Auth sistemi yaz"` yerine `"JWT doğrulama middleware'i yaz, refresh token şimdilik kapsam dışı"` gibi daralt.

---

## 3. Her Çıktıyı Kritik Gözle İncele — Kör Güvenme

> AI hatalı kod üretir.

- Özellikle async hatalar, edge case boşlukları ve güvenlik açıkları gözden kaçar.
- Her PR'da: `"Bunu ben yazmış olsam reviewer olarak ne sorardım?"` sorusunu sor.
- Sen onaylayan değil, **mühendis gözüyle inceleyen tarafsın**.

---

## 4. Test-First Düşün, Sonra Yazdır

> AI'a önce testleri yazdır, sonra implementasyonu.

- Önce beklenen davranışı ve edge case'leri tanımla.
- Bu hem çıktı kalitesini artırır hem de seni kör bir kabul döngüsünden çıkarır.

---

## 5. Kod Sahipliği Sende — AI'da Değil

> Üretilen her satır sana aittir.

- AI'ın ürettiği kodu **kendi yazmış gibi açıklayabilmelisin**.
- Açıklayamazsan, henüz o kodu commit etme.
- Anlamadığın kodu production'a taşıma.

---

## 6. Çalış → Gözlemle → Yeniden Yönlendir Döngüsünü Kur

> İlk çıktı nadiren nihai.

- Agentic döngülerde sapmanı erken yakala.
- Hata mesajını, stack trace'i ve beklentini somut şekilde geri besle.
- Her iterasyonu bilinçli yönet.

---

## 7. Guardrail'ları Önceden Koy

> Lint, type check, güvenlik taraması gibi otomatik ağlar olmadan agentic akış tehlikelidir.

- AI'ın ürettiği kodu CI/CD pipeline'a dahil etmeden önce **statik analiz** ve **güvenlik taraması** katmanlarını ekle.

---

## 8. Mimari Kararları AI'ya Bırakma

> Bağımlılık seçimi, veri modeli, servis sınırları gibi kararlar insana ait.

- AI mevcut codebase'i, ekip standartlarını ve teknik borcunu bilmez.
- Bu kararları **sen ver**, AI sadece uygulasın.

---

## 9. Versiyon Kontrolünü Bilinçli Kullan

> AI ile çalışırken commit'leri daha küçük ve sık at.

- Bir şey bozulduğunda hangi AI değişikliğinin neden sapma yarattığını izleyebilmek kritik değer taşır.
- **Her anlamlı adımı ayrı commit'e ayır.**

---

## 10. Context Penceresini Temiz Tut

> Uzun bir agentic oturumda AI önceki kararlarla çelişmeye başlar.

- Kritik kararları, kısıtları ve kabul edilmiş pattern'ları **her oturumun başında sistematik olarak besle**.
- Bir `project brief` dosyası oluştur ve her oturuma ekle.

---

## 11. "Neden?" Sorusunu Sorma Alışkanlığı Edin

> AI'a ürettiği yapıyı açıklat.

- `"Bu yaklaşımı neden seçtin, alternatifleri neler?"` sorusu hem çıktıyı iyileştirir hem seni körleşmekten korur.
- AI'ın halüsinasyon üretip üretmediğini erken teşhis ettirir.

---

## 12. Clean Architecture Prensiplerine Önyüzde de Uy

> UI katmanını business logic'ten ayır.

- Component'leri **"ne iş yaptığına"** göre değil **"kime ait olduğuna"** göre organize et.
- Bağımlılıklar her zaman içe doğru akmalı: `UI → domain → altyapı`
- **Pragmatik ol**; basit değişiklikler için fazla katman geçmek gerekmez.

---

## 13. SOLID Prensiplerini Component Tasarımına Taşı

| Prensip | Açıklama |
|---------|----------|
| **S** — Single Responsibility | Her component tek bir şeyi iyi yapsın |
| **O** — Open/Closed | Mevcut kodu değiştirmek yerine genişletilebilir yap |
| **L** — Liskov Substitution | Alt component'ler üst contract'ı bozmamalı |
| **I** — Interface Segregation | Prop arayüzlerini küçük ve odaklı tut |
| **D** — Dependency Inversion | Concrete implementasyona değil soyutlamaya bağlan; servisi inject et |

---

## 14. i18n Altyapısını Baştan Kur, Sonraya Bırakma

> "Şimdilik tek dil" diye başlamak en yaygın teknik borç kaynaklarından biridir.

- i18n altyapısını **ilk component'ten itibaren** ekle; hard-coded string bırakma.
- AI'a bunu kural olarak ver: her ürettiği component'te metinler mutlaka **çeviri anahtarı** üzerinden geçmeli.

---

## 15. Belirli Bir UI Kütüphanesine Bağlı Kal ve Bunu AI'a Bildir

> Her oturumun başında hangi UI kütüphanesini, hangi sürümü ve hangi pattern'ları
> tercih ettiğini AI'a açıkça söyle.

- Aksi hâlde AI her seferinde farklı bir kütüphaneyle çözüm önerir ve codebase tutarsızlaşır.
- Standartlarını bir `project brief` dosyasında tut.

---

## Özet: İnsan + AI İş Birliği Modeli

```
🧠 İnsan (Sen)                        🤖 AI (İş Ortağı)
─────────────────────────────         ─────────────────────────────
• Niyet / intent tanımlar             • Kod üretir
• Mimari kararlar alır                • Pattern'ları uygular
• Kaliteyi denetler                   • Alternatif sunar
• Kod sahipliğini üstlenir            • Edge case'leri gösterir
• Guardrail'ları koyar                • Hataları minimize eder
• "Neden?" sorusunu sorar             • Açıklama yapar
• Commit stratejisini belirler        • Atomik adımları yönetir
```

> **Altın Kural**: AI'ın ürettiği her satırı anlamadan production'a taşıma.
> Açıklayamadığın kodu commit etme.
