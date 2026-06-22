// lib/features/product/providers/product_provider.dart

import 'package:flutter/foundation.dart';
import '../data/product_repository.dart';
import '../domain/product_model.dart';

enum ProductStatus { initial, loading, loaded, error }

class ProductProvider extends ChangeNotifier {
  final ProductRepository _repo;

  ProductProvider(this._repo);

  ProductStatus _status = ProductStatus.initial;
  String? _errorMessage;
  List<Product> _items = [];

  final Set<String> _pendingToggles = {};

  DateTime? _lastFetched;
  String? _cachedTenantId;
  static const _cacheDuration = Duration(minutes: 2);

  ProductStatus get status       => _status;
  String? get errorMessage       => _errorMessage;
  Set<String> get pendingToggles => _pendingToggles;

  List<Product> get needs => _items.where((p) => p.status == 'yok').toList();
  List<Product> get got => _items.where((p) => p.checked).toList();

  Future<void> loadItems(String tenantId, {bool forceRefresh = false}) async {
    if (!forceRefresh &&
        _cachedTenantId == tenantId &&
        _lastFetched != null &&
        DateTime.now().difference(_lastFetched!) < _cacheDuration &&
        _status == ProductStatus.loaded) {
      return;
    }

    _status = ProductStatus.loading;
    _errorMessage = null;
    notifyListeners();

    try {
      _items = await _repo.getItems(tenantId);
      _status = ProductStatus.loaded;
      _lastFetched = DateTime.now();
      _cachedTenantId = tenantId;
    } catch (e) {
      _errorMessage = e.toString();
      _status = ProductStatus.error;
    }
    notifyListeners();
  }

  void invalidateCache() {
    _lastFetched = null;
    _cachedTenantId = null;
    _items = [];
    _status = ProductStatus.initial;
  }

  Future<void> toggleItem(String id) async {
    if (_pendingToggles.contains(id)) return;

    final idx = _items.indexWhere((p) => p.id == id);
    if (idx == -1) return;

    _pendingToggles.add(id);
    notifyListeners();

    await Future.delayed(const Duration(milliseconds: 320));

    final current = _items[idx];
    final newStatus = current.status == 'yok' ? 'var' : 'yok';

    _items[idx] = current.copyWith(status: newStatus);
    _pendingToggles.remove(id);
    notifyListeners();

    try {
      final updated = await _repo.updateStatus(
        productId: id,
        status: newStatus,
      );
      _items[idx] = updated;
      
      _lastFetched = null;
      _cachedTenantId = null;
      
      notifyListeners();
    } catch (e) {
      _items[idx] = current;
      _errorMessage = e.toString();
      notifyListeners();
    }
  }

  Future<void> addItem({
    required String tenantId,
    required String name,
    required int quantity,
    double? price,
    String? marketName,
    String? category,
    String? unit,
    String? expirationDate,
  }) async {
    try {
      final product = await _repo.addProduct(
        tenantId: tenantId,
        name: name,
        quantity: quantity,
        price: price,
        marketName: marketName,
        category: category,
        unit: unit,
        expirationDate: expirationDate,
      );
      final newProduct = product.copyWith(status: 'yok');
      _items.insert(0, newProduct);
      
      _lastFetched = null;
      _cachedTenantId = null;
      
      notifyListeners();
    } catch (e) {
      _errorMessage = e.toString();
      notifyListeners();
    }
  }

  Future<void> deleteItem(String id) async {
    final idx = _items.indexWhere((p) => p.id == id);
    if (idx == -1) return;

    final backup = _items[idx];

    _pendingToggles.add(id);
    notifyListeners();

    await Future.delayed(const Duration(milliseconds: 280));

    _items.removeWhere((p) => p.id == id);
    _pendingToggles.remove(id);
    notifyListeners();

    try {
      await _repo.deleteProduct(id);
      
      _lastFetched = null;
      _cachedTenantId = null;
    } catch (e) {
      _items.insert(idx, backup);
      _errorMessage = e.toString();
      notifyListeners();
    }
  }

  Future<void> updateStatus(String id, String status) async {
    final idx = _items.indexWhere((p) => p.id == id);
    if (idx == -1) return;

    final backup = _items[idx];
    _items[idx] = backup.copyWith(status: status);
    notifyListeners();

    try {
      final updated = await _repo.updateStatus(productId: id, status: status);
      _items[idx] = updated;
      
      _lastFetched = null;
      _cachedTenantId = null;
      
      notifyListeners();
    } catch (e) {
      _items[idx] = backup;
      _errorMessage = e.toString();
      notifyListeners();
    }
  }

  Future<void> updateItem(String id, String status) => updateStatus(id, status);
}
