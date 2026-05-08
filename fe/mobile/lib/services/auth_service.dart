// login, register, upgrade role, logout
import '../utils/secure_storage.dart';

class AuthService {
  // Simulated login — replace with real API call in production
  Future<Map<String, String>> login({
    required String email,
    required String password,
  }) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 600));

    // Placeholder validation
    if (email.isEmpty || password.isEmpty) {
      throw Exception('Email and password are required.');
    }

    const fakeToken = 'fake-jwt-token-12345';
    const fakeName = 'Linguistic Dev';

    await SecureStorage.saveToken(fakeToken);
    await SecureStorage.saveUserEmail(email);
    await SecureStorage.saveUserName(fakeName);

    return {'token': fakeToken, 'name': fakeName, 'email': email};
  }

  Future<void> register({
    required String name,
    required String email,
    required String password,
    required String role,
  }) async {
    await Future.delayed(const Duration(milliseconds: 600));

    if (name.isEmpty || email.isEmpty || password.isEmpty) {
      throw Exception('All fields are required.');
    }

    const fakeToken = 'fake-jwt-token-register-67890';
    await SecureStorage.saveToken(fakeToken);
    await SecureStorage.saveUserEmail(email);
    await SecureStorage.saveUserName(name);
  }

  Future<void> logout() async {
    await SecureStorage.clearAll();
  }

  Future<bool> isAuthenticated() async {
    final token = await SecureStorage.getToken();
    return token != null && token.isNotEmpty;
  }
}
