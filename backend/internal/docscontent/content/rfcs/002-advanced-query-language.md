---
title: "002 Advanced Query Language"
summary: "**Status:** Accepted"
tags: ["rfcs"]
area: "rfcs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# RFC 002: Advanced Query Language

**Status:** Accepted  
**Date:** 2025-11-09  
**Authors:** Clipper Team  
**Version:** 1.0.0

## Summary

This RFC defines a human-readable query language for advanced filtering, sorting, and searching of clips, creators, games, and tags in Clipper. The language enables users to construct sophisticated queries using intuitive syntax similar to popular search engines.

## Context

The current search implementation supports basic filtering through structured query parameters (e.g., `?game_id=123&tags=funny&min_votes=10`). While functional, this approach has limitations:

1. **Not user-friendly**: Difficult to construct complex queries in a text search box
2. **Limited composability**: Cannot easily combine multiple filters with boolean logic
3. **Poor discoverability**: Users don't know what filters are available
4. **No negation**: Cannot exclude specific criteria
5. **Verbose**: Requires multiple parameters for simple queries

We need a concise, expressive query language that users can type directly into a search box, similar to:

- Google Search advanced operators (`site:`, `filetype:`, `-`)
- GitHub Search syntax (`is:open`, `language:go`, `stars:>100`)
- Gmail Search filters (`from:`, `has:attachment`, `after:`)

## Goals

1. **Human-readable**: Natural language syntax that's easy to learn and remember
2. **Expressive**: Support complex queries with boolean logic, ranges, and negation
3. **Discoverable**: Auto-complete suggestions for filters and values
4. **Extensible**: Easy to add new filters and operators
5. **Backward compatible**: Co-exist with existing API parameters
6. **URL-safe**: Can be encoded in URLs without special handling

## Non-Goals

1. Full SQL-like query language (too complex)
2. Regular expressions (security risk, performance concern)
3. Arbitrary field queries (maintain control over searchable fields)
4. Query builder UI (out of scope, may come later)

## Query Language Specification

### Basic Syntax

A query consists of:

1. **Free-text search terms**: Words or phrases to search for
2. **Field filters**: Key-value pairs that filter results
3. **Boolean operators**: Combine conditions with AND, OR, NOT
4. **Sorting directives**: Control result ordering

**General Form:**

```
[free text] [filter:value] [filter:value] ...
```

### Supported Filters

#### Clip Filters

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `game:` | string | Filter by game name or ID | `game:valorant` |
| `creator:` | string | Filter by clip creator username | `creator:shroud` |
| `broadcaster:` | string | Filter by channel/broadcaster | `broadcaster:pokimane` |
| `tag:` | string | Filter by tag (can specify multiple) | `tag:funny` |
| `language:` | string | Filter by language code | `language:en` |
| `duration:` | range | Filter by clip duration in seconds | `duration:10..30` |
| `views:` | range | Filter by view count | `views:>1000` |
| `votes:` | range | Filter by vote score | `votes:>=10` |
| `after:` | date | Clips created after date | `after:2025-01-01` |
| `before:` | date | Clips created before date | `before:2025-12-31` |
| `is:` | flag | Boolean flags | `is:featured`, `is:nsfw` |
| `sort:` | order | Sort order | `sort:popular`, `sort:recent` |

#### User Filters (when searching creators)

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `karma:` | range | Filter by karma points | `karma:>100` |
| `role:` | string | Filter by user role | `role:moderator` |

#### Universal Filters

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `type:` | enum | Limit to specific result type | `type:clips`, `type:creators` |
| `sort:` | enum | Sort results | `sort:relevance`, `sort:recent`, `sort:popular` |

### Operators

#### Negation Operator (`-`)

Prefix any filter or term with `-` to exclude results matching that criterion:

```
game:valorant -tag:nsfw
-is:featured
epic play -game:fortnite
```

#### Boolean AND (implicit)

Multiple filters are combined with AND by default:

```
game:valorant tag:clutch votes:>50
# Returns clips that are: Valorant AND tagged clutch AND have >50 votes
```

#### Boolean OR (explicit)

Use `OR` keyword (case-insensitive) between filters:

```
game:valorant OR game:csgo
tag:funny OR tag:fail
```

#### Grouping with Parentheses

Group conditions with parentheses for complex logic:

```
(game:valorant OR game:csgo) tag:clutch
game:minecraft (tag:funny OR tag:creative) -is:nsfw
```

### Value Types

#### String Values

Simple strings (alphanumeric, hyphen, underscore):

```
game:valorant
creator:shroud
```

Quoted strings for spaces or special characters:

