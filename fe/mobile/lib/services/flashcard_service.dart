// fetch flashcards with filter, conversation scenarios
import '../models/flashcard.dart';

class FlashcardService {
  // Returns local seed data; replace with API call in production
  Future<List<Flashcard>> fetchFlashcards({String? language}) async {
    await Future.delayed(const Duration(milliseconds: 300));

    if (language == null) return kRegionalFlashcards;

    return kRegionalFlashcards
        .where(
          (card) =>
              card.dialect?.toLowerCase() == language.toLowerCase() ||
              (language.toLowerCase() == 'javanese' && card.dialect == null),
        )
        .toList();
  }

  Future<Flashcard> fetchFlashcardById(String id) async {
    await Future.delayed(const Duration(milliseconds: 200));
    return kRegionalFlashcards.firstWhere(
      (card) => card.id == id,
      orElse: () => throw Exception('Flashcard not found: $id'),
    );
  }
}
