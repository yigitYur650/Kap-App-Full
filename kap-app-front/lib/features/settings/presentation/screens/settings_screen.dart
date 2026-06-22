// lib/features/settings/presentation/screens/settings_screen.dart

import 'package:flutter/material.dart' as m;
import 'package:shadcn_flutter/shadcn_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter_colorpicker/flutter_colorpicker.dart' as cp;
import '../../../../core/theme/app_theme.dart';
import '../../../../core/localization/app_localizations.dart';

// Prototip aşamasında dil seçimini tüm uygulamada (ana layout'ta) anında
// yansıtmak için ValueNotifier kullanıyoruz. (Tam çözüm için main.dart'ta Provider gerekir.)
final m.ValueNotifier<String> globalSelectedLang = m.ValueNotifier<String>('tr');

class SettingsScreen extends m.StatefulWidget {
  const SettingsScreen({super.key});

  @override
  m.State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends m.State<SettingsScreen> {
  double _fontSize = 14;
  m.Color _accentColor = AppTheme.netflixRed;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // _globalSelectedLang ilk açılışta veya sekmeler arası geçişte her zaman kullanılacak.
    // LocaleOf(context) ile ezmeyerek kullanıcının seçimini koruyoruz.
  }

  Future<void> _launchURL(String urlString) async {
    final Uri url = Uri.parse(urlString);
    if (!await launchUrl(url, mode: LaunchMode.externalApplication)) {
      m.debugPrint('Could not launch $urlString');
    }
  }

