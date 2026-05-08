import 'package:flutter/material.dart';
import '../models/language.dart';
import '../utils/app_theme.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 100),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Active Tracks
          const Text('Active Tracks', style: AppTextStyles.headlineMedium),
          const SizedBox(height: 16),
          SizedBox(
            height: 72,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: kLanguageTracks.length,
              separatorBuilder: (_, __) => const SizedBox(width: 12),
              itemBuilder: (context, i) {
                final lang = kLanguageTracks[i];
                return _LanguageTrackChip(language: lang);
              },
            ),
          ),
          const SizedBox(height: 28),

          // Progress Card
          _ProgressCard(),
          const SizedBox(height: 28),

          // Learning Modes
          const Text('Learning Modes', style: AppTextStyles.headlineMedium),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _LearningModeCard(
                  icon: Icons.history_rounded,
                  title: 'Flashcards',
                  description: 'Master vocabulary with spaced-repetition cards.',
                  onTap: () => Navigator.pushNamed(context, '/practice'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _LearningModeCard(
                  icon: Icons.chat_bubble_outline_rounded,
                  title: 'Conversations',
                  description: 'Roleplay real-world regional dialect scenarios.',
                  onTap: () => Navigator.pushNamed(context, '/practice'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 28),

          // Contribution Banner
          _ContributionBanner(
            onTap: () => Navigator.pushNamed(context, '/contribute'),
          ),
        ],
      ),
    );
  }
}

class _LanguageTrackChip extends StatelessWidget {
  final Language language;

  const _LanguageTrackChip({required this.language});

  @override
  Widget build(BuildContext context) {
    return AnimatedOpacity(
      duration: const Duration(milliseconds: 200),
      opacity: language.isActive ? 1.0 : 0.6,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(
            color: language.isActive
                ? AppColors.primary
                : AppColors.onSurface.withValues(alpha: 0.1),
            width: language.isActive ? 2 : 1,
          ),
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black..withValues(alpha: 0.05),
              blurRadius: 6,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Image.network(
                language.imageUrl,
                width: 32,
                height: 32,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => Container(
                  width: 32,
                  height: 32,
                  decoration: BoxDecoration(
                    color: AppColors.secondaryContainer,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: const Icon(Icons.language, size: 16, color: AppColors.primary),
                ),
              ),
            ),
            const SizedBox(width: 10),
            Text(
              language.name,
              style: const TextStyle(
                fontFamily: 'PlusJakartaSans',
                fontWeight: FontWeight.w600,
                color: AppColors.primary,
                fontSize: 14,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ProgressCard extends StatelessWidget {
  const _ProgressCard();

  @override
  Widget build(BuildContext context) {
    const double progress = 0.75;

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: AppColors.onSurface.withValues(alpha: 0.1)),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8),
        ],
      ),
      child: Row(
        children: [
          SizedBox(
            width: 100,
            height: 100,
            child: Stack(
              alignment: Alignment.center,
              children: [
                SizedBox(
                  width: 100,
                  height: 100,
                  child: CircularProgressIndicator(
                    value: progress,
                    strokeWidth: 8,
                    backgroundColor: AppColors.onSurface..withValues(alpha: 0.05),
                    valueColor:
                        const AlwaysStoppedAnimation<Color>(AppColors.primaryContainer),
                    strokeCap: StrokeCap.round,
                  ),
                ),
                Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      '${(progress * 100).toInt()}%',
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w800,
                        fontSize: 22,
                        color: AppColors.primary,
                      ),
                    ),
                    const Text(
                      'Daily Goal',
                      style: TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontSize: 10,
                        color: AppColors.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(width: 20),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Track Progress', style: AppTextStyles.titleMedium),
                const SizedBox(height: 6),
                const Text(
                  "You're doing great! Only 15 more minutes of Javanese practice to maintain your 5-day streak.",
                  style: AppTextStyles.bodyMedium,
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  children: [
                    _Chip(label: '⚡ 5 Days', color: AppColors.primary),
                    _Chip(label: '📖 Intermediate', color: AppColors.secondary),
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

class _Chip extends StatelessWidget {
  final String label;
  final Color color;

  const _Chip({required this.label, required this.color});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontFamily: 'PlusJakartaSans',
          fontWeight: FontWeight.w700,
          fontSize: 11,
          color: color,
        ),
      ),
    );
  }
}

class _LearningModeCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final String description;
  final VoidCallback onTap;

  const _LearningModeCard({
    required this.icon,
    required this.title,
    required this.description,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(color: AppColors.onSurface.withValues(alpha: 0.1)),
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 6),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: AppColors.primary.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: AppColors.primary, size: 22),
            ),
            const SizedBox(height: 14),
            Text(title, style: AppTextStyles.titleMedium),
            const SizedBox(height: 4),
            Text(description, style: AppTextStyles.bodyMedium.copyWith(fontSize: 12)),
          ],
        ),
      ),
    );
  }
}

class _ContributionBanner extends StatelessWidget {
  final VoidCallback onTap;

  const _ContributionBanner({required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: AppColors.primaryContainer.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.primary.withValues(alpha: 0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(20),
            ),
            child: const Text(
              '✨ Community Contributor',
              style: TextStyle(
                fontFamily: 'PlusJakartaSans',
                fontWeight: FontWeight.w700,
                fontSize: 11,
                color: AppColors.onPrimaryContainer,
              ),
            ),
          ),
          const SizedBox(height: 12),
          const Text('Quick Contribution', style: AppTextStyles.titleLarge),
          const SizedBox(height: 6),
          const Text(
            'Help preserve local heritage by verifying translations from other learners. Your expertise matters.',
            style: AppTextStyles.bodyMedium,
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: onTap,
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.onPrimaryContainer,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Contribute Now',
              style: TextStyle(
                fontFamily: 'PlusJakartaSans',
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
