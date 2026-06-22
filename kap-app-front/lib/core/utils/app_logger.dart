// lib/core/utils/app_logger.dart

import 'package:flutter/foundation.dart';

class AppLogger {
  static void d(String tag, String message) {
    if (kDebugMode) debugPrint('[$tag] $message');
  }

  static void e(String tag, String message, [Object? error]) {
    if (kDebugMode) debugPrint('[ERROR][$tag] $message ${error ?? ''}');
  }
}
