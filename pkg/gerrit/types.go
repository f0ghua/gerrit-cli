package gerrit

type AccountInfo struct {
	AccountID int    `json:"_account_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

func (a *AccountInfo) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}
	if a.Username != "" {
		return a.Username
	}
	if a.Email != "" {
		return a.Email
	}
	return "unknown"
}

type LabelInfo struct {
	Approved     *AccountInfo      `json:"approved,omitempty"`
	Rejected     *AccountInfo      `json:"rejected,omitempty"`
	Recommended  *AccountInfo      `json:"recommended,omitempty"`
	Disliked     *AccountInfo      `json:"disliked,omitempty"`
	All          []ApprovalInfo    `json:"all,omitempty"`
	Values       map[string]string `json:"values,omitempty"`
	DefaultValue int               `json:"default_value"`
}

type ApprovalInfo struct {
	AccountInfo
	Value int    `json:"value"`
	Date  string `json:"date"`
}

type RevisionInfo struct {
	Number  int        `json:"_number"`
	Ref     string     `json:"ref"`
	Created string     `json:"created"`
	Commit  CommitInfo `json:"commit"`
}

type CommitInfo struct {
	Subject   string        `json:"subject"`
	Message   string        `json:"message"`
	Author    GitPersonInfo `json:"author"`
	Committer GitPersonInfo `json:"committer"`
	Parents   []ParentInfo  `json:"parents"`
}

type GitPersonInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

type ParentInfo struct {
	Commit  string `json:"commit"`
	Subject string `json:"subject"`
}

type ChangeInfo struct {
	ID                     string                   `json:"id"`
	Project                string                   `json:"project"`
	Branch                 string                   `json:"branch"`
	Topic                  string                   `json:"topic,omitempty"`
	ChangeID               string                   `json:"change_id"`
	Subject                string                   `json:"subject"`
	Status                 string                   `json:"status"`
	Created                string                   `json:"created"`
	Updated                string                   `json:"updated"`
	Submitted              string                   `json:"submitted,omitempty"`
	Submitter              *AccountInfo             `json:"submitter,omitempty"`
	Owner                  AccountInfo              `json:"owner"`
	Number                 int                      `json:"_number"`
	Labels                 map[string]LabelInfo     `json:"labels,omitempty"`
	Reviewers              map[string][]AccountInfo `json:"reviewers,omitempty"`
	CurrentRevision        string                   `json:"current_revision,omitempty"`
	Revisions              map[string]RevisionInfo  `json:"revisions,omitempty"`
	Mergeable              bool                     `json:"mergeable,omitempty"`
	Insertions             int                      `json:"insertions"`
	Deletions              int                      `json:"deletions"`
	TotalCommentCount      int                      `json:"total_comment_count"`
	UnresolvedCommentCount int                      `json:"unresolved_comment_count"`
	MoreChanges            bool                     `json:"_more_changes,omitempty"`
}

type FileInfo struct {
	Status        string `json:"status"`
	LinesInserted int    `json:"lines_inserted,omitempty"`
	LinesDeleted  int    `json:"lines_deleted,omitempty"`
	SizeDelta     int    `json:"size_delta"`
	Size          int    `json:"size"`
	OldPath       string `json:"old_path,omitempty"`
}

type DiffInfo struct {
	MetaA      DiffFileMetaInfo `json:"meta_a,omitempty"`
	MetaB      DiffFileMetaInfo `json:"meta_b,omitempty"`
	ChangeType string           `json:"change_type"`
	Content    []DiffContent    `json:"content"`
}

type DiffFileMetaInfo struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Lines       int    `json:"lines"`
}

type DiffContent struct {
	A    []string `json:"a,omitempty"`
	B    []string `json:"b,omitempty"`
	AB   []string `json:"ab,omitempty"`
	Skip int      `json:"skip,omitempty"`
}

type CommentInfo struct {
	ID         string        `json:"id"`
	Path       string        `json:"path,omitempty"`
	Line       int           `json:"line,omitempty"`
	Range      *CommentRange `json:"range,omitempty"`
	InReplyTo  string        `json:"in_reply_to,omitempty"`
	Message    string        `json:"message"`
	Updated    string        `json:"updated"`
	Author     AccountInfo   `json:"author"`
	Unresolved bool          `json:"unresolved"`
	PatchSet   int           `json:"patch_set,omitempty"`
}

type CommentRange struct {
	StartLine      int `json:"start_line"`
	StartCharacter int `json:"start_character"`
	EndLine        int `json:"end_line"`
	EndCharacter   int `json:"end_character"`
}

type ChangeMessageInfo struct {
	ID             string      `json:"id"`
	Author         AccountInfo `json:"author"`
	RealAuthor     AccountInfo `json:"real_author,omitempty"`
	Date           string      `json:"date"`
	Message        string      `json:"message"`
	RevisionNumber int         `json:"_revision_number"`
}

type ReviewInput struct {
	Message  string                    `json:"message,omitempty"`
	Labels   map[string]int            `json:"labels,omitempty"`
	Comments map[string][]CommentInput `json:"comments,omitempty"`
}

type CommentInput struct {
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message"`
	Unresolved bool   `json:"unresolved,omitempty"`
}

type DraftInput struct {
	Path       string        `json:"path"`
	Line       int           `json:"line,omitempty"`
	Range      *CommentRange `json:"range,omitempty"`
	InReplyTo  string        `json:"in_reply_to,omitempty"`
	Message    string        `json:"message"`
	Unresolved bool          `json:"unresolved,omitempty"`
}

type ReviewerInput struct {
	Reviewer string `json:"reviewer"`
	State    string `json:"state,omitempty"`
}

type RebaseInput struct {
	Base           string `json:"base,omitempty"`
	AllowConflicts bool   `json:"allow_conflicts,omitempty"`
}

type AbandonInput struct {
	Message string `json:"message,omitempty"`
}

type SubmitInput struct{}

type AddReviewerResult struct {
	Input     string        `json:"input"`
	Reviewers []AccountInfo `json:"reviewers,omitempty"`
	CCs       []AccountInfo `json:"ccs,omitempty"`
	Error     string        `json:"error,omitempty"`
}
