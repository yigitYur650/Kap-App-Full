// lib/features/home/presentation/screens/main_layout.dart
//
// Ana arayüz iskeleti — Netflix Dark teması.
//   - PageView swipe navigasyon (Evim ↔ Liste ↔ Ayarlar)
//   - Bottom Navbar: siyah arka plan, kırmızı aktif gösterge
//   - Persistent FAB: Netflix kırmızısı "+" butonu
//   - FAB → BottomSheet: Ürün Ekle formu

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../../product/providers/product_provider.dart';
import 'hub_screen.dart';
import '../../../product/presentation/screens/shopping_list_screen.dart';
import '../../../settings/presentation/screens/settings_screen.dart';
import '../../../../core/localization/app_localizations.dart';

class MainLayout extends StatefulWidget {
  const MainLayout({super.key});

  @override
  State<MainLayout> createState() => _MainLayoutState();
}

class _MainLayoutState extends State<MainLayout> {
  int _currentIndex = 1; // Liste sekmesi varsayılan
  late final PageController _pageController;

  static const _pages = <Widget>[
    HubScreen(),
    ShoppingListScreen(),
    SettingsScreen(),
  ];



  @override
  void initState() {
    super.initState();
    _pageController = PageController(initialPage: _currentIndex);
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _onNavTap(int index) {
    if (index == _currentIndex) return;
    setState(() => _currentIndex = index);
    _pageController.animateToPage(
      index,
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
    );
  }

  void _onPageChanged(int index) {
    setState(() => _currentIndex = index);
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);

    // Update nav items using translation keys
    final translatedNavItems = [
      _NavItem(icon: Icons.home_outlined, activeIcon: Icons.home_rounded, label: l.t('nav.hub')),
      _NavItem(icon: Icons.list_alt_outlined, activeIcon: Icons.list_alt_rounded, label: l.t('nav.list')),
      _NavItem(icon: Icons.settings_outlined, activeIcon: Icons.settings_rounded, label: l.t('nav.settings')),
    ];

    return Scaffold(
      backgroundColor: AppTheme.bgBlack,
      extendBody: true,
      body: PageView(
        controller: _pageController,
        onPageChanged: _onPageChanged,
        physics: const BouncingScrollPhysics(),
        children: _pages,
      ),
      floatingActionButton: _PersistentFab(
        onPressed: () => _openAddProductSheet(context),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.endFloat,
      bottomNavigationBar: _BottomNavBar(
        currentIndex: _currentIndex,
        items: translatedNavItems,
        onTap: _onNavTap,
      ),
    );
  }

  void _openAddProductSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => _AddProductSheet(),
    );
  }
}

// ─── Nav Item Data ────────────────────────────────────────────────────────────

class _NavItem {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  const _NavItem({
    required this.icon,
    required this.activeIcon,
    required this.label,
  });
}

// ─── Netflix Dark Bottom Navbar ───────────────────────────────────────────────

class _BottomNavBar extends StatelessWidget {
  final int currentIndex;
  final List<_NavItem> items;
  final ValueChanged<int> onTap;

  const _BottomNavBar({
    required this.currentIndex,
    required this.items,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: AppTheme.bgCard,
        border: Border(
          top: BorderSide(color: AppTheme.borderDefault, width: 0.5),
        ),
      ),
      child: SafeArea(
        child: SizedBox(
          height: 58,
          child: Row(
            children: List.generate(items.length, (i) {
              final item = items[i];
              final isSelected = i == currentIndex;
              return Expanded(
                child: _NavTabButton(
                  icon: item.icon,
                  activeIcon: item.activeIcon,
                  label: item.label,
                  isSelected: isSelected,
                  onTap: () => onTap(i),
                ),
              );
            }),
          ),
        ),
      ),
    );
  }
}

class _NavTabButton extends StatelessWidget {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _NavTabButton({
    required this.icon,
    required this.activeIcon,
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Kırmızı gösterge çizgisi (üst)
          AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            height: 2,
            width: isSelected ? 28 : 0,
            margin: const EdgeInsets.only(bottom: 6),
            decoration: BoxDecoration(
              color: AppTheme.netflixRed,
              borderRadius: BorderRadius.circular(1),
            ),
          ),
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 180),
            child: Icon(
              isSelected ? activeIcon : icon,
              key: ValueKey('$icon-$isSelected'),
              color: isSelected ? AppTheme.netflixRed : AppTheme.textDisabled,
              size: 22,
            ),
          ),
          const SizedBox(height: 3),
          AnimatedDefaultTextStyle(
            duration: const Duration(milliseconds: 180),
            style: TextStyle(
              fontSize: 10,
              fontWeight: isSelected ? FontWeight.w700 : FontWeight.w400,
              color: isSelected ? AppTheme.netflixRed : AppTheme.textDisabled,
            ),
            child: Text(label),
          ),
        ],
      ),
    );
  }
}