```
game:"League of Legends"
creator:"xQc Official"
tag:"top 10"
```

#### Range Values

Numeric ranges using comparison operators:

```
votes:>10        # Greater than 10
votes:>=10       # Greater than or equal to 10
votes:<100       # Less than 100
votes:<=100      # Less than or equal to 100
votes:10..100    # Between 10 and 100 (inclusive)
duration:30..60  # 30 to 60 seconds
```

#### Date Values

ISO 8601 date format (`YYYY-MM-DD`) or relative dates:

```
after:2025-01-01
before:2025-12-31
after:yesterday
after:last-week
after:last-month
```

Supported relative dates:

- `today`
- `yesterday`
- `last-week` (7 days ago)
- `last-month` (30 days ago)
- `last-year` (365 days ago)

#### Boolean Flags

Use `is:` prefix for boolean properties:

```
is:featured     # Featured clips
is:nsfw         # NSFW content
-is:nsfw        # Exclude NSFW content
```

#### Enumeration Values

Specific allowed values:

```
type:clips | type:creators | type:games | type:tags | type:all
sort:relevance | sort:recent | sort:popular
language:en | language:es | language:fr | etc.
```

### Escaping Rules

#### Quotes

Use double quotes to escape special characters and preserve spaces:

```
"epic moment"           # Phrase search
game:"Grand Theft Auto" # Game name with spaces
```

To include a quote within quoted string, escape with backslash:

```
tag:"\"quoted\" text"
```

#### Special Characters

The following characters have special meaning and should be quoted if used literally:

- `:` (filter separator)
- `-` (negation, at start of term)
- `(` `)` (grouping)
- `>` `<` `=` (comparisons)
- `..` (range separator)
- `"` (quotes)
- `\` (escape character)

Examples:

```
"C++"                  # Literal C++
"price:$10"           # Colon as literal character
"what???"             # Multiple question marks
```

#### Whitespace

Multiple spaces are treated as a single space. Leading/trailing whitespace is ignored:

```
game:valorant    tag:funny
# Same as:
game:valorant tag:funny
```

## Query Examples

### Simple Queries

```
# Free-text search
amazing play

# Single filter
game:valorant

# Multiple filters
game:valorant tag:clutch

# Negation
game:valorant -tag:nsfw

# Quoted values
game:"League of Legends"
```

### Date Filtering

```
# Recent clips
after:yesterday

# Date range
after:2025-01-01 before:2025-12-31

# Last week's popular clips
after:last-week votes:>100

# Specific date
after:2025-06-01
```

### Range Filtering

```
# Popular clips
votes:>100

# Specific vote range
votes:50..200

# Short clips only
duration:<30

# View count range
views:1000..10000
```

### Complex Boolean Logic

```
# Valorant or CS:GO clutches
(game:valorant OR game:csgo) tag:clutch

# Funny or fails, but not NSFW
(tag:funny OR tag:fail) -is:nsfw

# Multiple games with tags
(game:valorant OR game:csgo OR game:apex) (tag:clutch OR tag:ace)

# Popular recent clips from specific creator
creator:shroud after:last-month votes:>50 sort:popular
```

### Sorting

```
# Most recent first
sort:recent

# Most popular
sort:popular

# Best match (default)
sort:relevance

# Combined with filters
game:minecraft tag:creative sort:recent
```

### Type Filtering

```
# Only clips
type:clips valorant

# Only creators
type:creators shroud

# Search everything (default)
valorant
```

### Real-world Examples

```
# "Find featured Valorant clutch plays from the last week with at least 50 votes"
game:valorant tag:clutch after:last-week votes:>=50 is:featured

# "Show me funny Minecraft clips but exclude NSFW"
game:minecraft tag:funny -is:nsfw

# "Popular clips from either Valorant or CS:GO, sorted by recent"
(game:valorant OR game:csgo) votes:>100 sort:recent

# "Short, high-quality clips from shroud"
creator:shroud duration:<60 votes:>20 sort:popular

# "Find 'epic' moments in League of Legends or Dota 2"
epic (game:"League of Legends" OR game:"Dota 2")
```

## Formal Grammar (EBNF)

```ebnf
(* Clipper Query Language Grammar v1.0.0 *)

query           = [ term_list ] , [ filter_list ] ;

term_list       = term , { whitespace , term } ;

term            = [ negation ] , ( word | phrase ) ;

filter_list     = filter_expr , { whitespace , filter_expr } ;

filter_expr     = filter | grouped_filter | boolean_expr ;

filter          = [ negation ] , filter_name , ":" , filter_value ;

grouped_filter  = "(" , filter_list , ")" ;

