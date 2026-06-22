// lib/features/home/presentation/screens/hub_screen.dart
//
// Evim (Hub) ekranı - No Home State ve Dashboard geçiş simülasyonu.
// Netflix Dark temasına sadık, ortalanmış (max 450px) ve responsive.

import 'package:flutter/material.dart' as m;
import 'package:shadcn_flutter/shadcn_flutter.dart';
import '../../../../core/localization/app_localizations.dart';
import '../../../../core/theme/app_theme.dart';

class HubScreen extends StatefulWidget {
  const HubScreen({super.key});

  @override
  State<HubScreen> createState() => _HubScreenState();
}

class _HubScreenState extends State<HubScreen> {
  bool _hasHome = false;
  String _homeName = "";

  void _openCreateHomeSheet(m.BuildContext context) {
    m.showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: m.Colors.transparent,
      builder: (_) => _CreateHomeSheet(
        onSuccess: (name) {
          setState(() {
            _hasHome = true;
            _homeName = name;
          });
        },
      ),
    );
  }

  void _openJoinHomeSheet(m.BuildContext context) {
    m.showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: m.Colors.transparent,
      builder: (_) => _JoinHomeSheet(
        onSuccess: (name) {
          setState(() {
            _hasHome = true;
            _homeName = name;
          });
        },
      ),
    );
  }

  @override
  m.Widget build(m.BuildContext context) {
    final l = AppLocalizations.of(context);

    if (_hasHome) {
      return _DashboardView(
        homeName: _homeName.isNotEmpty ? _homeName : l.t('tenant.dashboard_family_name'),
        onLeaveHome: () {
          setState(() {
            _hasHome = false;
            _homeName = "";
          });
        },
        onInvite: () {
          // Kodu Paylaş / Davet Et simülasyonu (İleride clipboard veya paylaşım eklenebilir)
        },
      );
    }

    return m.Center(
      child: m.SingleChildScrollView(
        padding: const m.EdgeInsets.symmetric(horizontal: 24, vertical: 48),
        child: m.ConstrainedBox(
          constraints: const m.BoxConstraints(maxWidth: 450),
          child: m.Column(
            mainAxisSize: m.MainAxisSize.min,
            crossAxisAlignment: m.CrossAxisAlignment.stretch,
            children: [
              // ── Glowing Logo/Icon ──────────────────────────────────
              m.Center(
                child: m.Container(
                  width: 80,
                  height: 80,
                  decoration: m.BoxDecoration(
                    color: AppTheme.bgCard,
                    border: m.Border.all(color: AppTheme.netflixRed.withValues(alpha: 0.5), width: 1.5),
                    borderRadius: m.BorderRadius.circular(24),
                    boxShadow: [
                      m.BoxShadow(
                        color: AppTheme.netflixRed.withValues(alpha: 0.25),
                        blurRadius: 24,
                        offset: const m.Offset(0, 8),
                      ),
                    ],
                  ),
                  child: const m.Icon(
                    m.Icons.home_rounded,
                    color: AppTheme.netflixRed,
                    size: 38,
                  ),
                ),
              ),
              const m.SizedBox(height: 28),

              // ── Başlıklar ──────────────────────────────────────────
              Text('Kap-App')
                  .h2()
                  .bold()
                  .textCenter()
                  .foreground(),
              const m.SizedBox(height: 8),
              Text(l.t('tenant.no_home'))
                  .small()
                  .muted()
                  .textCenter()
                  .foreground(),
              const m.SizedBox(height: 40),

              // ── Seçenek Kartları ───────────────────────────────────
              _buildOptionCard(
                context: context,
                icon: m.Icons.add_circle_outline_rounded,
                title: l.t('tenant.create_home'),
                description: l.t('tenant.create_home_desc'),
                onTap: () => _openCreateHomeSheet(context),
              ),
              const m.SizedBox(height: 16),
              _buildOptionCard(
                context: context,
                icon: m.Icons.group_add_outlined,
                title: l.t('tenant.join_home'),
                description: l.t('tenant.join_home_desc'),
                onTap: () => _openJoinHomeSheet(context),
              ),
            ],
          ),
        ),
      ),
    );
  }

  m.Widget _buildOptionCard({
    required m.BuildContext context,
    required m.IconData icon,
    required String title,
    required String description,
    required m.VoidCallback onTap,
  }) {
    return m.Container(
      decoration: m.BoxDecoration(
        color: AppTheme.bgCard,
        border: m.Border.all(color: AppTheme.borderDefault, width: 0.8),
        borderRadius: m.BorderRadius.circular(14),
      ),
      child: m.Material(
        color: m.Colors.transparent,
        child: m.InkWell(
          onTap: onTap,
          borderRadius: m.BorderRadius.circular(14),
          splashColor: AppTheme.netflixRed.withValues(alpha: 0.08),
          highlightColor: AppTheme.netflixRed.withValues(alpha: 0.04),
          child: m.Padding(
            padding: const m.EdgeInsets.all(20),
            child: m.Row(
              children: [
                m.Container(
                  padding: const m.EdgeInsets.all(10),
                  decoration: m.BoxDecoration(
                    color: AppTheme.bgElevated,
                    borderRadius: m.BorderRadius.circular(10),
                  ),
                  child: m.Icon(icon, color: AppTheme.netflixRed, size: 24),
                ),
                const m.SizedBox(width: 16),
                m.Expanded(
                  child: m.Column(
                    crossAxisAlignment: m.CrossAxisAlignment.start,
                    children: [
                      Text(title)
                          .semiBold()
                          .large()
                          .foreground(),
                      const m.SizedBox(height: 4),
                      Text(description)
                          .small()
                          .muted()
                          .foreground(),
                    ],
                  ),
                ),
                const m.SizedBox(width: 8),
                const m.Icon(
                  m.Icons.chevron_right_rounded,
                  color: AppTheme.textSecondary,
                  size: 20,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ─── Dashboard Görünümü (Senaryo 2) ─────────────────────────────────────────

class _DashboardView extends StatelessWidget {
  final String homeName;
  final m.VoidCallback onLeaveHome;
  final m.VoidCallback onInvite;

  const _DashboardView({
    required this.homeName,
    required this.onLeaveHome,
    required this.onInvite,
  });

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context);

    return m.Center(
      child: m.SingleChildScrollView(
        padding: const m.EdgeInsets.symmetric(horizontal: 24, vertical: 48),
        child: m.ConstrainedBox(
          constraints: const m.BoxConstraints(maxWidth: 450),
          child: m.Column(
            crossAxisAlignment: m.CrossAxisAlignment.stretch,
            children: [
              // ── Üst Kısım (Başlık & Davet Butonu) ────────────────────────
              m.Row(
                mainAxisAlignment: m.MainAxisAlignment.spaceBetween,
                children: [
                  m.Expanded(
                    child: Text(homeName)
                        .h1()
                        .bold()
                        .ellipsis()
                        .foreground(),
                  ),
                  Button.ghost(
                    onPressed: onInvite,
                    child: const Icon(
                      m.Icons.share_rounded,
                      color: AppTheme.netflixRed,
                      size: 22,
                    ),
                  ),
                ],
              ),
              const m.SizedBox(height: 32),

              // ── Orta Kısım (Netflix Dark Summary Card) ──────────────────
              m.Container(
                decoration: m.BoxDecoration(
                  color: AppTheme.bgCard,
                  border: m.Border.all(color: AppTheme.netflixRed.withValues(alpha: 0.3), width: 1),
                  borderRadius: m.BorderRadius.circular(16),
                  boxShadow: [
                    m.BoxShadow(
                      color: AppTheme.netflixRed.withValues(alpha: 0.1),
                      blurRadius: 16,
                      offset: const m.Offset(0, 4),
                    ),
                  ],
                ),
                child: m.Material(
                  color: m.Colors.transparent,
                  child: m.InkWell(
                    onTap: () {
                      // Listeye gitme simülasyonu
                    },
                    borderRadius: m.BorderRadius.circular(16),
                    splashColor: AppTheme.netflixRed.withValues(alpha: 0.1),
                    highlightColor: AppTheme.netflixRed.withValues(alpha: 0.05),
                    child: m.Padding(
                      padding: const m.EdgeInsets.all(24),
                      child: m.Column(
                        crossAxisAlignment: m.CrossAxisAlignment.start,
                        children: [
                          m.Row(
                            children: [
                              m.Container(
                                padding: const m.EdgeInsets.all(8),
                                decoration: m.BoxDecoration(
                                  color: AppTheme.netflixRed.withValues(alpha: 0.1),
                                  borderRadius: m.BorderRadius.circular(8),
                                ),
                                child: const Icon(
                                  m.Icons.shopping_cart_rounded,
                                  color: AppTheme.netflixRed,
                                  size: 20,
                                ),
                              ),
                              const m.SizedBox(width: 12),
                              m.Expanded(
                                child: Text(l.t('tenant.dashboard_list_summary'))
                                    .semiBold()
                                    .foreground(),
                              ),
                            ],
                          ),
                          const m.SizedBox(height: 20),
                          m.Row(
                            mainAxisAlignment: m.MainAxisAlignment.end,
                            children: [
                              Text(
                                l.t('tenant.dashboard_go_to_list'),
                                style: const TextStyle(color: AppTheme.netflixRed),
                              )
                                  .small()
                                  .bold(),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              const m.SizedBox(height: 40),

              // ── Alt Kısım (Üyeler / Avatarlar) ─────────────────────────
              Text(l.t('tenant.members'))
                  .semiBold()
                  .muted()
                  .foreground(),
              const m.SizedBox(height: 12),
              m.Row(
                children: [
                  AvatarGroup.toRight(
                    children: const [
                      Avatar(initials: 'YÖ', backgroundColor: Color(0xFF2C3E50)),
                      Avatar(initials: 'MY', backgroundColor: Color(0xFF16A085)),
                      Avatar(initials: 'AY', backgroundColor: Color(0xFF2980B9)),
                    ],
                  ),
                  const m.SizedBox(width: 12),
                  Text('+3')
                      .small()
                      .muted()
                      .foreground(),
                ],
              ),
              const m.SizedBox(height: 48),

              // ── Evden Ayrıl Butonu ──────────────────────────────────────
              Button.destructive(
                onPressed: onLeaveHome,
                child: Text(
                  l.t('tenant.leave_home'),
                  style: const TextStyle(
                    fontWeight: m.FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Yeni Ev Oluştur Bottom Sheet ──────────────────────────────────────────

class _CreateHomeSheet extends m.StatefulWidget {
  final ValueChanged<String> onSuccess;
  
  const _CreateHomeSheet({required this.onSuccess});

  @override
  m.State<_CreateHomeSheet> createState() => _CreateHomeSheetState();
}

class _CreateHomeSheetState extends m.State<_CreateHomeSheet> {
  final m.TextEditingController _nameController = m.TextEditingController();
  bool _isLoading = false;

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  m.Widget build(m.BuildContext context) {
    final l = AppLocalizations.of(context);

    return m.Padding(
      padding: m.EdgeInsets.only(bottom: m.MediaQuery.of(context).viewInsets.bottom),
      child: m.Container(
        decoration: const m.BoxDecoration(
          color: AppTheme.bgSheet,
          borderRadius: m.BorderRadius.vertical(top: m.Radius.circular(20)),
        ),
        child: m.SafeArea(
          child: m.SingleChildScrollView(
            padding: const m.EdgeInsets.fromLTRB(24, 16, 24, 24),
            child: m.Column(
              mainAxisSize: m.MainAxisSize.min,
              crossAxisAlignment: m.CrossAxisAlignment.stretch,
              children: [
                // Handle bar
                m.Center(
                  child: m.Container(
                    width: 36,
                    height: 4,
                    decoration: m.BoxDecoration(
                      color: AppTheme.borderDefault,
                      borderRadius: m.BorderRadius.circular(2),
                    ),
                  ),
                ),
                const m.SizedBox(height: 20),

                // Başlık
                m.Row(
                  children: [
                    m.Container(
                      width: 3,
                      height: 18,
                      decoration: m.BoxDecoration(
                        color: AppTheme.netflixRed,
                        borderRadius: m.BorderRadius.circular(1.5),
                      ),
                    ),
                    const m.SizedBox(width: 8),
                    Text(l.t('tenant.create_home'))
                        .bold()
                        .large()
                        .foreground(),
                  ],
                ),
                const m.SizedBox(height: 24),

                // Ev Adı Inputu
                Text(l.t('tenant.home_name'))
                    .semiBold()
                    .small()
                    .foreground(),
                const m.SizedBox(height: 8),
                TextField(
                  controller: _nameController,
                  placeholder: Text(l.t('tenant.home_name')),
                  features: const [
                    InputLeadingFeature(
                      m.Icon(m.Icons.home_outlined, size: 18, color: AppTheme.textSecondary),
                    ),
                  ],
                ),
                const m.SizedBox(height: 24),

                // Oluştur Butonu
                Button.primary(
                  onPressed: _isLoading ? null : () async {
                    final text = _nameController.text.trim();
                    if (text.isEmpty) return;
                    setState(() => _isLoading = true);
                    // Mock API call
                    await Future.delayed(const Duration(milliseconds: 1000));
                    if (context.mounted) {
                      widget.onSuccess(text);
                      m.Navigator.pop(context);
                    }
                  },
                  child: _isLoading
                      ? const m.SizedBox(
                          width: 20,
                          height: 20,
                          child: m.CircularProgressIndicator(
                            strokeWidth: 2,
                            color: m.Colors.white,
                          ),
                        )
                      : Text(l.t('tenant.create_button')),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ─── Davet Koduyla Katıl Bottom Sheet ────────────────────────────────────────

class _JoinHomeSheet extends m.StatefulWidget {
  final ValueChanged<String> onSuccess;

  const _JoinHomeSheet({required this.onSuccess});

  @override
  m.State<_JoinHomeSheet> createState() => _JoinHomeSheetState();
}

class _JoinHomeSheetState extends m.State<_JoinHomeSheet> {
  final m.TextEditingController _codeController = m.TextEditingController();
  bool _isLoading = false;

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  @override
  m.Widget build(m.BuildContext context) {
    final l = AppLocalizations.of(context);

    return m.Padding(
      padding: m.EdgeInsets.only(bottom: m.MediaQuery.of(context).viewInsets.bottom),
      child: m.Container(
        decoration: const m.BoxDecoration(
          color: AppTheme.bgSheet,
          borderRadius: m.BorderRadius.vertical(top: m.Radius.circular(20)),
        ),
        child: m.SafeArea(
          child: m.SingleChildScrollView(
            padding: const m.EdgeInsets.fromLTRB(24, 16, 24, 24),
            child: m.Column(
              mainAxisSize: m.MainAxisSize.min,
              crossAxisAlignment: m.CrossAxisAlignment.stretch,
              children: [
                // Handle bar
                m.Center(
                  child: m.Container(
                    width: 36,
                    height: 4,
                    decoration: m.BoxDecoration(
                      color: AppTheme.borderDefault,
                      borderRadius: m.BorderRadius.circular(2),
                    ),
                  ),
                ),
                const m.SizedBox(height: 20),

                // Başlık
                m.Row(
                  children: [
                    m.Container(
                      width: 3,
                      height: 18,
                      decoration: m.BoxDecoration(
                        color: AppTheme.netflixRed,
                        borderRadius: m.BorderRadius.circular(1.5),
                      ),
                    ),
                    const m.SizedBox(width: 8),
                    Text(l.t('tenant.join_home'))
                        .bold()
                        .large()
                        .foreground(),
                  ],
                ),
                const m.SizedBox(height: 24),

                // Davet Kodu Inputu
                Text(l.t('tenant.home_code'))
                    .semiBold()
                    .small()
                    .foreground(),
                const m.SizedBox(height: 8),
                TextField(
                  controller: _codeController,
                  placeholder: Text(l.t('tenant.home_code')),
                  features: const [
                    InputLeadingFeature(
                      m.Icon(m.Icons.vpn_key_outlined, size: 18, color: AppTheme.textSecondary),
                    ),
                  ],
                ),
                const m.SizedBox(height: 24),

                // Katıl Butonu
                Button.primary(
                  onPressed: _isLoading ? null : () async {
                    final code = _codeController.text.trim();
                    if (code.isEmpty) return;
                    setState(() => _isLoading = true);
                    // Mock API call
                    await Future.delayed(const Duration(milliseconds: 1000));
                    if (context.mounted) {
                      widget.onSuccess(""); // Boş gönderilirse varsayılan isim kullanılacak
                      m.Navigator.pop(context);
                    }
                  },
                  child: _isLoading
                      ? const m.SizedBox(
                          width: 20,
                          height: 20,
                          child: m.CircularProgressIndicator(
                            strokeWidth: 2,
                            color: m.Colors.white,
                          ),
                        )
                      : Text(l.t('tenant.join_button')),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
