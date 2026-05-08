import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../store/phrase_store.dart';
import '../models/flashcard.dart';
import '../utils/app_theme.dart';

class ContributePage extends StatelessWidget {
  const ContributePage({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => PhraseStore(),
      child: const _ContributeView(),
    );
  }
}

class _ContributeView extends StatelessWidget {
  const _ContributeView();

  @override
  Widget build(BuildContext context) {
    final store = context.watch<PhraseStore>();

    if (store.currentView == ContributeView.submit) {
      return const _SubmitPhraseView();
    }
    return const _ValidateFeedView();
  }
}

// ─────────────────────────────────────────────────
// Validate Feed
// ─────────────────────────────────────────────────
class _ValidateFeedView extends StatelessWidget {
  const _ValidateFeedView();

  @override
  Widget build(BuildContext context) {
    final store = context.read<PhraseStore>();

    return Stack(
      children: [
        SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 120),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Progress tracker
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.6),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppColors.onSurface..withValues(alpha: 0.05)),
                ),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Expanded(
                          child: Text(
                            'Keep going! 3 more reviews to unlock rewards',
                            style: AppTextStyles.bodyMedium,
                          ),
                        ),
                        const Text(
                          '7/10',
                          style: TextStyle(
                            fontFamily: 'PlusJakartaSans',
                            fontWeight: FontWeight.w700,
                            fontSize: 13,
                            color: AppColors.primary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: LinearProgressIndicator(
                        value: 0.7,
                        minHeight: 10,
                        backgroundColor: AppColors.onSurface.withValues(alpha: 0.1),
                        valueColor: const AlwaysStoppedAnimation<Color>(
                            AppColors.primaryContainer),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              // Filter chips
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  children: ['All', 'Audio', 'Native Script'].map((label) {
                    final isActive = store.filterMode == label;
                    return Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: GestureDetector(
                        onTap: () => store.setFilter(label),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 20, vertical: 10),
                          decoration: BoxDecoration(
                            color: isActive ? AppColors.primary : Colors.white,
                            border: Border.all(
                              color: isActive
                                  ? AppColors.primary
                                  : AppColors.onSurface.withValues(alpha: 0.1),
                            ),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Text(
                            label,
                            style: TextStyle(
                              fontFamily: 'PlusJakartaSans',
                              fontWeight: FontWeight.w700,
                              fontSize: 12,
                              color: isActive
                                  ? Colors.white
                                  : AppColors.onSurface,
                            ),
                          ),
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ),
              const SizedBox(height: 20),

              // Validation cards
              ...kRegionalFlashcards.map((card) => Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: _ValidationCard(card: card),
                  )),
            ],
          ),
        ),

        // FAB
        Positioned(
          bottom: 100,
          right: 20,
          child: GestureDetector(
            onTap: () => store.showSubmit(),
            child: Container(
              width: 60,
              height: 60,
              decoration: BoxDecoration(
                color: AppColors.primary,
                borderRadius: BorderRadius.circular(18),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.4),
                    blurRadius: 16,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: const Icon(Icons.add_rounded, color: Colors.white, size: 30),
            ),
          ),
        ),
      ],
    );
  }
}

class _ValidationCard extends StatelessWidget {
  final Flashcard card;

  const _ValidationCard({required this.card});

