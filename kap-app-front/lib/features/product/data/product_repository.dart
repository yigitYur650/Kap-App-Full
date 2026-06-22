// lib/features/product/data/product_repository.dart

import '../../../core/network/api_client.dart';
import '../domain/product_model.dart';

class ProductRepository {
  final ApiClient _client;
  ProductRepository(this._client);

  Future<List<Product>> getItems(String tenantId) async {
    final raw = await _client.getList('/products?tenant_id=$tenantId');
    return raw
        .map((e) => Product.fromMap(e as Map<String, dynamic>))
        .toList();
  }

  Future<Product> addProduct({
    required String tenantId,
    required String name,
    required int quantity,
    double? price,
    String? marketName,
    String? category,
    String? unit,
    String? expirationDate,
  }) async {
    final res = await _client.post('/products', {
      'tenant_id': tenantId,
      'name': name,
      'quantity': quantity,
      'status': 'yok',
      'is_completed': false,
      'bought': false,
      if (price != null) 'price': price,
      if (marketName != null) 'market_name': marketName,
      if (category != null) 'category': category,
      if (unit != null && unit.isNotEmpty) 'unit': unit,
      if (expirationDate != null) 'expiration_date': expirationDate,
    });
    return Product.fromMap(res);
  }

  Future<Product> updateStatus({
    required String productId,
    required String status,
  }) async {
    final res = await _client.patch('/products/$productId/status', {'status': status});
    return Product.fromMap(res);
  }

  Future<void> deleteProduct(String productId) async {
    await _client.delete('/products/$productId');
  }
}