boolean_expr    = filter_expr , whitespace , "OR" , whitespace , filter_expr ;

filter_name     = "game" | "creator" | "broadcaster" | "tag" | "language" 
                | "duration" | "views" | "votes" | "after" | "before" 
                | "is" | "sort" | "type" | "karma" | "role" ;

filter_value    = range_value | date_value | flag_value | string_value ;

range_value     = comparison_op , number 
                | number , ".." , number ;

comparison_op   = ">" | ">=" | "<" | "<=" | "=" ;

date_value      = iso_date | relative_date ;

iso_date        = year , "-" , month , "-" , day ;

relative_date   = "today" | "yesterday" | "last-week" 
                | "last-month" | "last-year" ;

flag_value      = "featured" | "nsfw" ;

string_value    = word | phrase ;

word            = letter , { letter | digit | "-" | "_" } ;

phrase          = '"' , { any_char - '"' | '\"' } , '"' ;

negation        = "-" ;

whitespace      = " " | "\t" | "\n" | "\r" ;

letter          = "a" | "b" | ... | "z" | "A" | "B" | ... | "Z" ;

digit           = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;

number          = digit , { digit } ;

year            = digit , digit , digit , digit ;

month           = "01" | "02" | ... | "12" ;

day             = "01" | "02" | ... | "31" ;

any_char        = ? any Unicode character ? ;
```

## Edge Cases and Validation

### Empty Queries

```
# Empty query - returns default results (most recent)
""

# Only whitespace - treated as empty
"   "
```

### Invalid Filter Names

```
# Unknown filter - treated as free-text search
unknownfilter:value
# Searches for literal text "unknownfilter:value"
```

### Invalid Filter Values

```
# Invalid date format
after:2025/01/01  # Error: Invalid date format, expected YYYY-MM-DD

# Invalid range
votes:100..50     # Error: Invalid range, min > max

# Missing value
game:             # Error: Filter requires a value
```

### Quote Mismatches

```
# Unclosed quote
game:"League of   # Error: Unclosed quote

# Escaped quotes
game:"\"League\"" # Valid: Searches for game named "League"
```

### Special Character Handling

```
# Colon in free text
C++: the language # "C++:" treated as free text (not a valid filter)

# Dash in words
game:counter-strike # Valid: Dash within word
-game:fortnite      # Valid: Negation of filter
- game:fortnite     # Invalid: Space after negation
```

### Case Sensitivity

```
# Filter names: Case-insensitive
GAME:valorant    # Same as game:valorant
Game:Valorant    # Same as game:valorant

# Filter values: Case-insensitive for keywords, case-sensitive for strings
game:Valorant    # Matches "valorant", "Valorant", "VALORANT"
creator:Shroud   # Exact match for username (case-sensitive)
is:FEATURED      # Same as is:featured
sort:RECENT      # Same as sort:recent
```

### OR Operator Precedence

```
# OR has lower precedence than AND (implicit)
game:valorant tag:funny OR tag:fail
# Parsed as: (game:valorant AND tag:funny) OR tag:fail

# Use parentheses for clarity
game:valorant (tag:funny OR tag:fail)
# Parsed as: game:valorant AND (tag:funny OR tag:fail)
```

### Maximum Query Length

- Maximum query length: 1000 characters
- Maximum filters per query: 50
- Maximum nesting depth (parentheses): 10

### Reserved Words

The following words are reserved and cannot be used as free-text search terms without quoting:

- `OR`
- `AND` (reserved for future use)
- `NOT` (reserved for future use)

To search for these literally, use quotes:

```
"OR"  # Searches for the word "OR"
```

## Implementation Notes

### Parser Implementation

The query language should be implemented using a two-phase approach:

1. **Lexical Analysis (Tokenization)**
   - Split query into tokens: words, phrases, filters, operators
   - Handle quoted strings and escape sequences
   - Normalize whitespace

2. **Syntax Analysis (Parsing)**
   - Build abstract syntax tree (AST)
   - Validate filter names and values
   - Handle operator precedence and grouping
   - Generate errors for invalid syntax

### Translation to Backend Queries

The parsed AST should be translated to backend-specific query formats:

1. **OpenSearch/Elasticsearch**: Bool query with must, should, must_not clauses
2. **PostgreSQL**: WHERE clause with AND/OR/NOT conditions
3. **API Parameters**: Structured filter objects

Example translation:

```
Query: game:valorant tag:clutch -is:nsfw votes:>50

OpenSearch Bool Query:
{
  "bool": {
    "must": [
      { "match": { "game_name": "valorant" } },
      { "term": { "tags": "clutch" } },
      { "range": { "vote_score": { "gt": 50 } } }
    ],
    "must_not": [
      { "term": { "is_nsfw": true } }
    ]
  }
}