  @override
  Widget build(BuildContext context) {
    return Container(
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
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: AppColors.secondaryContainer,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Text(
                      '${card.dialect ?? 'Javanese'} (${card.toneLabel})',
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w700,
                        fontSize: 10,
                        color: AppColors.onSecondaryContainer,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                  if (card.isVerified) ...[
                    const SizedBox(width: 8),
                    const Icon(Icons.auto_awesome_rounded,
                        size: 14, color: Color(0xFFD97706)),
                    const Text(
                      ' Community Verified',
                      style: TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w700,
                        fontSize: 10,
                        color: Color(0xFFD97706),
                      ),
                    ),
                  ],
                ],
              ),
              const Icon(Icons.flag_outlined,
                  size: 18, color: AppColors.onSurfaceVariant),
            ],
          ),
          const SizedBox(height: 12),
          if (card.script != null)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(20),
              margin: const EdgeInsets.only(bottom: 12),
              decoration: BoxDecoration(
                color: AppColors.practiceCardBackground,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  Text(card.script!,
                      style: const TextStyle(fontSize: 40),
                      textAlign: TextAlign.center),
                  const SizedBox(height: 8),
                  Text(card.front, style: AppTextStyles.headlineMedium,
                      textAlign: TextAlign.center),
                  Text(card.back,
                      style: const TextStyle(
                          fontFamily: 'Lexend',
                          color: AppColors.primary,
                          fontStyle: FontStyle.italic,
                          fontSize: 13),
                      textAlign: TextAlign.center),
                ],
              ),
            )
          else
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(card.front,
                    style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w800,
                        fontSize: 26,
                        color: AppColors.onSurface)),
                Text(card.back,
                    style: const TextStyle(
                        fontFamily: 'Lexend',
                        fontSize: 16,
                        color: AppColors.primary,
                        fontStyle: FontStyle.italic)),
              ],
            ),
          const SizedBox(height: 12),
          Divider(color: AppColors.onSurface..withValues(alpha: 0.05)),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: AppColors.primary..withValues(alpha: 0.05),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.play_circle_fill_rounded,
                        color: AppColors.primary, size: 26),
                  ),
                  
                ],
              ),
              Row(
                children: [
                  _VoteButton(
                      icon: Icons.thumb_up_outlined, count: '12', isPrimary: false),
                  const SizedBox(width: 8),
                  _VoteButton(
                      icon: Icons.thumb_down_outlined, count: '2', isPrimary: false),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 14, vertical: 8),
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Row(
                      children: [
                        Icon(Icons.check_circle_outline_rounded,
                            size: 14, color: Colors.white),
                        SizedBox(width: 4),
                        Text('Confirm',
                            style: TextStyle(
                                fontFamily: 'PlusJakartaSans',
                                fontWeight: FontWeight.w700,
                                fontSize: 12,
                                color: Colors.white)),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _VoteButton extends StatelessWidget {
  final IconData icon;
  final String count;
  final bool isPrimary;

  const _VoteButton(
      {required this.icon, required this.count, required this.isPrimary});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: AppColors.onSurface.withValues(alpha: 0.1)),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: [
          Icon(icon, size: 16, color: AppColors.onSurface),
          const SizedBox(width: 4),
          Text(
            count,
            style: const TextStyle(
              fontFamily: 'PlusJakartaSans',
              fontWeight: FontWeight.w700,
              fontSize: 12,
              color: AppColors.onSurface,
            ),
          ),
        ],
      ),
    );
  }
}

// ─────────────────────────────────────────────────
// Submit Phrase
// ─────────────────────────────────────────────────
class _SubmitPhraseView extends StatelessWidget {
  const _SubmitPhraseView();

