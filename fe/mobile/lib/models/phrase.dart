// Phrase model
class Phrase {
  final String id;
  final String original;
  final String translation;
  final String language;
  final String? dialect;
  final String? tone;
  final String category;
  final bool isVerified;
  final String? script;
  final int upvotes;
  final int downvotes;

  const Phrase({
    required this.id,
    required this.original,
    required this.translation,
    required this.language,
    this.dialect,
    this.tone,
    required this.category,
    required this.isVerified,
    this.script,
    this.upvotes = 0,
    this.downvotes = 0,
  });
}
