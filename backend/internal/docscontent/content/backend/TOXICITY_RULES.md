---
title: "TOXICITY RULES"
summary: "This document explains how to manage and update the rule-based toxicity detection system."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Toxicity Detection Rules Management

This document explains how to manage and update the rule-based toxicity detection system.

## Overview

The toxicity detection system uses a YAML configuration file located at `backend/config/toxicity_rules.yaml` that defines patterns for detecting toxic content.

## Rule Structure

Each rule consists of:
- **Pattern**: A regular expression pattern to match toxic content
- **Category**: The type of toxic content (hate_speech, profanity, harassment, sexual_content, violence, spam)
- **Severity**: The severity level (low, medium, high)
- **Weight**: A numeric weight (0.0-1.0) that affects the toxicity score
- **Description**: A human-readable description of what the rule detects

### Example Rule

```yaml
- pattern: '\b(f[u*@]ck|sh[i1!]t)\b'
  category: profanity
  severity: medium
  weight: 0.5
  description: "Common profanity"
```

## Categories

### Hate Speech
Racial, ethnic, religious slurs and identity-based attacks.
- **Severity**: Typically high
- **Weight**: 0.9-1.0

### Profanity
Common swear words and variations.
- **Severity**: Medium to high
- **Weight**: 0.5-0.7

### Harassment
Personal attacks, threats, and intimidation.
- **Severity**: Medium to high
- **Weight**: 0.6-0.9

### Sexual Content
Explicit sexual language and sexual violence.
- **Severity**: Medium to high
- **Weight**: 0.5-0.95

### Violence
Violent threats and graphic descriptions.
- **Severity**: High
- **Weight**: 0.8-0.9

### Spam
Promotional content and repetitive text.
- **Severity**: Low to medium
- **Weight**: 0.3-0.4

## Pattern Writing Guidelines

### Case Insensitivity
All patterns are automatically case-insensitive (the `(?i)` flag is added automatically).

### Word Boundaries
Use `\b` to match word boundaries:
```yaml
pattern: '\b(word)\b'  # Matches "word" but not "keyword"
```

### Character Alternatives
Use character classes to handle l33tspeak and obfuscation:
```yaml
pattern: 'f[u*@]ck'  # Matches: fuck, f*ck, f@ck
pattern: 'sh[i1!]t'  # Matches: shit, sh1t, sh!t
```

### Go Regex Limitations
**Important**: Go's `regexp` package does NOT support backreferences!

❌ **Wrong** (will cause errors):
```yaml
pattern: '(.)\1{10,}'  # Backreference \1 not supported
```

✅ **Correct** (explicit repetition):
```yaml
pattern: '(a{10,}|b{10,}|c{10,}|...)'  # Explicit character repetition
```

### Optional Groups
Use `()?` for optional parts:
```yaml
pattern: 'k[i1!]ll\s+(your)?self'  # Matches: kill yourself, kill self
```

## Whitelist Management

The whitelist prevents false positives for legitimate words that contain toxic patterns.

### Adding to Whitelist

Add words to the `whitelist` section:
```yaml
whitelist:
  - "scunthorpe"  # Town name containing profanity pattern
  - "assassin"    # Valid word containing "ass"
  - "class"       # Valid word
```

### Whitelist Behavior
- Whitelist checking happens during pattern matching
- Only content where ALL tokens exactly match whitelisted words are exempt from toxicity rules
- If all tokens in the text are whitelisted, the text is treated as non-toxic
- Otherwise, non-whitelisted tokens are still fully evaluated
- Use sparingly to avoid over-whitelisting and missing toxic content

## Text Normalization

Before matching, text is normalized to handle obfuscation:

| Input | Replacement |
|-------|-------------|
| @ | a |
| 4 | a |
| 3 | e |
| 1, ! | i |
| 0 | o |
| $, 5 | s |
| 7, + | t |
| * | u (for censoring) |

Repeated characters are also normalized (e.g., "asssss" becomes "ass").

