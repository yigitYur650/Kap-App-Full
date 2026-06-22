// lib/features/product/domain/product_model.dart

class Product {
  final String id;
  final String tenantId;
  final String name;
  final int quantity;
  final double? price;
  final String? marketName;
  final String? category;
  final String? unit;
  final String? addedBy;
  // status: 'var' | 'azaldı' | 'yok'
  final String status;

  const Product({
    required this.id,
    required this.tenantId,
    required this.name,
    required this.quantity,
    required this.status,
    this.price,
    this.marketName,
    this.category,
    this.unit,
    this.addedBy,
  });

  /// UI kolaylığı: 'var' veya 'azaldı' → alındı (checked)
  bool get checked => status == 'var' || status == 'azaldı';

  Product copyWith({
    String? id,
    String? tenantId,
    String? name,
    int? quantity,
    double? price,
    String? marketName,
    String? category,
    String? unit,
    String? addedBy,
    String? status,
  }) {
    return Product(
      id: id ?? this.id,
      tenantId: tenantId ?? this.tenantId,
      name: name ?? this.name,
      quantity: quantity ?? this.quantity,
      price: price ?? this.price,
      marketName: marketName ?? this.marketName,
      category: category ?? this.category,
      unit: unit ?? this.unit,
      addedBy: addedBy ?? this.addedBy,
      status: status ?? this.status,
    );
  }

  /// Backend Go JSON'undan oluşturur.
  factory Product.fromMap(Map<String, dynamic> map) {
    return Product(
      id: map['id'] as String,
      tenantId: map['tenant_id'] as String? ?? '',
      name: map['name'] as String,
      quantity: (map['quantity'] as num?)?.toInt() ?? 1,
      price: (map['price'] as num?)?.toDouble(),
      marketName: map['market_name'] as String?,
      category: map['category'] as String?,
      unit: map['unit'] as String?,
      addedBy: map['added_by'] as String?,
      status: map['status'] as String? ?? 'yok',
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'tenant_id': tenantId,
      'name': name,
      'quantity': quantity,
      if (price != null) 'price': price,
      if (marketName != null) 'market_name': marketName,
      if (category != null) 'category': category,
      if (unit != null) 'unit': unit,
      'status': status,
    };
  }
}
