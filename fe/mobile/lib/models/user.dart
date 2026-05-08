// User model
enum UserRole { learner, contributor }

class User {
  final String id;
  final String name;
  final String email;
  final UserRole role;
  final String? avatarUrl;
  final int level;
  final int streakDays;

  const User({
    required this.id,
    required this.name,
    required this.email,
    required this.role,
    this.avatarUrl,
    this.level = 1,
    this.streakDays = 0,
  });

  String get roleLabel {
    switch (role) {
      case UserRole.learner:
        return 'Learner';
      case UserRole.contributor:
        return 'Contributor';
    }
  }
}