## Context Awareness

The system reduces false positives by applying multipliers based on context:

- **Quoted text**: 0.5x multiplier (text in quotes is often being reported, not used)
- **Code snippets**: 0.6x multiplier (function names, variable names)
- **URLs**: 0.7x multiplier (sharing links)
- **@mentions**: 0.8x multiplier (quoting someone)
- **Short text**: 0.8x multiplier (less context, might be incomplete)

## Adding New Rules

1. **Identify the pattern**: Determine what text you want to detect
2. **Choose a category**: Select the appropriate category
3. **Set severity and weight**: Based on how serious the content is
4. **Write the regex pattern**: Using Go-compatible regex syntax
5. **Test the pattern**: Verify it matches intended text
6. **Add to config**: Update `toxicity_rules.yaml`
7. **Reload the service**: The config is loaded once at startup

### Example: Adding a New Slur

```yaml
rules:
  - pattern: '\b(newslur|n3wslur|new-slur)\b'
    category: hate_speech
    severity: high
    weight: 1.0
    description: "New slur and common variations"
```

## Testing Rules

After adding or modifying rules, test them:

```bash
cd backend
go test -v -count=1 ./internal/services -run TestToxicityClassifier
```

Create specific test cases for new patterns:

```go
{
    name:        "New slur detection",
    text:        "This contains a newslur",
    expectToxic: true,
    category:    "hate_speech",
}
```

## Performance Considerations

- **Pattern complexity**: Simpler patterns are faster
- **Number of rules**: Each rule is checked for every classification
- **Target**: < 50ms for a 500-character comment
- **Current performance**: ~4μs per comment (well under target)

## Best Practices

### DO:
✅ Start with high-confidence, high-severity patterns  
✅ Use word boundaries to avoid false positives  
✅ Test patterns against known examples  
✅ Document why each rule exists  
✅ Include common obfuscation variations  

### DON'T:
❌ Add overly broad patterns that cause false positives  
❌ Use backreferences (not supported in Go)  
❌ Add rules without testing  
❌ Forget to update whitelist for legitimate uses  
❌ Set weights too high for borderline content  

## Updating Rules in Production

1. **Test locally**: Verify patterns work as expected
2. **Review with team**: Get approval for new patterns
3. **Update config file**: Modify `toxicity_rules.yaml`
4. **Deploy changes**: Push to repository
5. **Restart service**: Rules are loaded at startup
6. **Monitor metrics**: Check for false positive rate changes

## Future Enhancements

- **Database-backed rules**: Allow runtime updates without deployment
- **A/B testing**: Test new rules on a subset of traffic
- **ML integration**: Combine rule-based and ML-based detection
- **Multi-language support**: Patterns for non-English content
- **Admin UI**: Web interface for managing rules

## Troubleshooting

### Pattern Not Matching

1. **Test regex**: Verify the pattern in a Go regex tester
2. **Check normalization**: Text is normalized before matching
3. **Case sensitivity**: Patterns are automatically case-insensitive
4. **Word boundaries**: Ensure `\b` is used correctly

### Too Many False Positives

1. **Add to whitelist**: For legitimate words
2. **Refine pattern**: Make it more specific
3. **Adjust weight**: Lower the weight if severity is too high
4. **Check context**: Context multipliers might not be working

### Pattern Compilation Error

If you see "error parsing regexp":
1. **Check Go regex syntax**: Ensure pattern is valid for Go's `regexp` package
2. **Avoid backreferences**: Use explicit alternatives instead
3. **Escape special characters**: Use `\` before special regex characters
4. **Test in isolation**: Test the pattern with Go's `regexp.MustCompile()`

## Support

For questions or issues with toxicity detection rules:
1. Check this documentation
2. Review existing patterns in `toxicity_rules.yaml`
3. Run tests to verify behavior
4. Consult the team for policy decisions on new content types

---

**Last Updated**: 2026-01-05  
**Version**: 1.0  
**Maintainer**: Backend Team
