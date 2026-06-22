// lib/core/theme/app_theme.dart
//
// Netflix Dark tema paleti.
// Primary: #E50914 (Netflix kırmızısı)
// Background: #000000 / #141414 (koyu kömür)

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class AppTheme {
  AppTheme._();

  // ── Netflix Marka Renkleri ────────────────────────────────────────────────
  static const Color netflixRed       = Color(0xFFE50914); // birincil / vurgu
  static const Color netflixRedDark   = Color(0xFFB20710); // basılı durum
  static const Color netflixRedGlow   = Color(0xFFE50914); // glow için

  // ── Arka Plan ─────────────────────────────────────────────────────────────
  static const Color bgBlack          = Color(0xFF000000); // tam siyah scaffold
  static const Color bgCard           = Color(0xFF141414); // kart / alt navbar
  static const Color bgElevated       = Color(0xFF1F1F1F); // input, chip arka plan
  static const Color bgSheet          = Color(0xFF1A1A1A); // bottom sheet

  // ── Kenarlık / Ayırıcı ────────────────────────────────────────────────────
  static const Color borderDefault    = Color(0xFF2A2A2A);
  static const Color borderFocused    = netflixRed;

  // ── Metin ─────────────────────────────────────────────────────────────────
  static const Color textPrimary      = Color(0xFFFFFFFF); // başlık / ana
  static const Color textSecondary    = Color(0xFFAAAAAA); // alt başlık / ipucu
  static const Color textDisabled     = Color(0xFF555555); // pasif

  // ── Durum ─────────────────────────────────────────────────────────────────
  static const Color success          = Color(0xFF2ECC71);
  static const Color error            = Color(0xFFE74C3C);
  static const Color warning          = Color(0xFFF39C12);

  // ── Tema Üretici ─────────────────────────────────────────────────────────

  static ThemeData darkTheme() {
    return ThemeData(
      useMaterial3: true,

      // ColorScheme
      colorScheme: const ColorScheme.dark(
        primary:       netflixRed,
        onPrimary:     Colors.white,
        secondary:     success,
        onSecondary:   Colors.white,
        error:         error,
        surface:       bgCard,
        onSurface:     textPrimary,
      ),

      scaffoldBackgroundColor: bgBlack,
      cardColor:               bgCard,

      // AppBar
      appBarTheme: const AppBarTheme(
        backgroundColor:  bgBlack,
        foregroundColor:  textPrimary,
        elevation:        0,
        centerTitle:      true,
        systemOverlayStyle: SystemUiOverlayStyle(
          statusBarColor:           Colors.transparent,
          statusBarIconBrightness:  Brightness.light,
        ),
      ),

      // ElevatedButton → Kırmızı
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor:  netflixRed,
          foregroundColor:  Colors.white,
          elevation:        0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
          textStyle: const TextStyle(
            fontWeight: FontWeight.w700,
            fontSize:   15,
          ),
        ),
      ),

      // TextButton
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(foregroundColor: netflixRed),
      ),

      // Input
      inputDecorationTheme: InputDecorationTheme(
        filled:       true,
        fillColor:    bgElevated,
        hintStyle:    const TextStyle(color: textSecondary, fontSize: 14),
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: borderDefault),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: borderDefault),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: netflixRed, width: 1.5),
        ),
      ),

      // Text
      textTheme: const TextTheme(
        titleLarge:  TextStyle(color: textPrimary,   fontWeight: FontWeight.w700),
        titleMedium: TextStyle(color: textPrimary,   fontWeight: FontWeight.w600),
        bodyLarge:   TextStyle(color: textPrimary),
        bodyMedium:  TextStyle(color: textSecondary),
        labelSmall:  TextStyle(color: textSecondary),
      ),

      // Divider
      dividerTheme: const DividerThemeData(color: borderDefault, thickness: 1),

      // BottomSheet
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor:    bgSheet,
        modalBackgroundColor: bgSheet,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
      ),

      fontFamily: 'Roboto',
    );
  }

  // Light tema (ileride kullanılabilir; şimdilik dark öncelikli)
  static ThemeData lightTheme() {
    return ThemeData(
      useMaterial3: true,
      colorScheme: const ColorScheme.light(primary: netflixRed),
      fontFamily: 'Roboto',
    );
  }
}