PostgreSQL WHERE:
WHERE game_name ILIKE '%valorant%'
  AND 'clutch' = ANY(tags)
  AND is_nsfw = false
  AND vote_score > 50
```

### Auto-complete Support

Provide auto-complete suggestions for:

1. **Filter names**: Show available filters as user types
2. **Filter values**: Suggest common values (games, tags, creators)
3. **Recent queries**: Show user's query history

### URL Encoding

Queries can be URL-encoded for sharing:

```
# Original query
game:valorant tag:clutch votes:>50

# URL-encoded
?q=game%3Avalorant%20tag%3Aclutch%20votes%3A%3E50

# Alternative: Base64 encoding for complex queries
?q=Z2FtZTp2YWxvcmFudCB0YWc6Y2x1dGNoIHZvdGVzOj41MA==
```

## Migration Strategy

### Phase 1: Backward Compatibility (Week 1-2)

- Implement parser and translator
- Support both query language and existing API parameters
- If both provided, query language takes precedence

Example:

```
# Old way (still supported)
GET /api/v1/search?game_id=123&tags=funny&min_votes=10

# New way
GET /api/v1/search?q=game:valorant tag:funny votes:>10

# Both provided - query language wins
GET /api/v1/search?q=game:valorant&game_id=123
```

### Phase 2: UI Integration (Week 3-4)

- Add query builder/helper in search box
- Implement auto-complete suggestions
- Add "Advanced Search" UI for non-technical users

### Phase 3: Deprecation (Future)

- Mark structured parameters as deprecated (6 months notice)
- Migrate all clients to query language
- Eventually remove structured parameter support

## Security Considerations

### SQL Injection Prevention

- Never directly interpolate query values into SQL
- Use parameterized queries or ORM-safe methods
- Validate all filter values against allowed patterns

### Resource Exhaustion

- Limit query complexity (max 50 filters, max 10 nesting levels)
- Limit query length (1000 characters)
- Rate limit search API endpoints
- Implement query timeout (5 seconds max)

### Information Disclosure

- Validate filter names against whitelist (no arbitrary field access)
- Filter out sensitive fields from search results
- Respect user permissions and privacy settings

## Testing Strategy

### Unit Tests

1. **Lexer/Tokenizer**:
   - Test tokenization of simple and complex queries
   - Test quote handling and escape sequences
   - Test special character handling

2. **Parser**:
   - Test AST generation for valid queries
   - Test error handling for invalid queries
   - Test operator precedence and grouping

3. **Translator**:
   - Test translation to OpenSearch queries
   - Test translation to PostgreSQL queries
   - Test edge cases and boundary conditions

### Integration Tests

1. **API Tests**:
   - Test query execution against real database
   - Test result correctness for various queries
   - Test performance with large datasets

2. **End-to-End Tests**:
   - Test search UI with query language
   - Test auto-complete functionality
   - Test URL encoding/decoding

### Test Cases

```
# Simple queries
"valorant"
"game:valorant"
"game:valorant tag:clutch"

# Negation
"-game:fortnite"
"game:valorant -tag:nsfw"

# Ranges
"votes:>50"
"votes:10..100"
"duration:<60"

# Dates
"after:2025-01-01"
"after:yesterday"
"after:last-week before:today"

# Boolean logic
"game:valorant OR game:csgo"
"(game:valorant OR game:csgo) tag:clutch"
"(tag:funny OR tag:fail) -is:nsfw"

# Complex queries
"creator:shroud after:last-month votes:>50 sort:popular"
"(game:valorant OR game:csgo) (tag:clutch OR tag:ace) -is:nsfw votes:>=10"

