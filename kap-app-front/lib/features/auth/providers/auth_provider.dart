// lib/features/auth/providers/auth_provider.dart

import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../../core/constants/app_constants.dart';
import '../../../core/network/api_client.dart';
import '../data/auth_repository.dart';

enum AuthStatus { initial, loading, authenticated, unauthenticated, error }

class AuthProvider extends ChangeNotifier {
  final AuthRepository _repo;
  final ApiClient _apiClient;

  AuthProvider(this._repo, this._apiClient, {String? initialUserId, String? initialTenantId}) {
    _initToken();
    if (initialUserId != null) {
      _userId = initialUserId;
      _status = AuthStatus.authenticated;
    }
    if (initialTenantId != null) {
      _currentTenantId = initialTenantId;
    }
  }

  Future<void> _initToken() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final savedToken = prefs.getString('auth_token') ?? prefs.getString(kTokenKey);
      if (savedToken != null) {
        _apiClient.updateToken(savedToken);
      }
    } catch (e) {
      debugPrint('AuthProvider._initToken error: $e');
    }
  }

  AuthStatus _status = AuthStatus.initial;
  String? _errorMessage;
  String? _userId;
  String? _currentTenantId;

  AuthStatus get status      => _status;
  String? get errorMessage   => _errorMessage;
  String? get userId         => _userId;
  String? get currentTenantId => _currentTenantId;
  bool get isAuthenticated   => _status == AuthStatus.authenticated;

  /// Kullanıcının aktif tenant'unu ayarlar ve SharedPreferences'a kaydeder.
  Future<void> setCurrentTenant(String tenantId) async {
    _currentTenantId = tenantId;
    notifyListeners();
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('current_tenant_id', tenantId);
  }

  /// Backend'den kullanıcının evlerini sorgular, eğer varsa ilkini aktif ev olarak ayarlar.
  Future<void> fetchAndSetDefaultTenant() async {
    try {
      final List<dynamic> tenants = await _repo.listTenants();
      if (tenants.isNotEmpty) {
        final first = tenants.first as Map<String, dynamic>;
        final id = first['id'] as String?;
        if (id != null) {
          await setCurrentTenant(id);
        }
      }
    } catch (e) {
      debugPrint('AuthProvider.fetchAndSetDefaultTenant error: $e');
    }
  }

  Future<void> login(String email, String password) async {
    _status = AuthStatus.loading;
    _errorMessage = null;
    notifyListeners();

    try {
      final res = await _repo.login(email: email, password: password);
      final token = res['access_token'] as String?;
      _userId = res['user_id'] as String?;

      if (token != null && _userId != null) {
        _repo.setToken(token);
        _apiClient.updateToken(token);

        final prefs = await SharedPreferences.getInstance();
        await prefs.setString(kTokenKey, token);
        await prefs.setString('auth_token', token);
        await prefs.setString('user_id', _userId!);

        _status = AuthStatus.authenticated;
      } else {
        throw const ApiException(400, '{"error": "Giriş cevabında oturum anahtarı bulunamadı."}');
      }
    } catch (e) {
      _errorMessage = _cleanErrorMessage(e);
      _status = AuthStatus.error;
    }
    notifyListeners();
  }

  Future<void> register(String email, String password, String name) async {
    _status = AuthStatus.loading;
    _errorMessage = null;
    notifyListeners();

    try {
      final res = await _repo.register(email: email, password: password, name: name);
      final token = res['access_token'] as String?;
      _userId = res['user_id'] as String?;

      if (token != null && _userId != null) {
        _repo.setToken(token);
        _apiClient.updateToken(token);

        final prefs = await SharedPreferences.getInstance();
        await prefs.setString(kTokenKey, token);
        await prefs.setString('auth_token', token);
        await prefs.setString('user_id', _userId!);

        _status = AuthStatus.authenticated;
      } else if (_userId != null) {
        // Token dönmediyse ama kayıt başarılıysa (örn. e-posta onayı gerekiyorsa)
        _status = AuthStatus.unauthenticated;
        _errorMessage = "Kayıt başarılı. Lütfen e-posta adresinizi onaylayın.";
      } else {
        throw const ApiException(400, '{"error": "Kayıt cevabında kullanıcı bilgiisi bulunamadı."}');
      }
    } catch (e) {
      _errorMessage = _cleanErrorMessage(e);
      _status = AuthStatus.error;
    }
    notifyListeners();
  }

  Future<void> logout() async {
    try {
      await _repo.logout();
    } catch (_) {
      // API isteği başarısız olsa bile yerel oturumu temizlemeye devam et
    } finally {
      _apiClient.updateToken(null);
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(kTokenKey);
      await prefs.remove('auth_token');
      await prefs.remove('user_id');

      _status = AuthStatus.unauthenticated;
      _userId = null;
      notifyListeners();
    }
  }

  String _cleanErrorMessage(dynamic e) {
    if (e is ApiException) {
      try {
        final decoded = json.decode(e.message);
        if (decoded is Map && decoded.containsKey('error')) {
          return decoded['error'].toString();
        }
      } catch (_) {
        // JSON değilse ham metni kullan
      }
      return e.message;
    }
    final errStr = e.toString();
    if (errStr.contains('Connection refused') || errStr.contains('SocketException')) {
      return "Sunucuya bağlanılamadı. Lütfen sunucunun çalıştığından emin olun.";
    }
    return errStr;
  }
}
