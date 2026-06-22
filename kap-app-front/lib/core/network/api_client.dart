// lib/core/network/api_client.dart
//
// HTTP istemci sarmalayıcısı. Supabase REST + custom Go API destekler.

import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../constants/app_constants.dart';

class ApiException implements Exception {
  final int statusCode;
  final String message;
  const ApiException(this.statusCode, this.message);

  @override
  String toString() => 'ApiException($statusCode): $message';
}

class ApiClient {
  final String baseUrl;
  String? _token;

  ApiClient({this.baseUrl = kApiBaseUrl});

  void updateToken(String? token) {
    _token = token;
  }

  void setToken(String token) => updateToken(token);
  void clearToken() => updateToken(null);

  Map<String, String> _buildHeaders() {
    final headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    };
    if (_token != null) {
      headers['Authorization'] = 'Bearer $_token';
    }
    return headers;
  }

  Future<Map<String, dynamic>> get(String path) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.get(uri, headers: _buildHeaders());
    return _handle(response);
  }

  /// GET endpoint'i JSON array dönüyorsa bu metodu kullan.
  Future<List<dynamic>> getList(String path) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.get(uri, headers: _buildHeaders());
    _logStatus(response);
    if (response.statusCode >= 200 && response.statusCode < 300) {
      if (response.body.isEmpty) return [];
      final decoded = json.decode(response.body);
      if (decoded is List) return decoded;
      // Tek Map dönüyorsa items veya data anahtarını dene
      if (decoded is Map) {
        return decoded['items'] as List<dynamic>? ??
            decoded['data'] as List<dynamic>? ?? [];
      }
      return [];
    }
    throw ApiException(response.statusCode, response.body);
  }

  Future<Map<String, dynamic>> post(
      String path, Map<String, dynamic> body) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.post(
      uri,
      headers: _buildHeaders(),
      body: json.encode(body),
    );
    return _handle(response);
  }

  Future<Map<String, dynamic>> put(
      String path, Map<String, dynamic> body) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.put(
      uri,
      headers: _buildHeaders(),
      body: json.encode(body),
    );
    return _handle(response);
  }

  Future<Map<String, dynamic>> patch(
      String path, Map<String, dynamic> body) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.patch(
      uri,
      headers: _buildHeaders(),
      body: json.encode(body),
    );
    return _handle(response);
  }

  Future<void> delete(String path) async {
    final uri = Uri.parse('$baseUrl$path');
    final response = await http.delete(uri, headers: _buildHeaders());
    _logStatus(response);
    if (response.statusCode >= 200 && response.statusCode < 300) return;
    throw ApiException(response.statusCode, response.body);
  }

  Map<String, dynamic> _handle(http.Response response) {
    _logStatus(response);
    if (response.statusCode >= 200 && response.statusCode < 300) {
      if (response.body.isEmpty) return {};
      return json.decode(response.body) as Map<String, dynamic>;
    }
    throw ApiException(response.statusCode, response.body);
  }

  void _logStatus(http.Response response) {
    if (kDebugMode) {
      debugPrint('[ApiClient] ${response.statusCode} ${response.request?.url}');
    }
  }
}

