// upvote/downvote phrase, audio, script
enum VoteType { upvote, downvote, confirm }

class VoteService {
  final Map<String, VoteType> _localVotes = {};

  Future<void> vote({
    required String entityId,
    required VoteType voteType,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));
    _localVotes[entityId] = voteType;
    // PATCH to API in production
  }

  VoteType? getLocalVote(String entityId) => _localVotes[entityId];

  Future<void> confirmPhrase(String phraseId) async {
    await vote(entityId: phraseId, voteType: VoteType.confirm);
  }

  Future<void> flagPhrase(String phraseId) async {
    await Future.delayed(const Duration(milliseconds: 200));
    // POST flag to API in production
  }
}
