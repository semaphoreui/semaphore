package db_lib

// Matches `commit_message VARCHAR(100)` in db/sql/migrations.
const commitMessageMaxRunes = 100

// Slice by runes — byte slicing can split a multi-byte UTF-8 sequence.
func truncateCommitMessage(msg string) string {
	runes := []rune(msg)
	if len(runes) > commitMessageMaxRunes {
		return string(runes[:commitMessageMaxRunes])
	}
	return msg
}
