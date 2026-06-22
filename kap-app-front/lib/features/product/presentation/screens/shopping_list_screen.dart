// lib/features/product/presentation/screens/shopping_list_screen.dart
//
// Netflix Dark — Gelişmiş UX ile Alışveriş Listesi
//
// UX İyileştirmeleri:
//   • Ürün kartlarında 3-nokta popup menü → Düzenle / Sil
//   • Checkbox tıklanınca → 320ms yumuşak çıkış animasyonu (scale+fade+slide)
//   • Silme işlemi de aynı animasyonu kullanır
//   • Düzenleme → Netflix temalı EditProductSheet bottom sheet
//   • AnimatedOpacity + AnimatedSlide + AnimatedScale (ImplicitAnimation)

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shimmer/shimmer.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/localization/app_localizations.dart';
import '../../domain/product_model.dart';
import '../../providers/product_provider.dart';
import '../../../auth/providers/auth_provider.dart';

// ─── Ana Ekran ───────────────────────────────────────────────────────────────

class ShoppingListScreen extends StatefulWidget {
  const ShoppingListScreen({super.key});

  @override
  State<ShoppingListScreen> createState() => _ShoppingListScreenState();
}

class _ShoppingListScreenState extends State<ShoppingListScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);

    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final provider  = context.read<ProductProvider>();
      final auth      = context.read<AuthProvider>();
      
      // Eğer henüz ev seçilmemişse backend'den ev listesini çekip ilkini seçelim
      if (auth.currentTenantId == null) {
        await auth.fetchAndSetDefaultTenant();
      }

      final tenantId  = auth.currentTenantId;
      if (tenantId != null && provider.status == ProductStatus.initial) {
        provider.loadItems(tenantId);
      }
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final tenantId = auth.currentTenantId;

    if (tenantId == null) {
      final l = AppLocalizations.of(context);
      return Column(
        children: [
          _ShoppingAppBar(tabController: _tabController),
          Expanded(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: AppTheme.bgCard,
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(color: AppTheme.netflixRed.withValues(alpha: 0.3), width: 1.5),
                      ),
                      child: const Icon(
                        Icons.home_outlined,
                        color: AppTheme.netflixRed,
                        size: 48,
                      ),
                    ),
                    const SizedBox(height: 24),
                    Text(
                      l.t('tenant.no_home_selected'),
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: AppTheme.textPrimary,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      l.t('tenant.no_home_desc'),
                      style: const TextStyle(
                        fontSize: 14,
                        color: AppTheme.textSecondary,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      );
    }

    return Column(
      children: [
        _ShoppingAppBar(tabController: _tabController),
        Expanded(
          child: Consumer<ProductProvider>(
            builder: (context, provider, _) {
              if (provider.status == ProductStatus.loading ||
                  provider.status == ProductStatus.initial) {
                return _SkeletonPage();
              }
              if (provider.status == ProductStatus.error) {
                return _ErrorView(message: provider.errorMessage ?? 'Hata');
              }
              return TabBarView(
                controller: _tabController,
                physics: const NeverScrollableScrollPhysics(),
                children: [
                  _NeedsTab(provider: provider),
                  _GotTab(provider: provider),
                ],
              );
            },
          ),
        ),
      ],
    );
  }
}

// ─── Üst Bar + Tab ────────────────────────────────────────────────────────────