# Edge cases
""  # Empty query
"   "  # Whitespace only
"game:"  # Missing value (error)
"after:invalid-date"  # Invalid date (error)
"votes:100..50"  # Invalid range (error)
"game:\"unclosed  # Unclosed quote (error)
```

## Performance Considerations

### Query Optimization

1. **Filter Order**: Apply most selective filters first
2. **Index Usage**: Ensure all filterable fields are indexed
3. **Caching**: Cache popular queries and auto-complete suggestions
4. **Query Analysis**: Log slow queries for optimization

### Benchmarks

Target performance metrics:

- Parse time: <5ms for typical query
- Translate time: <2ms
- Total query execution: <100ms (p95)

## Documentation

### User-Facing Documentation

Create comprehensive documentation in:

- `docs/user-guide.md`: Add "Advanced Search" section
- `docs/SEARCH.md`: Add query language reference

### API Documentation

Update `docs/API.md` with:

- Query language syntax
- Examples for common use cases
- Error responses for invalid queries

### Developer Documentation

Create `docs/QUERY_LANGUAGE_IMPLEMENTATION.md` with:

- Parser architecture
- AST structure
- Translation logic
- Extension guide (adding new filters)

## Future Enhancements

### Version 1.1 (Future)

- `AND` operator (explicit, currently implicit)
- `NOT` operator (alternative to `-`)
- Field boosting: `game:valorant^2` (boost relevance)
- Fuzzy matching: `game:valorant~` (typo tolerance)
- Wildcards: `game:valo*` (prefix matching)

### Version 2.0 (Future)

- Saved queries/filters
- Query templates
- Collaborative filtering: `similar-to:clip-id`
- Semantic search: `meaning:"epic comeback"`
- Natural language queries: "show me funny valorant clips from last week"

## Success Metrics

### Adoption Metrics

- % of searches using query language (target: >50% in 3 months)
- Number of filters used per query (target: avg 2-3)
- Query language error rate (target: <5%)

### Performance Metrics

- Query parse time (target: <5ms p95)
- Search latency with query language vs. structured params (target: no regression)
- Auto-complete response time (target: <100ms)

### User Satisfaction

- User feedback on query language usability
- Search result relevance (click-through rate)
- Feature requests for new filters/operators

## References

- [Google Search Operators](https://support.google.com/websearch/answer/2466433)
- [GitHub Search Syntax](https://docs.github.com/en/search-github/searching-on-github)
- [Gmail Search Operators](https://support.google.com/mail/answer/7190)
- [Elasticsearch Query DSL](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl.html)
- [EBNF Notation](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form)

## Appendix A: Complete Filter Reference

### Clip Filters

| Filter | Type | Values | Example | Description |
|--------|------|--------|---------|-------------|
| `game:` | string | Game name or ID | `game:valorant` | Filter by game |
| `creator:` | string | Username | `creator:shroud` | Filter by clip creator |
| `broadcaster:` | string | Username | `broadcaster:pokimane` | Filter by channel |
| `tag:` | string | Tag name | `tag:funny` | Filter by tag (multiple allowed) |
| `language:` | string | ISO 639-1 code | `language:en` | Filter by language |
| `duration:` | range | Seconds | `duration:10..30` | Filter by clip length |
| `views:` | range | Integer | `views:>1000` | Filter by view count |
| `votes:` | range | Integer | `votes:>=10` | Filter by vote score |
| `after:` | date | YYYY-MM-DD or relative | `after:2025-01-01` | Clips created after date |
| `before:` | date | YYYY-MM-DD or relative | `before:2025-12-31` | Clips created before date |
| `is:` | flag | `featured`, `nsfw` | `is:featured` | Boolean properties |
| `sort:` | enum | `relevance`, `recent`, `popular` | `sort:popular` | Sort order |

### User/Creator Filters

| Filter | Type | Values | Example | Description |
|--------|------|--------|---------|-------------|
| `karma:` | range | Integer | `karma:>100` | Filter by karma points |
| `role:` | string | Role name | `role:moderator` | Filter by user role |

### Universal Filters

| Filter | Type | Values | Example | Description |
|--------|------|--------|---------|-------------|
| `type:` | enum | `clips`, `creators`, `games`, `tags`, `all` | `type:clips` | Result type |

## Appendix B: Error Codes

| Error Code | Description | Example |
|------------|-------------|---------|
| `QE001` | Invalid filter name | `unknownfilter:value` |
| `QE002` | Missing filter value | `game:` |
| `QE003` | Invalid date format | `after:2025/01/01` |
| `QE004` | Invalid range | `votes:100..50` |
| `QE005` | Unclosed quote | `game:"League of` |
| `QE006` | Invalid comparison operator | `votes:><10` |
| `QE007` | Query too long | (>1000 chars) |
| `QE008` | Too many filters | (>50 filters) |
| `QE009` | Nesting too deep | (>10 levels) |
| `QE010` | Invalid enum value | `sort:invalid` |
| `QE011` | Too many OR clauses | (>20 OR operators) |
| `QE012` | Too many terms in term list | (>100 terms) |

## Conclusion

This query language provides a powerful, user-friendly way to search and filter content in Clipper. It balances expressiveness with simplicity, making it accessible to casual users while supporting advanced power users. The formal grammar ensures consistent parsing and validation, while comprehensive examples guide implementation and testing.

**Decision Date:** 2025-11-09  
**Review Date:** 2026-05-09 (6 months)

---

**Approved By:** Clipper Engineering Team
