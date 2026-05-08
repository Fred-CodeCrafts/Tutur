import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../store/flashcard_store.dart';
import '../utils/app_theme.dart';

class PracticePage extends StatelessWidget {
  const PracticePage({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => FlashcardStore(),
      child: const _PracticeView(),
    );
  }
}

class _PracticeView extends StatelessWidget {
  const _PracticeView();

  @override
  Widget build(BuildContext context) {
    final store = context.watch<FlashcardStore>();
    final card = store.currentCard;
    final progress = (store.currentIndex + 12) / 25;

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 100),
      child: Column(
        children: [
          // Progress bar
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Daily Goal', style: AppTextStyles.bodyMedium),
                  Text(
                    '${store.currentIndex + 12}/25 completed',
                    style: const TextStyle(
                      fontFamily: 'PlusJakartaSans',
                      fontWeight: FontWeight.w700,
                      fontSize: 12,
                      color: AppColors.primary,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: LinearProgressIndicator(
                  value: progress,
                  minHeight: 10,
                  backgroundColor: AppColors.onSurface.withValues(alpha: 0.1),
                  valueColor:
                      const AlwaysStoppedAnimation<Color>(AppColors.primaryContainer),
                ),
              ),
            ],
          ),
          const SizedBox(height: 28),

          // Question prompt
          const Text(
            'What does this mean?',
            style: TextStyle(
              fontFamily: 'PlusJakartaSans',
              fontWeight: FontWeight.w700,
              fontSize: 16,
              color: AppColors.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 16),

          // Flashcard front
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(vertical: 40, horizontal: 24),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border.all(color: AppColors.onSurface.withValues(alpha: 0.1)),
              borderRadius: BorderRadius.circular(24),
              boxShadow: [
                BoxShadow(
                  color: AppColors.onSurface.withValues(alpha: 0.1),
                  blurRadius: 0,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Stack(
              clipBehavior: Clip.none,
              children: [
                Column(
                  children: [
                    if (card.script != null) ...[
                      Text(
                        card.script!,
                        style: const TextStyle(
                          fontSize: 36,
                          color: AppColors.onSurface,
                          fontStyle: FontStyle.normal,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 12),
                    ],
                    Text(
                      card.front,
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w800,
                        fontSize: 36,
                        color: AppColors.primary,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 6),
                    Text(
                      '${card.dialect ?? 'Javanese'} • ${card.toneLabel}',
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w600,
                        fontSize: 11,
                        letterSpacing: 1.5,
                        color: AppColors.secondary,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
                Positioned(
                  top: -28,
                  right: -12,
                  child: Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
                    decoration: BoxDecoration(
                      color: AppColors.tertiaryContainer,
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: const Text(
                      'New Word',
                      style: TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w700,
                        fontSize: 11,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Action area
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 200),
            child: store.isFlipped
                ? _AnswerArea(store: store)
                : SizedBox(
                    key: const ValueKey('show_answer'),
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: store.flip,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 20),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(18),
                        ),
                        elevation: 4,
                        shadowColor: AppColors.primary.withValues(alpha: 0.3),
                      ),
                      child: const Text(
                        'Show Answer',
                        style: TextStyle(
                          fontFamily: 'PlusJakartaSans',
                          fontWeight: FontWeight.w700,
                          fontSize: 18,
                        ),
                      ),
                    ),
                  ),
          ),
        ],
      ),
    );
  }
}

class _AnswerArea extends StatelessWidget {
  final FlashcardStore store;

  const _AnswerArea({required this.store});

  @override
  Widget build(BuildContext context) {
    final card = store.currentCard;

    return Column(
      key: const ValueKey('answer'),
      children: [
        // Answer card
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: AppColors.practiceCardBackground,
            border: Border.all(color: AppColors.onSurface..withValues(alpha: 0.05)),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          card.back,
                          style: const TextStyle(
                            fontFamily: 'PlusJakartaSans',
                            fontWeight: FontWeight.w700,
                            fontSize: 22,
                            color: AppColors.primary,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(card.example, style: AppTextStyles.bodyMedium),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Container(
                    width: 52,
                    height: 52,
                    decoration: BoxDecoration(
                      color: AppColors.primaryContainer,
                      borderRadius: BorderRadius.circular(26),
                    ),
                    child:
                        const Icon(Icons.volume_up_rounded, color: Colors.white, size: 24),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Divider(color: AppColors.onSurface.withValues(alpha: 0.1)),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Root Word', style: AppTextStyles.labelSmall),
                        const SizedBox(height: 2),
                        Text(
                          card.rootWord,
                          style: const TextStyle(
                            fontFamily: 'Lexend',
                            fontWeight: FontWeight.w600,
                            fontSize: 13,
                            color: AppColors.onSurface,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Status', style: AppTextStyles.labelSmall),
                        const SizedBox(height: 4),
                        Row(
                          children: const [
                            Icon(Icons.check_circle_rounded,
                                color: Color(0xFF059669), size: 14),
                            SizedBox(width: 4),
                            Text(
                              'Verified',
                              style: TextStyle(
                                fontFamily: 'PlusJakartaSans',
                                fontWeight: FontWeight.w700,
                                fontSize: 12,
                                color: Color(0xFF059669),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),

        // Assessment buttons
        Row(
          children: [
            Expanded(
              child: _AssessButton(
                icon: Icons.check_circle_outline_rounded,
                label: 'Know',
                isPrimary: true,
                onTap: store.next,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _AssessButton(
                icon: Icons.replay_rounded,
                label: "Don't Know",
                isPrimary: false,
                onTap: store.next,
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _AssessButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool isPrimary;
  final VoidCallback onTap;

  const _AssessButton({
    required this.icon,
    required this.label,
    required this.isPrimary,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 24),
        decoration: BoxDecoration(
          color: isPrimary
              ? AppColors.primaryContainer.withValues(alpha: 0.1)
              : Colors.white,
          border: Border.all(
            color: isPrimary
                ? AppColors.primaryContainer
                : AppColors.onSurface..withValues(alpha: 0.05),
            width: 2,
          ),
          borderRadius: BorderRadius.circular(18),
        ),
        child: Column(
          children: [
            Icon(
              icon,
              size: 32,
              color: isPrimary ? AppColors.primary : AppColors.onSurface,
            ),
            const SizedBox(height: 8),
            Text(
              label,
              style: TextStyle(
                fontFamily: 'PlusJakartaSans',
                fontWeight: FontWeight.w700,
                fontSize: 14,
                color: isPrimary ? AppColors.primary : AppColors.onSurface,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
