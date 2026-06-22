// lib/features/auth/data/auth_repository.dart

import '../../../core/network/api_client.dart';

class AuthRepository {
  final ApiClient _client;
  AuthRepository(this._client);

  Future<Map<String, dynamic>> login({
    required String email,
    required String password,
  }) async {
    return _client.post('/auth/login', {
      'email': email,
      'password': password,
    });
  }

  Future<Map<String, dynamic>> register({
    required String email,
    required String password,
    required String name,
  }) async {
    return _client.post('/auth/register', {
      'email': email,
      'password': password,
      'name': name,
    });
  }

  Future<void> logout() async {
    await _client.post('/auth/logout', {});
    _client.clearToken();
  }

  Future<List<dynamic>> listTenants() async {
    return _client.getList('/tenants');
  }

  void setToken(String token) => _client.setToken(token);
  void clearToken() => _client.clearToken();
}
