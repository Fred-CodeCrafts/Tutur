// Flashcard model
enum FlashcardStatus { newWord, reviewing, mastered }

enum ToneType { formal, ngoko, krama }

class Flashcard {
  final String id;
  final String front;
  final String back;
  final String category;
  final String? dialect;
  final String example;
  final String rootWord;
  final ToneType? tone;
  final FlashcardStatus status;
  final bool isVerified;
  final String? script;

  const Flashcard({
    required this.id,
    required this.front,
    required this.back,
    required this.category,
    this.dialect,
    required this.example,
    required this.rootWord,
    this.tone,
    required this.status,
    required this.isVerified,
    this.script,
  });

  String get toneLabel {
    switch (tone) {
      case ToneType.formal:
        return 'Formal';
      case ToneType.ngoko:
        return 'Ngoko';
      case ToneType.krama:
        return 'Krama';
      default:
        return 'Ngoko';
    }
  }

  String get statusLabel {
    switch (status) {
      case FlashcardStatus.newWord:
        return 'New';
      case FlashcardStatus.reviewing:
        return 'Reviewing';
      case FlashcardStatus.mastered:
        return 'Mastered';
    }
  }
}

const List<Flashcard> kRegionalFlashcards = [
  Flashcard(
    id: '1',
    front: 'Sugeng Enjang',
    back: 'Selamat Pagi (Good Morning)',
    category: 'Greetings',
    example:
        'Used to show high respect to elders or in formal social settings during early hours.',
    rootWord: 'Enjang',
    tone: ToneType.formal,
    status: FlashcardStatus.reviewing,
    isVerified: true,
  ),
  Flashcard(
    id: '2',
    front: 'Sugeng Rawuh',
    back: 'Selamat Datang (Welcome)',
    category: 'Greetings',
    example:
        'Literally meaning "Safe Arrival". Commonly used at events or welcoming guests.',
    rootWord: 'Rawuh',
    tone: ToneType.ngoko,
    status: FlashcardStatus.newWord,
    isVerified: true,
  ),
  Flashcard(
    id: '3',
    front: 'Om Swastiastu',
    back: 'May God bless you',
    category: 'Blessings',
    dialect: 'Balinese',
    example: 'Common greeting and blessing used in Bali.',
    rootWord: 'Swastiastu',
    status: FlashcardStatus.newWord,
    isVerified: true,
    script: 'ᬒᬁᬲ᭄ᬯᬲ᭄ᬢ᭄ᬬᬲ᭄ᬢᬸ',
  ),
  Flashcard(
    id: '4',
    front: 'Maturnuwun',
    back: 'Terima Kasih (Thank You)',
    category: 'Daily',
    example: 'The standard way to express gratitude in Javanese.',
    rootWord: 'Nuwun',
    tone: ToneType.ngoko,
    status: FlashcardStatus.mastered,
    isVerified: true,
  ),
  Flashcard(
    id: '5',
    front: 'Wilujeng Sumping',
    back: 'Selamat Datang (Welcome)',
    category: 'Greetings',
    dialect: 'Sundanese',
    example: 'Greeting used to welcome people in West Java.',
    rootWord: 'Sumping',
    tone: ToneType.formal,
    status: FlashcardStatus.newWord,
    isVerified: true,
  ),
];
