import 'package:flutter/material.dart';
import '../models/flashcard.dart';
import '../utils/app_theme.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 100),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Profile header
          Row(
            children: [
              Container(
                width: 88,
                height: 88,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: AppColors.primary.withValues(alpha: 0.2),
                    width: 3,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.1),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: ClipOval(
                  child: Image.network(
                    'https://lh3.googleusercontent.com/aida-public/AB6AXuD3HV_SEgDM3Pxu1g6uK5qfHmbxI3OWzNrqpnFHUti0Evse_ZDOIHSKMrdGSrhvFjbwBoXk3WKiwmBU0wDSRN3IR2-sWWsinAo_b4cTX28xegx5hBS-5y2uzn6JVbth3wMFLDZssv3DCUZGxUsXMWveuJclBnmjPFyTxo8qRodJaIkSRutFJvdcTkl8G-1YemJW5RHV9gtQ0GnMPr1ghpS4s--K9aK0WiXQFZqelnMzsgS4SVgt5hapqvtB3YIMYnf2ntLWJtN6Nn73',
                    width: 88,
                    height: 88,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => const CircleAvatar(
                      radius: 44,
                      backgroundColor: AppColors.primaryContainer,
                      child: Icon(Icons.person_rounded,
                          size: 44, color: Colors.white),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 20),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Linguistic Dev',
                    style: TextStyle(
                      fontFamily: 'PlusJakartaSans',
                      fontWeight: FontWeight.w800,
                      fontSize: 26,
                      color: AppColors.onSurface,
                    ),
                  ),
                  const Text(
                    'Senior Javanese Contributor',
                    style: TextStyle(
                      fontFamily: 'Lexend',
                      fontSize: 13,
                      fontStyle: FontStyle.italic,
                      color: AppColors.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 32),

          const Text('Recent Reviews', style: AppTextStyles.titleLarge),
          Divider(color: AppColors.onSurface..withValues(alpha: 0.05), height: 16),
          const SizedBox(height: 12),

          ...kRegionalFlashcards.map((card) => Padding(
                padding: const EdgeInsets.only(bottom: 16),
                child: _ReviewCard(card: card),
              )),
        ],
      ),
    );
  }
}

class _ReviewCard extends StatelessWidget {
  final Flashcard card;

  const _ReviewCard({required this.card});

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
                    padding: const EdgeInsets.symmetric(
                        horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: AppColors.secondaryContainer,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Text(
                      card.category.toUpperCase(),
                      style: const TextStyle(
                        fontFamily: 'PlusJakartaSans',
                        fontWeight: FontWeight.w700,
                        fontSize: 10,
                        color: AppColors.onSecondaryContainer,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Icon(Icons.auto_awesome_rounded,
                      size: 14, color: Color(0xFFD97706)),
                  const Text(
                    ' Verified',
                    style: TextStyle(
                      fontFamily: 'PlusJakartaSans',
                      fontWeight: FontWeight.w700,
                      fontSize: 10,
                      color: Color(0xFFD97706),
                    ),
                  ),
                ],
              ),
              const Icon(Icons.flag_outlined,
                  size: 18, color: AppColors.onSurfaceVariant),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            card.front,
            style: const TextStyle(
              fontFamily: 'PlusJakartaSans',
              fontWeight: FontWeight.w700,
              fontSize: 18,
              color: AppColors.onSurface,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            card.back,
            style: const TextStyle(
              fontFamily: 'Lexend',
              fontSize: 13,
              color: AppColors.primary,
              fontStyle: FontStyle.italic,
            ),
          ),
          const SizedBox(height: 14),
          Divider(color: AppColors.onSurface..withValues(alpha: 0.05)),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: AppColors.primary..withValues(alpha: 0.05),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.play_circle_fill_rounded,
                        color: AppColors.primary, size: 22),
                  ),
                  const SizedBox(width: 8),
                  const Text(
                    'Audio Contribution Available',
                    style: TextStyle(
                      fontFamily: 'PlusJakartaSans',
                      fontSize: 11,
                      color: AppColors.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
              Row(
                children: [
                  IconButton(
                    onPressed: () {},
                    icon: const Icon(Icons.thumb_up_outlined,
                        size: 18, color: AppColors.onSurfaceVariant),
                  ),
                  IconButton(
                    onPressed: () {},
                    icon: const Icon(Icons.thumb_down_outlined,
                        size: 18, color: AppColors.onSurfaceVariant),
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
