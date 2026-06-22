// lib/features/auth/presentation/screens/login_screen.dart
//
// PWA-uyumlu, responsive Giriş/Kayıt ekranı - Netflix Dark Konsepti.
// Max genişlik: 450px — tarayıcıda dikey/yatay ortala, mobilde tam ekran.

import 'package:flutter/material.dart' as m;
import 'package:provider/provider.dart';
import 'package:shadcn_flutter/shadcn_flutter.dart';
import '../../../../core/localization/app_localizations.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../home/presentation/screens/main_layout.dart';
import '../../providers/auth_provider.dart';

class LoginScreen extends m.StatefulWidget {
  const LoginScreen({super.key});

  @override
  m.State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends m.State<LoginScreen> {
  final m.TextEditingController _emailController = m.TextEditingController();
  final m.TextEditingController _nameController = m.TextEditingController();
  final m.TextEditingController _passwordController = m.TextEditingController();
  
  bool _isLoginMode = true;
  bool _isLoading = false;
  String? _errorText;

  @override
  void dispose() {
    _emailController.dispose();
    _nameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  bool _isValidEmail(String email) {
    return RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$').hasMatch(email.trim());
  }

  Future<void> _submitForm() async {
    final l = AppLocalizations.of(context);
    final email = _emailController.text.trim();
    final password = _passwordController.text.trim();
    final name = _nameController.text.trim();

    if (email.isEmpty || password.isEmpty) {
      setState(() => _errorText = "Lütfen tüm alanları doldurun.");
      return;
    }

    if (!_isValidEmail(email)) {
      setState(() => _errorText = l.t('auth.login_error'));
      return;
    }

    if (password.length < 6) {
      setState(() => _errorText = "Şifre en az 6 karakter olmalıdır.");
      return;
    }

    if (!_isLoginMode && name.isEmpty) {
      setState(() => _errorText = "Lütfen adınızı girin.");
      return;
    }

    setState(() {
      _isLoading = true;
      _errorText = null;
    });

    final authProvider = context.read<AuthProvider>();
    
    if (_isLoginMode) {
      await authProvider.login(email, password);
    } else {
      await authProvider.register(email, password, name);
    }

    if (!mounted) return;
    setState(() => _isLoading = false);

    if (authProvider.status == AuthStatus.error) {
      setState(() => _errorText = authProvider.errorMessage);
    } else if (authProvider.isAuthenticated) {
      m.Navigator.of(context).pushReplacement(
        m.MaterialPageRoute(builder: (_) => const MainLayout()),
      );
    }
  }

  void _toggleMode() {
    setState(() {
      _isLoginMode = !_isLoginMode;
      _errorText = null;
      _passwordController.clear();
      _nameController.clear();
    });
  }

  @override
  m.Widget build(m.BuildContext context) {
    final l = AppLocalizations.of(context);

    return Scaffold(
      backgroundColor: AppTheme.bgBlack,
      child: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 48),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 450),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // ── Logo & Başlık ──────────────────────────────────
                _buildHeader(l),
                const SizedBox(height: 36),

                // ── Kart İçeriği ───────────────────────────────────
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(28),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        // İsim etiketi ve girdisi (Sadece Kayıt modunda görünür)
                        if (!_isLoginMode) ...[
                          Text(l.t('auth.name'))
                              .semiBold()
                              .small()
                              .foreground(),
                          const SizedBox(height: 8),
                          TextField(
                            controller: _nameController,
                            placeholder: Text(l.t('auth.name_hint')),
                            keyboardType: TextInputType.name,
                            onSubmitted: (_) => _submitForm(),
                            features: [
                              const InputLeadingFeature(
                                m.Icon(m.Icons.person_outline_rounded, size: 18, color: AppTheme.textSecondary),
                              ),
                            ],
                          ),
                          const SizedBox(height: 20),
                        ],

                        // E-posta etiketi
                        Text(l.t('auth.email'))
                            .semiBold()
                            .small()
                            .foreground(),
                        const SizedBox(height: 8),

                        // E-posta girdi
                        TextField(
                          controller: _emailController,
                          placeholder: Text(l.t('auth.email_hint')),
                          keyboardType: TextInputType.emailAddress,
                          onSubmitted: (_) => _submitForm(),
                          features: [
                            const InputLeadingFeature(
                              m.Icon(m.Icons.email_outlined, size: 18, color: AppTheme.textSecondary),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),

                        // Şifre etiketi
                        Text(l.t('auth.password'))
                            .semiBold()
                            .small()
                            .foreground(),
                        const SizedBox(height: 8),

                        // Şifre girdi
                        TextField(
                          controller: _passwordController,
                          placeholder: Text(l.t('auth.password_hint')),
                          obscureText: true,
                          keyboardType: TextInputType.visiblePassword,
                          onSubmitted: (_) => _submitForm(),
                          features: [
                            const InputLeadingFeature(
                              m.Icon(m.Icons.lock_outlined, size: 18, color: AppTheme.textSecondary),
                            ),
                            const InputPasswordToggleFeature(
                              icon: m.Icon(m.Icons.visibility_outlined, size: 18, color: AppTheme.textSecondary),
                              iconShow: m.Icon(m.Icons.visibility_off_outlined, size: 18, color: AppTheme.textSecondary),
                            ),
                          ],
                        ),

                        // Hata mesajı
                        if (_errorText != null) ...[
                          const SizedBox(height: 12),
                          Text(
                            _errorText!,
                            style: const TextStyle(
                              color: AppTheme.error,
                              fontWeight: m.FontWeight.w500,
                            ),
                          ).small(),
                        ],

                        const SizedBox(height: 28),

                        // Giriş Yap / Kayıt Ol Butonu (Netflix Red)
                        Button.primary(
                          onPressed: _isLoading ? null : _submitForm,
                          child: _isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: m.CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Color(0xFFFFFFFF),
                                  ),
                                )
                              : Text(
                                  _isLoginMode ? l.t('auth.login') : l.t('auth.register'),
                                  style: const TextStyle(
                                    fontWeight: m.FontWeight.bold,
                                    fontSize: 16,
                                  ),
                                ),
                        ),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 28),

                // Alt bilgi metni (Hesabın yok mu? Kayıt Ol)
                m.GestureDetector(
                  onTap: _isLoading ? null : _toggleMode,
                  child: m.MouseRegion(
                    cursor: m.SystemMouseCursors.click,
                    child: m.Center(
                      child: m.RichText(
                        text: m.TextSpan(
                          style: const m.TextStyle(
                            color: AppTheme.textSecondary,
                            fontSize: 14,
                          ),
                          children: [
                            m.TextSpan(
                              text: _isLoginMode
                                  ? "${l.t('auth.no_account')} "
                                  : "${l.t('auth.have_account')} ",
                            ),
                            m.TextSpan(
                              text: _isLoginMode ? l.t('auth.register') : l.t('auth.login'),
                              style: const m.TextStyle(
                                color: AppTheme.netflixRed,
                                fontWeight: m.FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(AppLocalizations l) {
    return Column(
      children: [
        // Netflix-style glowing logo mark
        m.Container(
          width: 68,
          height: 68,
          decoration: m.BoxDecoration(
            color: AppTheme.bgCard,
            border: m.Border.all(color: AppTheme.netflixRed.withValues(alpha: 0.5), width: 1.5),
            borderRadius: m.BorderRadius.circular(18),
            boxShadow: [
              m.BoxShadow(
                color: AppTheme.netflixRed.withValues(alpha: 0.2),
                blurRadius: 16,
                offset: const m.Offset(0, 4),
              ),
            ],
          ),
          child: const m.Icon(
            m.Icons.shopping_bag,
            color: AppTheme.netflixRed,
            size: 32,
          ),
        ),
        const SizedBox(height: 20),
        Text('Kap-App', style: const TextStyle(letterSpacing: 0.5))
            .h2()
            .bold()
            .foreground(),
        const SizedBox(height: 6),
        Text(_isLoginMode ? l.t('auth.login') : l.t('auth.register'))
            .large()
            .muted()
            .foreground(),
      ],
    );
  }
}
