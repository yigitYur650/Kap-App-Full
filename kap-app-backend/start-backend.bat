@echo off
:: Kap-App Go Backend Başlatıcı
:: Bu script kap-app-backend klasöründen çalışmalıdır.
:: .env dosyası bu klasörde bulunmalıdır.

cd /d "%~dp0"

echo [%date% %time%] Backend baslatiliyor...

:LOOP
    api.exe
    echo [%date% %time%] Backend durdu! 3 saniye sonra yeniden baslatiliyor...
    timeout /t 3 /nobreak
goto LOOP