// ─── Persistent FAB ───────────────────────────────────────────────────────────

class _PersistentFab extends StatefulWidget {
  final VoidCallback onPressed;
  const _PersistentFab({required this.onPressed});

  @override
  State<_PersistentFab> createState() => _PersistentFabState();
}

class _PersistentFabState extends State<_PersistentFab>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 120),
      lowerBound: 0.86,
      upperBound: 1.0,
      value: 1.0,
    );
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _handleTap() async {
    await _ctrl.reverse();
    await _ctrl.forward();
    widget.onPressed();
  }

  @override
  Widget build(BuildContext context) {
    return ScaleTransition(
      scale: _ctrl,
      child: GestureDetector(
        onTap: _handleTap,
        child: Container(
          width: 54,
          height: 54,
          decoration: BoxDecoration(
            color: AppTheme.netflixRed,
            borderRadius: BorderRadius.circular(14),
            boxShadow: [
              BoxShadow(
                color: AppTheme.netflixRedGlow.withValues(alpha: 0.50),
                blurRadius: 20,
                offset: const Offset(0, 6),
              ),
            ],
          ),
          child: const Icon(Icons.add_rounded, color: Colors.white, size: 28),
        ),
      ),
    );
  }
}

// ─── Ürün Ekleme BottomSheet ──────────────────────────────────────────────────

class _AddProductSheet extends StatefulWidget {
  @override
  State<_AddProductSheet> createState() => _AddProductSheetState();
}

class _AddProductSheetState extends State<_AddProductSheet> {
  final _formKey  = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _qtyCtrl  = TextEditingController(text: '1');
  final _unitCtrl = TextEditingController();
  String? _selectedCategory;
  bool _isSaving = false;



  @override
  void dispose() {
    _nameCtrl.dispose();
    _qtyCtrl.dispose();
    _unitCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    final auth = context.read<AuthProvider>();
    final tenantId = auth.currentTenantId;
    if (tenantId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Önce bir ev seçmelisiniz.')),
      );
      return;
    }
    setState(() => _isSaving = true);
    final qty = int.tryParse(_qtyCtrl.text.trim()) ?? 1;
    final unit = _unitCtrl.text.trim();
    await context.read<ProductProvider>().addItem(
          tenantId: tenantId,
          name: _nameCtrl.text.trim(),
          quantity: qty,
          category: _selectedCategory,
          unit: unit.isNotEmpty ? unit : null,
        );
    if (mounted) Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);
    final localizedCategories = [
      ('produce',   '🥦 ${l.t('product.categories.produce')}'),
      ('dairy',     '🥛 ${l.t('product.categories.dairy')}'),
      ('meat',      '🥩 ${l.t('product.categories.meat')}'),
      ('bakery',    '🍞 ${l.t('product.categories.bakery')}'),
      ('beverages', '🧃 ${l.t('product.categories.beverages')}'),
      ('cleaning',  '🧹 ${l.t('product.categories.cleaning')}'),
      ('personal',  '🧴 ${l.t('product.categories.personal')}'),
      ('other',     '📦 ${l.t('product.categories.other')}'),
    ];

    return Container(
      decoration: const BoxDecoration(
        color: AppTheme.bgSheet,
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      child: SafeArea(
        child: SingleChildScrollView(
          padding: EdgeInsets.fromLTRB(20, 14, 20, 20 + MediaQuery.of(context).viewInsets.bottom),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Handle
                Center(
                  child: Container(
                    width: 36,
                    height: 4,
                    decoration: BoxDecoration(
                      color: AppTheme.borderDefault,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                const SizedBox(height: 16),

                // Başlık + kırmızı çizgi
                Row(
                  children: [
                    Container(width: 3, height: 20,
                      decoration: BoxDecoration(
                        color: AppTheme.netflixRed,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Text(
                      l.t('add_product.title'),
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
                  children: localizedCategories.map((cat) {
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
                          child: Text(l.t('product.add_item')),
                        ),
                ),
              ],
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
