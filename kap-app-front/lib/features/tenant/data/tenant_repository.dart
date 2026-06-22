// lib/features/tenant/data/tenant_repository.dart

import '../../../core/network/api_client.dart';

class TenantRepository {
  final ApiClient _client;
  TenantRepository(this._client);

  Future<Map<String, dynamic>> createHome(String name) async {
    return _client.post('/tenants', {'name': name});
  }

  Future<Map<String, dynamic>> joinHome(String inviteCode) async {
    return _client.post('/tenants/join', {'invite_code': inviteCode});
  }

  Future<Map<String, dynamic>> getHome(String tenantId) async {
    return _client.get('/tenants/$tenantId');
  }

  Future<List<dynamic>> getMembers(String tenantId) async {
    final res = await _client.get('/tenants/$tenantId/members');
    return res['members'] as List<dynamic>? ?? [];
  }
}
