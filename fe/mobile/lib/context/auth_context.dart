// Auth state provider — current user, token, login/logout
import 'package:flutter/foundation.dart';

class AuthContext extends ChangeNotifier {
  bool _isLoggedIn = false;
  String? _userEmail;
  String? _userName;

  bool get isLoggedIn => _isLoggedIn;
  String? get userEmail => _userEmail;
  String? get userName => _userName;

  void login({required String email, String? name}) {
    _isLoggedIn = true;
    _userEmail = email;
    _userName = name ?? 'Linguistic Dev';
    notifyListeners();
  }

  void logout() {
    _isLoggedIn = false;
    _userEmail = null;
    _userName = null;
    notifyListeners();
  }
}