  @override
  m.Widget build(m.BuildContext context) {
    final l = AppLocalizations.of(context);

    return m.Center(
      child: m.SingleChildScrollView(
        // Dikey padding daraltıldı
        padding: const m.EdgeInsets.symmetric(horizontal: 24, vertical: 16),
        child: m.ConstrainedBox(
          constraints: const m.BoxConstraints(maxWidth: 450),
          child: m.Column(
            crossAxisAlignment: m.CrossAxisAlignment.stretch,
            children: [
              // ── Başlık ──────────────────────────────────────────────
              m.Center(
                child: m.Container(
                  width: 64, // 80'den 64'e düşürüldü
                  height: 64, // 80'den 64'e düşürüldü
                  decoration: m.BoxDecoration(
                    color: AppTheme.bgCard,
                    border: m.Border.all(color: _accentColor.withValues(alpha: 0.5), width: 1.5),
                    borderRadius: m.BorderRadius.circular(20),
                    boxShadow: [
                      m.BoxShadow(
                        color: _accentColor.withValues(alpha: 0.25),
                        blurRadius: 16,
                        offset: const m.Offset(0, 6),
                      ),
                    ],
                  ),
                  child: m.Icon(
                    m.Icons.settings_rounded,
                    color: _accentColor,
                    size: 32, // 38'den 32'ye düşürüldü
                  ),
                ),
              ),
              const m.SizedBox(height: 12),
              Text(l.t('settings.title'))
                  .h2()
                  .bold()
                  .textCenter()
                  .foreground(),
              const m.SizedBox(height: 16), // 24'ten 16'ya düşürüldü

              // ── Görünüm & Tema Bölümü (_ThemeSection) ───────────────
              _ThemeSection(
                l: l, 
                fontSize: _fontSize, 
                selectedLang: globalSelectedLang.value,
                accentColor: _accentColor,
                onFontSizeChanged: (val) => setState(() => _fontSize = val),
                onLangChanged: (val) {
                  if (val != null && val != globalSelectedLang.value) {
                    globalSelectedLang.value = val;
                  }
                },
                onColorChanged: (color) => setState(() => _accentColor = color),
              ),
              const m.SizedBox(height: 12), // 16'dan 12'ye düşürüldü

              // ── Hakkında Bölümü (_AboutSection) ─────────────────────
              _AboutSection(l: l, onLaunchURL: _launchURL),
              const m.SizedBox(height: 16), // 24'ten 16'ya düşürüldü

              // ── Çıkış Yap Butonu ────────────────────────────────────
              Button.destructive(
                onPressed: () {
                  // Çıkış yapma işlemi
                },
                child: Text(
                  l.t('auth.logout').toUpperCase(),
                  style: const m.TextStyle(fontWeight: m.FontWeight.bold, letterSpacing: 1.2),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Görünüm & Tema Bölümü ────────────────────────────────────────────────

class _ThemeSection extends m.StatelessWidget {
  final AppLocalizations l;
  final double fontSize;
  final String selectedLang;
  final m.Color accentColor;
  final m.ValueChanged<double> onFontSizeChanged;
  final m.ValueChanged<String?> onLangChanged;
  final m.ValueChanged<m.Color> onColorChanged;

  const _ThemeSection({
    required this.l,
    required this.fontSize,
    required this.selectedLang,
    required this.accentColor,
    required this.onFontSizeChanged,
    required this.onLangChanged,
    required this.onColorChanged,
  });

  @override
  m.Widget build(m.BuildContext context) {
    return m.Container(
      decoration: m.BoxDecoration(
        color: AppTheme.bgCard,
        border: m.Border.all(color: AppTheme.borderDefault, width: 0.8),
        borderRadius: m.BorderRadius.circular(16),
      ),
      child: m.Column(
        crossAxisAlignment: m.CrossAxisAlignment.stretch,
        children: [
          m.Padding(
            padding: const m.EdgeInsets.all(16), // 20'den 16'ya
            child: m.Row(
              children: [
                m.Icon(m.Icons.palette_outlined, color: accentColor, size: 22),
                const m.SizedBox(width: 12),
                Text(l.t('settings.appearance_and_theme')).semiBold().large().foreground(),
              ],
            ),
          ),
          const m.Divider(height: 1, color: AppTheme.borderDefault),
          m.Padding(
            padding: const m.EdgeInsets.all(16), // 20'den 16'ya
            child: m.Column(
              crossAxisAlignment: m.CrossAxisAlignment.start,
              children: [
                // Dil Seçimi
                Text(l.t('settings.language')).semiBold().small().foreground(),
                const m.SizedBox(height: 8),
                m.Container(
                  padding: const m.EdgeInsets.symmetric(horizontal: 16),
                  decoration: m.BoxDecoration(
                    color: AppTheme.bgElevated,
                    border: m.Border.all(color: AppTheme.borderDefault),
                    borderRadius: m.BorderRadius.circular(8),
                  ),
                  child: m.DropdownButtonHideUnderline(
                    child: m.DropdownButton<String>(
                      value: selectedLang,
                      dropdownColor: AppTheme.bgSheet,
                      icon: const m.Icon(m.Icons.keyboard_arrow_down_rounded, color: AppTheme.textSecondary),
                      style: const m.TextStyle(color: AppTheme.textPrimary, fontSize: 14, fontFamily: 'NetflixSans'),
                      isExpanded: true,
                      items: const [
                        m.DropdownMenuItem(value: 'tr', child: m.Text('Türkçe')),
                        m.DropdownMenuItem(value: 'en', child: m.Text('English')),
                      ],
                      onChanged: onLangChanged,
                    ),
                  ),
                ),
                const m.SizedBox(height: 16), // 24'ten 16'ya

                // Yazı Boyutu
                Text(l.t('settings.font_size')).semiBold().small().foreground(),
                const m.SizedBox(height: 12), // 16'dan 12'ye
                Slider(
                  value: SliderValue.single(fontSize),
                  min: 10,
                  max: 24,
                  onChanged: (val) => onFontSizeChanged(val.value),
                ),
                const m.SizedBox(height: 16), // 24'ten 16'ya

                // Renk Özelleştirme
                Text(l.t('settings.color_customization')).semiBold().small().foreground(),
                const m.SizedBox(height: 12),
                m.Row(
                  children: [
                    _ColorButton(
                      color1: AppTheme.bgBlack,
                      color2: AppTheme.netflixRed,
                      isSelected: accentColor.toARGB32() == AppTheme.netflixRed.toARGB32(),
                      onTap: () => onColorChanged(AppTheme.netflixRed),
                    ),
                    const m.SizedBox(width: 12),
                    _ColorButton(
                      color1: const m.Color(0xFF1A1A2E),
                      color2: const m.Color(0xFFE94560),
                      isSelected: accentColor.toARGB32() == const m.Color(0xFFE94560).toARGB32(),
                      onTap: () => onColorChanged(const m.Color(0xFFE94560)),
                    ),
                    const m.SizedBox(width: 12),
                    Button.outline(
                      onPressed: () {
                        m.Color tempColor = accentColor;
                        m.showDialog(
                          context: context,
                          builder: (context) {
                            return m.AlertDialog(
                              backgroundColor: AppTheme.bgCard,
                              title: const m.Text('Vurgu Rengini Seçin', style: m.TextStyle(color: m.Colors.white)),
                              content: m.SingleChildScrollView(
                                child: cp.ColorPicker(
                                  pickerColor: accentColor,
                                  onColorChanged: (color) {
                                    tempColor = color;
                                  },
                                  enableAlpha: false,
                                  pickerAreaHeightPercent: 0.8,
                                ),
                              ),
                              actions: [
                                m.TextButton(
                                  onPressed: () => m.Navigator.of(context).pop(),
                                  child: const m.Text('İptal', style: m.TextStyle(color: m.Colors.white70)),
                                ),
                                m.TextButton(
                                  onPressed: () {
                                    onColorChanged(tempColor);
                                    m.Navigator.of(context).pop();
                                  },
                                  child: m.Text('Tamam', style: m.TextStyle(color: m.Colors.white)),
                                ),
                              ],
                            );
                          },
                        );
                      },
                      child: const m.Icon(m.Icons.color_lens_outlined, size: 24),
                    ),
                  ],
                ),
              ],
            ),
          ),
          // UX Notu: Hata sarısından çıkarılıp şık Netflix Dark moduna alındı
          m.Container(
            padding: const m.EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: m.BoxDecoration(
              color: AppTheme.bgElevated.withValues(alpha: 0.5), // Transparan Elevated bg
              borderRadius: const m.BorderRadius.vertical(bottom: m.Radius.circular(16)),
              border: m.Border(top: m.BorderSide(color: accentColor.withValues(alpha: 0.3))),
            ),
            child: m.Row(
              children: [
                m.Icon(m.Icons.auto_awesome_rounded, color: accentColor, size: 20),
                const m.SizedBox(width: 12),
                m.Expanded(
                  child: Text(
                    l.t('settings.color_note'),
                    style: const m.TextStyle(color: m.Colors.white70, fontWeight: m.FontWeight.w500),
                  ).small(),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ColorButton extends m.StatelessWidget {
  final m.Color color1;
  final m.Color color2;
  final bool isSelected;
  final m.VoidCallback onTap;

  const _ColorButton({
    required this.color1,
    required this.color2,
    required this.isSelected,
    required this.onTap,
  });

  @override
  m.Widget build(m.BuildContext context) {
    return m.GestureDetector(
      onTap: onTap,
      child: m.Container(
        width: 48,
        height: 48,
        decoration: m.BoxDecoration(
          shape: m.BoxShape.circle,
          border: m.Border.all(
            color: isSelected ? m.Colors.white : AppTheme.borderDefault,
            width: isSelected ? 2 : 1,
          ),
          boxShadow: isSelected
              ? [m.BoxShadow(color: color2.withValues(alpha: 0.4), blurRadius: 8, offset: const m.Offset(0, 4))]
              : null,
        ),
        child: m.ClipOval(
          child: m.Row(
            children: [
              m.Expanded(child: m.Container(color: color1)),
              m.Expanded(child: m.Container(color: color2)),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Hakkında Bölümü ──────────────────────────────────────────────────────

class _AboutSection extends m.StatelessWidget {
  final AppLocalizations l;
  final m.ValueChanged<String> onLaunchURL;

  const _AboutSection({required this.l, required this.onLaunchURL});

  @override
  m.Widget build(m.BuildContext context) {
    return m.Container(
      decoration: m.BoxDecoration(
        color: AppTheme.bgCard,
        border: m.Border.all(color: AppTheme.borderDefault, width: 0.8),
        borderRadius: m.BorderRadius.circular(16),
      ),
      child: m.Column(
        crossAxisAlignment: m.CrossAxisAlignment.stretch,
        children: [
          m.Padding(
            padding: const m.EdgeInsets.all(16), // 20'den 16'ya
            child: m.Row(
              children: [
                const m.Icon(m.Icons.info_outline_rounded, color: AppTheme.textSecondary, size: 22),
                const m.SizedBox(width: 12),
                Text(l.t('settings.about')).semiBold().large().foreground(),
              ],
            ),
          ),
          const m.Divider(height: 1, color: AppTheme.borderDefault),
          m.Padding(
            padding: const m.EdgeInsets.all(16), // 20'den 16'ya
            child: m.Column(
              crossAxisAlignment: m.CrossAxisAlignment.stretch,
              children: [
                _TerminalInfoRow(label: l.t('settings.app_version_label'), value: 'v1.0.0'),
                const m.SizedBox(height: 8), // 12'den 8'e
                _TerminalInfoRow(label: l.t('settings.build_number_label'), value: '#42'),
                const m.SizedBox(height: 8), // 12'den 8'e
                _TerminalInfoRow(label: l.t('settings.release_date_label'), value: l.t('settings.release_date_value')),
                const m.SizedBox(height: 20), // 32'den 20'ye
                
                // Butonlar
                m.Row(
                  children: [
                    m.Expanded(
                      child: Button.secondary(
                        onPressed: () => onLaunchURL('https://www.instagram.com/yigityurrr?igsh=MTNvOWoweTNvajNlaA=='),
                        child: Text(l.t('settings.send_feedback')).small().textCenter(),
                      ),
                    ),
                    const m.SizedBox(width: 12),
                    m.Expanded(
                      child: Button.outline(
                        onPressed: () => onLaunchURL('https://github.com/yigitYur650'),
                        child: Text(l.t('settings.follow_github')).small().textCenter(),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _TerminalInfoRow extends m.StatelessWidget {
  final String label;
  final String value;

  const _TerminalInfoRow({required this.label, required this.value});

  @override
  m.Widget build(m.BuildContext context) {
    return m.Row(
      crossAxisAlignment: m.CrossAxisAlignment.start,
      children: [
        Text('$label: ').muted().small().mono(),
        m.Expanded(
          child: Text(value).foreground().small().bold().mono(),
        ),
      ],
    );
  }
}
