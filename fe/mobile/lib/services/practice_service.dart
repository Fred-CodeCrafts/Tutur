// submit phrase practice results
import '../models/flashcard.dart';

enum PracticeResult { know, dontKnow }

class PracticeSession {
  final List<Flashcard> cards;
  int currentIndex;
  final List<PracticeResult> results;
  final int dailyGoal;
  final int completedOffset;

  PracticeSession({
    required this.cards,
    this.currentIndex = 0,
    List<PracticeResult>? results,
    this.dailyGoal = 25,
    this.completedOffset = 12,
  }) : results = results ?? [];

  int get completedCount => completedOffset + currentIndex;
  double get progressPercent => completedCount / dailyGoal;
  bool get isComplete => currentIndex >= cards.length;
}

class PracticeService {
  Future<PracticeSession> startSession({String? language}) async {
    await Future.delayed(const Duration(milliseconds: 200));

    List<Flashcard> cards = kRegionalFlashcards;
    if (language != null) {
      cards = kRegionalFlashcards
          .where(
            (c) =>
                c.dialect?.toLowerCase() == language.toLowerCase() ||
                (language.toLowerCase() == 'javanese' && c.dialect == null),
          )
          .toList();
    }

    return PracticeSession(cards: cards);
  }

  Future<void> recordResult({
    required String cardId,
    required PracticeResult result,
  }) async {
    // POST result to API in production
    await Future.delayed(const Duration(milliseconds: 100));
  }
}