class _ShoppingAppBar extends StatelessWidget {
  final TabController tabController;
  const _ShoppingAppBar({required this.tabController});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: AppTheme.bgBlack,
      child: SafeArea(
        bottom: false,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 16, 20, 10),
              child: Consumer<ProductProvider>(
                builder: (context, provider, _) {
                  final total = provider.needs.length + provider.got.length;
                  final l = AppLocalizations.of(context);
                  return Row(
                    children: [
                      Text(
                        l.t('product.shopping_list'),
                        style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.w800,
                          color: AppTheme.textPrimary,
                        ),
                      ),
                      const SizedBox(width: 10),
                      if (total > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(
                            color: AppTheme.netflixRed,
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: Text(
                            '$total',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                    ],
                  );
                },
              ),
            ),
            TabBar(
              controller: tabController,
              indicatorColor: AppTheme.netflixRed,
              indicatorWeight: 2.5,
              indicatorSize: TabBarIndicatorSize.label,
              labelColor: AppTheme.netflixRed,
              unselectedLabelColor: AppTheme.textDisabled,
              labelStyle: const TextStyle(
                fontWeight: FontWeight.w700,
                fontSize: 14,
              ),
              unselectedLabelStyle: const TextStyle(
                fontWeight: FontWeight.w400,
                fontSize: 14,
              ),
              dividerColor: AppTheme.borderDefault,
              tabs: [
                _CountedTab(label: AppLocalizations.of(context).t('add_product.needs'), getter: (p) => p.needs.length),
                _CountedTab(label: AppLocalizations.of(context).t('add_product.got'),   getter: (p) => p.got.length),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _CountedTab extends StatelessWidget {
  final String label;
  final int Function(ProductProvider) getter;
  const _CountedTab({required this.label, required this.getter});

  @override
  Widget build(BuildContext context) {
    return Consumer<ProductProvider>(
      builder: (context, p, _) {
        final count = getter(p);
        return Tab(
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(label),
              if (count > 0) ...[
                const SizedBox(width: 6),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                  decoration: BoxDecoration(
                    color: AppTheme.bgElevated,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text('$count',
                      style: const TextStyle(fontSize: 11)),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

// ─── Sekmeler ────────────────────────────────────────────────────────────────

class _NeedsTab extends StatelessWidget {
  final ProductProvider provider;
  const _NeedsTab({required this.provider});

  @override
  Widget build(BuildContext context) {
    if (provider.needs.isEmpty) {
      final l = AppLocalizations.of(context);
      return _EmptyView(
        icon: Icons.celebration_outlined,
        message: l.t('add_product.empty_needs'),
      );
    }
    return _AnimatedItemList(items: provider.needs, provider: provider);
  }
}

class _GotTab extends StatelessWidget {
  final ProductProvider provider;
  const _GotTab({required this.provider});

  @override
  Widget build(BuildContext context) {
    if (provider.got.isEmpty) {
      final l = AppLocalizations.of(context);
      return _EmptyView(
        icon: Icons.shopping_cart_outlined,
        message: l.t('add_product.empty_got'),
      );
    }
    return _AnimatedItemList(items: provider.got, provider: provider);
  }
}

// ─── Animasyonlu Liste ────────────────────────────────────────────────────────
// Her kart kendi AnimatedProductCard'ı ile render edilir.
// pendingToggles içindeyse kart scale+fade out animasyonu oynar.

class _AnimatedItemList extends StatelessWidget {
  final List<Product> items;
  final ProductProvider provider;
  const _AnimatedItemList({required this.items, required this.provider});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
      itemCount: items.length,
      itemBuilder: (_, i) => AnimatedProductCard(
        key: ValueKey(items[i].id),
        product: items[i],
        provider: provider,
      ),
    );
  }
}

// ─── Animasyonlu Ürün Kartı ───────────────────────────────────────────────────

class AnimatedProductCard extends StatefulWidget {
  final Product product;
  final ProductProvider provider;

  const AnimatedProductCard({
    super.key,
    required this.product,
    required this.provider,
  });

  @override
  State<AnimatedProductCard> createState() => _AnimatedProductCardState();
}

class _AnimatedProductCardState extends State<AnimatedProductCard>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _opacity;
  late final Animation<double> _scale;
  late final Animation<Offset> _slide;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
      value: 1.0,
    );
    _opacity = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _ctrl, curve: Curves.easeOut),
    );
    _scale = Tween<double>(begin: 0.90, end: 1.0).animate(
      CurvedAnimation(parent: _ctrl, curve: Curves.easeOut),
    );
    _slide = Tween<Offset>(
      begin: const Offset(0.15, 0),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _ctrl, curve: Curves.easeOut));
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  /// Kartın pending durumuna göre animasyonu tetikle
  void _syncAnimation(bool isPending) {
    if (isPending) {
      _ctrl.reverse();
    } else if (!_ctrl.isCompleted) {
      _ctrl.forward();
    }
  }

  void _openEditSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => _EditProductSheet(
        product: widget.product,
        provider: widget.provider,
      ),
    );
  }

  void _confirmDelete() {
    final l = AppLocalizations.of(context);
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppTheme.bgCard,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
        title: Text(
          l.t('product.delete_item'),
          style: const TextStyle(color: AppTheme.textPrimary, fontWeight: FontWeight.w700),
        ),
        content: Text(
          '"${widget.product.name}" listeden kaldırılsın mı?',
          style: const TextStyle(color: AppTheme.textSecondary),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(
              l.t('common.cancel'),
              style: const TextStyle(color: AppTheme.textSecondary),
            ),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(ctx);
              widget.provider.deleteItem(widget.product.id);
            },
            child: Text(
              l.t('common.delete'),
              style: TextStyle(
                color: AppTheme.netflixRed,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Selector<ProductProvider, bool>(
      selector: (_, p) => p.pendingToggles.contains(widget.product.id),
      builder: (context, isPending, child) {
        _syncAnimation(isPending);
        return FadeTransition(
          opacity: _opacity,
          child: ScaleTransition(
            scale: _scale,
            child: SlideTransition(
              position: _slide,
              child: _ProductTile(
                product: widget.product,
                provider: widget.provider,
                onEdit: _openEditSheet,
                onDelete: _confirmDelete,
              ),
            ),
          ),
        );
      },
    );
  }
}

// ─── Ürün Tile (saf görsel, stateless) ───────────────────────────────────────

class _ProductTile extends StatelessWidget {
  final Product product;
  final ProductProvider provider;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _ProductTile({
    required this.product,
    required this.provider,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final isChecked = product.checked;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: AppTheme.borderDefault, width: 0.8),
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(10),
          onTap: () => provider.toggleItem(product.id),
          splashColor: AppTheme.netflixRed.withValues(alpha: 0.08),
          highlightColor: AppTheme.netflixRed.withValues(alpha: 0.04),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(14, 11, 8, 11),
            child: Row(
              children: [
                // ── Checkbox ─────────────────────────────────────────────
                _NxCheckbox(
                  checked: isChecked,
                  onTap: () => provider.toggleItem(product.id),
                ),
                const SizedBox(width: 12),

                // ── İsim + miktar ──────────────────────────────────────────
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      AnimatedDefaultTextStyle(
                        duration: const Duration(milliseconds: 200),
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: isChecked
                              ? AppTheme.textDisabled
                              : AppTheme.textPrimary,
                          decoration: isChecked
                              ? TextDecoration.lineThrough
                              : TextDecoration.none,
                          decorationColor: AppTheme.textDisabled,
                        ),
                        child: Text(product.name),
                      ),
                      if (product.quantity > 0 || product.marketName != null) ...[
                        const SizedBox(height: 2),
                        Text(
                          [
                            if (product.unit != null && product.unit!.isNotEmpty)
                              '${product.quantity} ${product.unit}'
                            else if (product.quantity > 1)
                              '${product.quantity} adet',
                            if (product.marketName != null) product.marketName!,
                          ].join(' · '),
                          style: const TextStyle(
                            fontSize: 12,
                            color: AppTheme.textSecondary,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),

                // ── Kategori Badge ────────────────────────────────────────
                if (product.category != null) ...[
                  _CategoryBadge(category: product.category!),
                  const SizedBox(width: 4),
                ],

                // ── 3-Nokta Menüsü ────────────────────────────────────────
                _ActionMenu(onEdit: onEdit, onDelete: onDelete),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ─── 3-Nokta Aksiyon Menüsü ──────────────────────────────────────────────────

class _ActionMenu extends StatelessWidget {
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  const _ActionMenu({required this.onEdit, required this.onDelete});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<_Action>(
      onSelected: (action) {
        if (action == _Action.edit)   onEdit();
        if (action == _Action.delete) onDelete();
      },
      color: AppTheme.bgElevated,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: const BorderSide(color: AppTheme.borderDefault, width: 0.5),
      ),
      elevation: 8,
      offset: const Offset(-8, 8),
      icon: const Icon(
        Icons.more_vert_rounded,
        color: AppTheme.textDisabled,
        size: 20,
      ),
      itemBuilder: (_) => [
        PopupMenuItem(
          value: _Action.edit,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          child: Row(
            children: [
              const Icon(Icons.edit_outlined, size: 16, color: AppTheme.textSecondary),
              const SizedBox(width: 10),
              Text(
                AppLocalizations.of(context).t('common.edit'),
                style: const TextStyle(
                  color: AppTheme.textPrimary,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
        const PopupMenuDivider(height: 1),
        PopupMenuItem(
          value: _Action.delete,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          child: Row(
            children: [
              const Icon(Icons.delete_outline_rounded, size: 16, color: AppTheme.netflixRed),
              const SizedBox(width: 10),
              Text(
                AppLocalizations.of(context).t('common.delete'),
                style: const TextStyle(
                  color: AppTheme.netflixRed,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

enum _Action { edit, delete }

// ─── Netflix Checkbox ─────────────────────────────────────────────────────────

class _NxCheckbox extends StatelessWidget {
  final bool checked;
  final VoidCallback onTap;
  const _NxCheckbox({required this.checked, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        width: 22,
        height: 22,
        decoration: BoxDecoration(
          color: checked ? AppTheme.netflixRed : Colors.transparent,
          border: Border.all(
            color: checked ? AppTheme.netflixRed : AppTheme.borderDefault,
            width: 1.8,
          ),
          borderRadius: BorderRadius.circular(5),
        ),
        child: checked
            ? const Icon(Icons.check_rounded, size: 14, color: Colors.white)
            : null,
      ),
    );
  }
}

// ─── Kategori Badge ───────────────────────────────────────────────────────────

class _CategoryBadge extends StatelessWidget {
  final String category;
  const _CategoryBadge({required this.category});

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final localizedLabels = {
      'produce':   ('🥦', l.t('product.categories.produce')),
      'dairy':     ('🥛', l.t('product.categories.dairy')),
      'meat':      ('🥩', l.t('product.categories.meat')),
      'bakery':    ('🍞', l.t('product.categories.bakery')),
      'beverages': ('🧃', l.t('product.categories.beverages')),
      'cleaning':  ('🧹', l.t('product.categories.cleaning')),
      'personal':  ('🧴', l.t('product.categories.personal')),
      'other':     ('📦', l.t('product.categories.other')),
    };
    final info = localizedLabels[category] ?? ('📦', category);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
      decoration: BoxDecoration(
        color: AppTheme.bgElevated,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: AppTheme.borderDefault, width: 0.5),
      ),
      child: Text(
        '${info.$1} ${info.$2}',
        style: const TextStyle(fontSize: 11, color: AppTheme.textSecondary),
      ),
    );
  }
}

// ─── Boş Durum ────────────────────────────────────────────────────────────────

class _EmptyView extends StatelessWidget {
  final IconData icon;
  final String message;
  const _EmptyView({required this.icon, required this.message});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 52, color: AppTheme.textDisabled),
          const SizedBox(height: 16),
          Text(
            message,
            style: const TextStyle(color: AppTheme.textSecondary, fontSize: 14),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

// ─── Hata Ekranı ─────────────────────────────────────────────────────────────

class _ErrorView extends StatelessWidget {
  final String message;
  const _ErrorView({required this.message});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.error_outline_rounded,
              color: AppTheme.error, size: 48),
          const SizedBox(height: 12),
          Text(message,
              textAlign: TextAlign.center,
              style: const TextStyle(color: AppTheme.textSecondary)),
        ],
      ),
    );
  }
}

// ─── Netflix Shimmer Skeleton ─────────────────────────────────────────────────

class _SkeletonPage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Shimmer.fromColors(
      baseColor: AppTheme.bgCard,
      highlightColor: AppTheme.bgElevated,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        itemCount: 8,
        itemBuilder: (_, i) => _SkeletonCard(narrow: i % 4 == 3),
      ),
    );
  }
}

class _SkeletonCard extends StatelessWidget {
  final bool narrow;
  const _SkeletonCard({required this.narrow});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: [
          Container(
            width: 22,
            height: 22,
            decoration: BoxDecoration(
              color: AppTheme.bgElevated,
              borderRadius: BorderRadius.circular(5),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  height: 13,
                  width: narrow ? 110 : double.infinity,
                  decoration: BoxDecoration(
                    color: AppTheme.bgElevated,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(height: 7),
                Container(
                  height: 10,
                  width: 56,
                  decoration: BoxDecoration(
                    color: AppTheme.bgElevated,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Container(
            width: 20, height: 20,
            decoration: BoxDecoration(
              color: AppTheme.bgElevated,
              borderRadius: BorderRadius.circular(4),
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Ürün Düzenleme Bottom Sheet ─────────────────────────────────────────────

class _EditProductSheet extends StatefulWidget {
  final Product product;
  final ProductProvider provider;
  const _EditProductSheet({required this.product, required this.provider});

  @override
  State<_EditProductSheet> createState() => _EditProductSheetState();
}

class _EditProductSheetState extends State<_EditProductSheet> {
  final _formKey  = GlobalKey<FormState>();
  late final TextEditingController _nameCtrl;
  late final TextEditingController _qtyCtrl;
  late final TextEditingController _unitCtrl;
  String? _selectedCategory;
  bool _isSaving = false;

  List<(String, String)> get _localizedCategories {
    final l = AppLocalizations.of(context);
    return [
      ('produce',   '🥦 ${l.t('product.categories.produce')}'),
      ('dairy',     '🥛 ${l.t('product.categories.dairy')}'),
      ('meat',      '🥩 ${l.t('product.categories.meat')}'),
      ('bakery',    '🍞 ${l.t('product.categories.bakery')}'),
      ('beverages', '🧃 ${l.t('product.categories.beverages')}'),
      ('cleaning',  '🧹 ${l.t('product.categories.cleaning')}'),
      ('personal',  '🧴 ${l.t('product.categories.personal')}'),
      ('other',     '📦 ${l.t('product.categories.other')}'),
    ];
  }

  @override
  void initState() {
    super.initState();
    final p = widget.product;
    _nameCtrl = TextEditingController(text: p.name);
    _qtyCtrl  = TextEditingController(text: '${p.quantity}');
    _unitCtrl = TextEditingController(text: p.unit ?? '');
    _selectedCategory = p.category;
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _qtyCtrl.dispose();
    _unitCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isSaving = true);
    // Backend şu an sadece status güncellemesini destekliyor.
    // Seçili kategori değişmişse bunu status ile birlikte iletilemez;
    // tam edit için backend'e PATCH /products/{id} eklenecek.
    await widget.provider.updateStatus(
      widget.product.id,
      _selectedCategory != null ? 'var' : widget.product.status,
    );
    if (mounted) Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      child: Container(
        decoration: const BoxDecoration(
          color: AppTheme.bgSheet,
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(20, 14, 20, 20),
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Handle
                  Center(
                    child: Container(
                      width: 36, height: 4,
                      decoration: BoxDecoration(
                        color: AppTheme.borderDefault,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Başlık
                  Row(
                    children: [
                      Container(
                        width: 3, height: 20,
                        decoration: BoxDecoration(
                          color: AppTheme.netflixRed,
                          borderRadius: BorderRadius.circular(2),
                        ),
                      ),
                      const SizedBox(width: 10),
                      Text(
                        l.t('common.edit'),
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w700,
                          color: AppTheme.textPrimary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 18),

                  // Ürün Adı
                  _SheetLabel(l.t('add_product.name_label').toUpperCase()),
                  const SizedBox(height: 6),
                  TextFormField(
                    controller: _nameCtrl,
                    autofocus: true,
                    style: const TextStyle(color: AppTheme.textPrimary),
                    decoration: InputDecoration(hintText: l.t('add_product.name_hint')),
                    validator: (v) =>
                        (v == null || v.trim().isEmpty) ? l.t('common.error') : null,
                  ),
                  const SizedBox(height: 14),

                  // Miktar + Birim
                  Row(
                    children: [
                      Expanded(
                        flex: 2,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            _SheetLabel(l.t('add_product.qty_label').toUpperCase()),
                            const SizedBox(height: 6),
                            TextFormField(
                              controller: _qtyCtrl,
                              keyboardType: TextInputType.number,
                              style: const TextStyle(color: AppTheme.textPrimary),
                              decoration: const InputDecoration(hintText: '1'),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        flex: 3,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            _SheetLabel(l.t('add_product.unit_label').toUpperCase()),
                            const SizedBox(height: 6),
                            TextFormField(
                              controller: _unitCtrl,
                              style: const TextStyle(color: AppTheme.textPrimary),
                              decoration: InputDecoration(hintText: l.t('add_product.unit_hint')),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),

                  // Kategori
                  _SheetLabel(l.t('add_product.category_label').toUpperCase()),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _localizedCategories.map((cat) {
                      final isSelected = _selectedCategory == cat.$1;
                      return GestureDetector(
                        onTap: () => setState(() =>
                            _selectedCategory = isSelected ? null : cat.$1),
                        child: AnimatedContainer(
                          duration: const Duration(milliseconds: 160),
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 6),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? AppTheme.netflixRed
                                : AppTheme.bgElevated,
                            border: Border.all(
                              color: isSelected
                                  ? AppTheme.netflixRed
                                  : AppTheme.borderDefault,
                            ),
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            cat.$2,
                            style: TextStyle(
                              fontSize: 12,
                              color: isSelected
                                  ? Colors.white
                                  : AppTheme.textSecondary,
                              fontWeight: isSelected
                                  ? FontWeight.w600
                                  : FontWeight.w400,
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                  const SizedBox(height: 22),

                  // Kaydet
                  SizedBox(
                    width: double.infinity,
                    height: 50,
                    child: _isSaving
                        ? const Center(
                            child: CircularProgressIndicator(
                              color: AppTheme.netflixRed,
                              strokeWidth: 2.5,
                            ),
                          )
                        : ElevatedButton(
                            onPressed: _save,
                            child: Text(l.t('common.save')),
                          ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _SheetLabel extends StatelessWidget {
  final String text;
  const _SheetLabel(this.text);

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      style: const TextStyle(
        fontSize: 10,
        fontWeight: FontWeight.w700,
        color: AppTheme.textSecondary,
        letterSpacing: 1.2,
      ),
    );
  }
}