  @override
  Widget build(BuildContext context) {
    final store = context.read<PhraseStore>();

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 100),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              IconButton(
                onPressed: store.showValidate,
                icon: const Icon(Icons.arrow_back_rounded, color: AppColors.primary),
              ),
              const SizedBox(width: 8),
              const Text('Submit New Phrase', style: AppTextStyles.headlineMedium),
            ],
          ),
          const SizedBox(height: 24),

          const Text('Language Details', style: AppTextStyles.titleLarge),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            initialValue: 'Javanese (Basa Jawa)',
            items: [
              'Javanese (Basa Jawa)',
              'Sundanese (Basa Sunda)',
              'Balinese (Basa Bali)',
              'Minangkabau',
            ]
                .map((lang) => DropdownMenuItem(value: lang, child: Text(lang)))
                .toList(),
            onChanged: (v) {},
            decoration: const InputDecoration(labelText: 'Select Language'),
          ),
          const SizedBox(height: 12),
          const TextField(
            decoration: InputDecoration(
              labelText: 'Dialect/Region',
              hintText: 'e.g. Yogyakarta, Surakarta',
            ),
          ),
          const SizedBox(height: 16),
          Wrap(
            spacing: 8,
            children: ['Daily Conversation', 'Proverb', 'Formal', 'Greeting']
                .map(
                  (tag) => Chip(
                    label: Text(
                      tag,
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w700,
                        fontSize: 11,
                      ),
                    ),
                    backgroundColor: AppColors.primary..withValues(alpha: 0.05),
                    side: const BorderSide(color: AppColors.primary),
                    labelStyle: const TextStyle(color: AppColors.primary),
                  ),
                )
                .toList(),
          ),
          const SizedBox(height: 24),

          const Text('Regional Phrase (Latin)', style: AppTextStyles.titleMedium),
          const SizedBox(height: 8),
          const TextField(
            decoration: InputDecoration(
                hintText: 'Sugeng enjang, pripun kabaripun?'),
          ),
          const SizedBox(height: 16),
          const Text('Indonesian Translation', style: AppTextStyles.titleMedium),
          const SizedBox(height: 8),
          const TextField(
            decoration: InputDecoration(
                hintText: 'Selamat pagi, bagaimana kabarnya?'),
          ),
          const SizedBox(height: 24),

          // Native script canvas
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Native Script', style: AppTextStyles.titleLarge),
                  Row(
                    children: [
                      TextButton.icon(
                        onPressed: () {},
                        icon: const Icon(Icons.undo_rounded, size: 16),
                        label: const Text('Undo'),
                        style: TextButton.styleFrom(
                            foregroundColor: AppColors.onSurfaceVariant),
                      ),
                      TextButton.icon(
                        onPressed: () {},
                        icon: const Icon(Icons.delete_outline_rounded, size: 16),
                        label: const Text('Clear'),
                        style:
                            TextButton.styleFrom(foregroundColor: Colors.red),
                      ),
                    ],
                  ),
                ],
              ),
              Container(
                width: double.infinity,
                height: 160,
                decoration: BoxDecoration(
                  color: AppColors.practiceCardBackground,
                  border: Border.all(
                    color: AppColors.onSurface..withValues(alpha: 0.15),
                    style: BorderStyle.solid,
                    width: 2,
                  ),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Icon(
                  Icons.edit_outlined,
                  size: 48,
                  color: AppColors.outlineVariant,
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),

          // Voice contribution
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(28),
            decoration: BoxDecoration(
              color: const Color(0xFFEBF1FE),
              borderRadius: BorderRadius.circular(28),
            ),
            child: Column(
              children: [
                const Text(
                  'Voice Contribution',
                  style: TextStyle(
                    fontFamily: 'PlusJakartaSans',
                    fontWeight: FontWeight.w700,
                    fontSize: 20,
                    color: Color(0xFF006A4E),
                  ),
                  textAlign: TextAlign.center,
                ),
                const Text(
                  'Help others with pronunciation',
                  style: AppTextStyles.bodyMedium,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 20),
                Container(
                  width: 88,
                  height: 88,
                  decoration: BoxDecoration(
                    color: const Color(0xFF006A4E),
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF006A4E).withValues(alpha: 0.3),
                        blurRadius: 20,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: const Icon(Icons.mic_rounded,
                      color: Colors.white, size: 36),
                ),
                const SizedBox(height: 16),
                const Text(
                  '00:00',
                  style: TextStyle(
                    fontFamily: 'monospace',
                    fontWeight: FontWeight.w700,
                    fontSize: 28,
                    color: Color(0xFF006A4E),
                  ),
                ),
                const Text(
                  'TAP TO RECORD',
                  style: TextStyle(
                    fontFamily: 'PlusJakartaSans',
                    fontWeight: FontWeight.w700,
                    fontSize: 11,
                    letterSpacing: 2.5,
                    color: AppColors.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 28),

          // Submit button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: store.showValidate,
              icon: const Icon(Icons.send_rounded),
              label: const Text(
                'Submit to Community',
                style: TextStyle(
                  fontFamily: 'PlusJakartaSans',
                  fontWeight: FontWeight.w700,
                  fontSize: 16,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 18),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                elevation: 4,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
