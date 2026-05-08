package commitstyle

// CommitType defines different types of commits
type CommitType struct {
	Type        string
	Emoji       string
	Description string
}

// CommitStyle defines a commit message style
type CommitStyle struct {
	Name        string
	Description string
	Types       []CommitType
	Examples    []string
}

// ConventionalTypes defines the types for conventional commits
var ConventionalTypes = []CommitType{
	{
		Type:        "feat",
		Emoji:       "✨",
		Description: "A new feature",
	},
	{
		Type:        "fix",
		Emoji:       "🐛",
		Description: "A bug fix",
	},
	{
		Type:        "docs",
		Emoji:       "📝",
		Description: "Documentation changes",
	},
	{
		Type:        "style",
		Emoji:       "💅",
		Description: "Code style changes (formatting, missing semicolons, etc)",
	},
	{
		Type:        "refactor",
		Emoji:       "♻️",
		Description: "Code refactoring without changing functionality",
	},
	{
		Type:        "perf",
		Emoji:       "⚡",
		Description: "Performance improvements",
	},
	{
		Type:        "test",
		Emoji:       "✅",
		Description: "Adding or updating tests",
	},
	{
		Type:        "chore",
		Emoji:       "🔧",
		Description: "Maintenance tasks or dependencies",
	},
	{
		Type:        "ci",
		Emoji:       "🔄",
		Description: "CI/CD configuration changes",
	},
	{
		Type:        "build",
		Emoji:       "🏗️",
		Description: "Build system or dependency changes",
	},
}

// SimpleTypes defines the types for simple commits
var SimpleTypes = []CommitType{
	{
		Type:        "update",
		Emoji:       "📦",
		Description: "Update something",
	},
	{
		Type:        "fix",
		Emoji:       "🐛",
		Description: "Fix a bug",
	},
	{
		Type:        "add",
		Emoji:       "➕",
		Description: "Add a new feature",
	},
	{
		Type:        "remove",
		Emoji:       "➖",
		Description: "Remove something",
	},
	{
		Type:        "improve",
		Emoji:       "✨",
		Description: "Improve something",
	},
}

// EmojiTypes defines the types for emoji-based commits
var EmojiTypes = []CommitType{
	{
		Type:        "✨",
		Emoji:       "✨",
		Description: "New feature",
	},
	{
		Type:        "🐛",
		Emoji:       "🐛",
		Description: "Bug fix",
	},
	{
		Type:        "📝",
		Emoji:       "📝",
		Description: "Documentation",
	},
	{
		Type:        "💅",
		Emoji:       "💅",
		Description: "Style changes",
	},
	{
		Type:        "♻️",
		Emoji:       "♻️",
		Description: "Refactoring",
	},
	{
		Type:        "⚡",
		Emoji:       "⚡",
		Description: "Performance",
	},
	{
		Type:        "✅",
		Emoji:       "✅",
		Description: "Tests",
	},
	{
		Type:        "🔧",
		Emoji:       "🔧",
		Description: "Chore",
	},
}

// Styles defines all available commit styles
var Styles = map[string]CommitStyle{
	"conventional": {
		Name:        "conventional",
		Description: "Conventional Commits format with type, scope, and description",
		Types:       ConventionalTypes,
		Examples: []string{
			"feat(api): add user authentication endpoint",
			"fix(database): resolve connection pool exhaustion",
			"docs(readme): update installation instructions",
			"refactor(utils): simplify error handling",
		},
	},
	"simple": {
		Name:        "simple",
		Description: "Simple, readable commit messages without strict format",
		Types:       SimpleTypes,
		Examples: []string{
			"Add user authentication",
			"Fix database connection issue",
			"Update README with new instructions",
			"Improve error handling in utils",
		},
	},
	"emoji": {
		Name:        "emoji",
		Description: "Emoji-based commit messages for visual clarity",
		Types:       EmojiTypes,
		Examples: []string{
			"✨ Add user authentication endpoint",
			"🐛 Fix database connection pool issue",
			"📝 Update installation instructions",
			"♻️ Simplify error handling in utils",
		},
	},
}

// GetStyle retrieves a commit style by name
func GetStyle(name string) *CommitStyle {
	if style, ok := Styles[name];ok {
		return &style
	}  
	style := Styles["simple"] 
	return &style
}

// GetStyleGuide returns a formatted guide for the given style
func GetStyleGuide(styleName string) string {
	style := GetStyle(styleName)
	guide := "Commit style guide:\n"

	switch styleName {
	case "conventional":
		guide += "- conventional: [type](scope): description\n"
		guide += "  where [type] is one of: feat, fix, docs, style, refactor, perf, test, chore, ci, build\n"
	case "simple":
		guide += "- simple: short, descriptive message of what changed\n"
	case "emoji":
		guide += "- emoji: [emoji] description\n"
		for _, t := range style.Types {
			guide += "  " + t.Emoji + " " + t.Description + "\n"
		}
	}

	return guide
}

// GetTypesString returns a formatted string of types for the style
func GetTypesString(styleName string) string {
	style := GetStyle(styleName)
	if styleName == "conventional" {
		types := ""
		for i, t := range style.Types {
			if i > 0 {
				types += ", "
			}
			types += t.Type
		}
		return "Types: " + types
	}
	return ""
}


// how the user style is coming to this getstyle method in this file
// 