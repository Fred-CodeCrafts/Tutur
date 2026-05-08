// submit phrase, list my phrases, list pending for voting
import '../models/phrase.dart';

class PhraseService {
  // Simulated phrase list — replace with real API call in production
  Future<List<Phrase>> fetchPhrases({String? filter}) async {
    await Future.delayed(const Duration(milliseconds: 300));

    final phrases = [
      const Phrase(
        id: 'p1',
        original: 'Sugeng Enjang',
        translation: 'Selamat Pagi (Good Morning)',
        language: 'Javanese',
        tone: 'Formal',
        category: 'Greetings',
        isVerified: true,
        upvotes: 12,
        downvotes: 2,
      ),
      const Phrase(
        id: 'p2',
        original: 'Om Swastiastu',
        translation: 'May God bless you',
        language: 'Balinese',
        dialect: 'Balinese',
        category: 'Blessings',
        isVerified: true,
        script: 'ᬒᬁᬲ᭄ᬯᬲ᭄ᬢ᭄ᬬᬲ᭄ᬢᬸ',
        upvotes: 20,
        downvotes: 1,
      ),
    ];

    if (filter == null || filter == 'All') return phrases;
    return phrases
        .where((p) => p.category.toLowerCase() == filter.toLowerCase())
        .toList();
  }

  Future<void> submitPhrase({
    required String language,
    required String dialect,
    required String phraseText,
    required String translation,
    String? script,
    String? audioPath,
    required String category,
  }) async {
    await Future.delayed(const Duration(milliseconds: 600));
    // POST to API here in production
  }

  Future<void> votePhrase({
    required String phraseId,
    required bool isUpvote,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));
    // PATCH to API here in production
  }
}
