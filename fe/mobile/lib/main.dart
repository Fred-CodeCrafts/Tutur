import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'context/auth_context.dart';
import 'pages/login_page.dart';
import 'pages/home_page.dart';
import 'pages/practice_page.dart';
import 'pages/contribute_page.dart';
import 'pages/profile_page.dart';
import 'utils/app_theme.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthContext()),
      ],
      child: const NusabahasaApp(),
    ),
  );
}

class NusabahasaApp extends StatelessWidget {
  const NusabahasaApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Nusabahasa',
      theme: buildAppTheme(),
      debugShowCheckedModeBanner: false,
      home: const RootScreen(),
    );
  }
}

class RootScreen extends StatelessWidget {
  const RootScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthContext>();
    return auth.isLoggedIn ? const MainShell() : const LoginPage();
  }
}

class MainShell extends StatefulWidget {
  const MainShell({super.key});

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _selectedIndex = 0;

  final List<_NavItem> _navItems = const [
    _NavItem(icon: Icons.home_outlined, activeIcon: Icons.home_rounded, label: 'Home'),
    _NavItem(
        icon: Icons.play_circle_outline_rounded,
        activeIcon: Icons.play_circle_rounded,
        label: 'Practice'),
    _NavItem(
        icon: Icons.add_circle_outline_rounded,
        activeIcon: Icons.add_circle_rounded,
        label: 'Contribute'),
    _NavItem(
        icon: Icons.person_outline_rounded,
        activeIcon: Icons.person_rounded,
        label: 'Profile'),
  ];

  String get _pageTitle {
    switch (_selectedIndex) {
      case 0:
        return 'Nusabahasa Home';
      case 1:
        return 'Practice Mode';
      case 2:
        return 'Submit New Phrase';
      case 3:
        return 'Review Phrases';
      default:
        return 'Nusabahasa';
    }
  }

  Widget get _currentPage {
    switch (_selectedIndex) {
      case 0:
        return const HomePage();
      case 1:
        return const PracticePage();
      case 2:
        return const ContributePage();
      case 3:
        return const ProfilePage();
      default:
        return const HomePage();
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.read<AuthContext>();

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: PreferredSize(
        preferredSize: const Size.fromHeight(60),
        child: _TopAppBar(
          title: _pageTitle,
          showBack: _selectedIndex != 0,
          onBack: () => setState(() => _selectedIndex = 0),
          onLogout: auth.logout,
        ),
      ),
      body: AnimatedSwitcher(
        duration: const Duration(milliseconds: 250),
        child: KeyedSubtree(
          key: ValueKey(_selectedIndex),
          child: _currentPage,
        ),
      ),
      bottomNavigationBar: _BottomNavBar(
        selectedIndex: _selectedIndex,
        items: _navItems,
        onTap: (i) => setState(() => _selectedIndex = i),
      ),
    );
  }
}

// ─────────────────────────────────────────────────
// Top App Bar
// ─────────────────────────────────────────────────
class _TopAppBar extends StatelessWidget {
  final String title;
  final bool showBack;
  final VoidCallback onBack;
  final VoidCallback onLogout;

  const _TopAppBar({
    required this.title,
    required this.showBack,
    required this.onBack,
    required this.onLogout,
  });

  @override
  Widget build(BuildContext context) {
    return AppBar(
      backgroundColor: AppColors.surface,
      elevation: 0,
      shadowColor: Colors.transparent,
      surfaceTintColor: Colors.transparent,
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(1),
        child: Container(
          height: 1,
          color: AppColors.onSurface..withValues(alpha: 0.08),
        ),
      ),
      leading: showBack
          ? IconButton(
              onPressed: onBack,
              icon: const Icon(Icons.arrow_back_rounded, color: AppColors.primary),
            )
          : Padding(
              padding: const EdgeInsets.all(8),
              child: Container(
                decoration: BoxDecoration(
                  color: AppColors.primary,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: const Icon(Icons.language, color: Colors.white, size: 22),
              ),
            ),
      title: Text(
        title,
        style: AppTextStyles.titleMedium,
      ),
      actions: [
        const Icon(Icons.language_rounded,
            color: AppColors.primary, size: 22),
        const SizedBox(width: 8),
        ClipOval(
          child: Image.network(
            'https://lh3.googleusercontent.com/aida-public/AB6AXuD3HV_SEgDM3Pxu1g6uK5qfHmbxI3OWzNrqpnFHUti0Evse_ZDOIHSKMrdGSrhvFjbwBoXk3WKiwmBU0wDSRN3IR2-sWWsinAo_b4cTX28xegx5hBS-5y2uzn6JVbth3wMFLDZssv3DCUZGxUsXMWveuJclBnmjPFyTxo8qRodJaIkSRutFJvdcTkl8G-1YemJW5RHV9gtQ0GnMPr1ghpS4s--K9aK0WiXQFZqelnMzsgS4SVgt5hapqvtB3YIMYnf2ntLWJtN6Nn73',
            width: 32,
            height: 32,
            fit: BoxFit.cover,
            errorBuilder: (_, __, ___) => const CircleAvatar(
              radius: 16,
              backgroundColor: AppColors.primaryContainer,
              child: Icon(Icons.person_rounded, size: 16, color: Colors.white),
            ),
          ),
        ),
        IconButton(
          onPressed: onLogout,
          icon: const Icon(Icons.logout_rounded,
              color: AppColors.onSurfaceVariant, size: 20),
          tooltip: 'Logout',
        ),
      ],
    );
  }
}

// ─────────────────────────────────────────────────
// Bottom Nav Bar
// ─────────────────────────────────────────────────
class _BottomNavBar extends StatelessWidget {
  final int selectedIndex;
  final List<_NavItem> items;
  final ValueChanged<int> onTap;

  const _BottomNavBar({
    required this.selectedIndex,
    required this.items,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        border: Border(
          top: BorderSide(color: AppColors.onSurface..withValues(alpha: 0.08)),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black..withValues(alpha: 0.05),
            blurRadius: 12,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: List.generate(items.length, (i) {
              final item = items[i];
              final isActive = i == selectedIndex;
              return GestureDetector(
                onTap: () => onTap(i),
                behavior: HitTestBehavior.opaque,
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(
                      horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: isActive
                        ? AppColors.primaryContainer.withValues(alpha: 0.15)
                        : Colors.transparent,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        isActive ? item.activeIcon : item.icon,
                        size: 22,
                        color: isActive
                            ? AppColors.primary
                            : AppColors.onSurfaceVariant,
                      ),
                      const SizedBox(height: 3),
                      Text(
                        item.label,
                        style: TextStyle(
                          fontFamily: 'PlusJakartaSans',
                          fontWeight: FontWeight.w600,
                          fontSize: 11,
                          color: isActive
                              ? AppColors.primary
                              : AppColors.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }),
          ),
        ),
      ),
    );
  }
}

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
