// Flashcard session state and offline cache management
import 'package:flutter/foundation.dart';
import '../models/flashcard.dart';

class FlashcardStore extends ChangeNotifier {
  int _currentIndex = 0;
  bool _isFlipped = false;
  final List<Flashcard> _cards = kRegionalFlashcards;

  int get currentIndex => _currentIndex;
  bool get isFlipped => _isFlipped;
  List<Flashcard> get cards => _cards;
  Flashcard get currentCard => _cards[_currentIndex];
  int get total => _cards.length;

  void flip() {
    _isFlipped = true;
    notifyListeners();
  }

  void next() {
    _isFlipped = false;
    Future.delayed(const Duration(milliseconds: 150), () {
      _currentIndex = (_currentIndex + 1) % _cards.length;
      notifyListeners();
    });
    notifyListeners();
  }

  void reset() {
    _currentIndex = 0;
    _isFlipped = false;
    notifyListeners();
  }
}
