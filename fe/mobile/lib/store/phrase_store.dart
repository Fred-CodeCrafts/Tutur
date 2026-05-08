// Phrase list and submission state
import 'package:flutter/foundation.dart';

enum ContributeView { validate, submit }

class PhraseStore extends ChangeNotifier {
  ContributeView _currentView = ContributeView.validate;
  String _selectedLanguage = 'Javanese (Basa Jawa)';
  String _filterMode = 'All';

  ContributeView get currentView => _currentView;
  String get selectedLanguage => _selectedLanguage;
  String get filterMode => _filterMode;

  void showSubmit() {
    _currentView = ContributeView.submit;
    notifyListeners();
  }

  void showValidate() {
    _currentView = ContributeView.validate;
    notifyListeners();
  }

  void setLanguage(String language) {
    _selectedLanguage = language;
    notifyListeners();
  }

  void setFilter(String filter) {
    _filterMode = filter;
    notifyListeners();
  }
}
