package services

import (
	"fmt"
	"strings"
)

type contentTagEvidence string

const (
	// visibleTag describes a scene, activity, or object that a thumbnail can
	// directly support without guessing what happened before or after it.
	visibleTag contentTagEvidence = "visible"
	// contextualTag requires explicit Twitch metadata or an authorized
	// transcript in addition to compatible visual evidence.
	contextualTag contentTagEvidence = "contextual"
	// strongTag describes an outcome or subjective interpretation. A catchy
	// source title or a single ambiguous frame is not sufficient evidence.
	strongTag contentTagEvidence = "strong"
)

type contentTagDefinition struct {
	Slug        string
	Name        string
	Description string
	Evidence    contentTagEvidence
}

// contentTagCatalog is the canonical model-facing content vocabulary. Keep
// these slugs aligned with migration 000143. Structural game, language, and
// duration tags remain deterministic and intentionally do not appear here.
var contentTagCatalog = []contentTagDefinition{
	// Broad formats and activities that can usually be seen in one frame.
	{"gameplay", "Gameplay", "video game play is visibly on screen", visibleTag},
	{"facecam", "Facecam", "streamer camera is the primary visible subject", visibleTag},
	{"vtuber", "VTuber", "a virtual avatar is the primary visible host", visibleTag},
	{"screen-share", "Screen Share", "a desktop, browser, or application is being shared", visibleTag},
	{"conversation", "Conversation", "people are visibly talking together", visibleTag},
	{"interview", "Interview", "host and guest interview setup", visibleTag},
	{"podcast", "Podcast", "podcast or panel recording setup", visibleTag},
	{"commentary", "Commentary", "commentator or analysis presentation", contextualTag},
	{"reaction", "Reaction", "a clearly visible reaction to content", visibleTag},
	{"review", "Review", "a product, game, or media review", contextualTag},
	{"educational", "Educational", "teaching or explanatory content", contextualTag},
	{"tutorial", "Tutorial", "step-by-step instructional content", contextualTag},
	{"demonstration", "Demonstration", "a process or technique being demonstrated", contextualTag},
	{"irl", "IRL", "real-world people or places rather than gameplay", visibleTag},
	{"event", "Event", "organized real-world live event or gathering", visibleTag},
	{"crowd", "Crowd", "a real-world audience or large group", visibleTag},
	{"stage", "Stage", "a real-world stage or live venue is visible", visibleTag},
	{"charity", "Charity", "fundraising or charity stream", contextualTag},

	// Creative, performance, and real-world activities.
	{"music", "Music", "music is the primary subject", contextualTag},
	{"music-performance", "Music Performance", "real-world live musical performance is visible", visibleTag},
	{"singing", "Singing", "a real person is visibly performing vocals", contextualTag},
	{"instrument", "Instrument", "a real musical instrument is being played", visibleTag},
	{"dance", "Dance", "real-world dance performance or practice", visibleTag},
	{"creative", "Creative", "real-world creative work is the primary subject", visibleTag},
	{"art", "Art", "real-world visual art creation or presentation", visibleTag},
	{"drawing", "Drawing", "real-world drawing, painting, or illustration in progress", visibleTag},
	{"cosplay", "Cosplay", "a real person wears a character costume", visibleTag},
	{"cooking", "Cooking", "meal preparation or cooking", visibleTag},
	{"baking", "Baking", "baking or pastry preparation", visibleTag},
	{"crafting", "Crafting", "hands-on craft work", visibleTag},
	{"maker", "Maker", "building or repairing a physical project", visibleTag},
	{"coding", "Coding", "software development or code is visibly central", visibleTag},
	{"fitness", "Fitness", "real-world exercise or workout activity", visibleTag},
	{"sports", "Sports", "real-world athletic activity", visibleTag},
	{"travel", "Travel", "real-world travel destination or transit experience", visibleTag},
	{"outdoors", "Outdoors", "real-world outdoor activity or natural setting", visibleTag},
	{"animals", "Animals", "a real animal is a central visible subject", visibleTag},
	{"unboxing", "Unboxing", "a product package is being opened", contextualTag},
	{"pack-opening", "Pack Opening", "a digital card or loot pack is being opened", contextualTag},

	// Gaming format and genre. Game/category metadata is supporting context, not
	// permission to invent an on-screen event.
	{"esports", "Esports", "organized competitive gaming", contextualTag},
	{"tournament", "Tournament", "bracketed or organized competition", contextualTag},
	{"ranked", "Ranked", "ranked or ladder gameplay", contextualTag},
	{"casual-play", "Casual Play", "non-competitive casual gameplay", contextualTag},
	{"multiplayer", "Multiplayer", "multiple human players are involved", contextualTag},
	{"cooperative", "Cooperative", "players are cooperating toward a goal", contextualTag},
	{"pvp", "PvP", "player-versus-player action", contextualTag},
	{"pve", "PvE", "players face computer-controlled enemies", contextualTag},
	{"boss-fight", "Boss Fight", "a boss encounter is visibly underway", contextualTag},
	{"puzzle", "Puzzle", "puzzle solving is central", contextualTag},
	{"exploration", "Exploration", "world or level exploration is central", contextualTag},
	{"building", "Building", "in-game construction is central", contextualTag},
	{"roleplay", "Roleplay", "roleplay performance or roleplay server", contextualTag},
	{"simulation", "Simulation", "simulation gameplay", contextualTag},
	{"racing-game", "Racing Game", "vehicle racing gameplay", visibleTag},
	{"fighting-game", "Fighting Game", "one-on-one or arena fighting gameplay", visibleTag},
	{"sports-game", "Sports Game", "sports video game gameplay", visibleTag},
	{"strategy-game", "Strategy Game", "strategy game interface or play", contextualTag},
	{"horror-game", "Horror Game", "horror game imagery or context", contextualTag},
	{"retro-game", "Retro Game", "retro or classic game play", contextualTag},
	{"shooter-game", "Shooter Game", "first-person or third-person shooter gameplay", visibleTag},
	{"moba", "MOBA", "multiplayer online battle arena gameplay", contextualTag},
	{"battle-royale", "Battle Royale", "battle royale gameplay", contextualTag},
	{"survival-game", "Survival Game", "survival-focused gameplay", contextualTag},
	{"sandbox-game", "Sandbox Game", "open-ended sandbox gameplay", contextualTag},
	{"card-game", "Card Game", "digital or physical card game play", visibleTag},
	{"cutscene", "Cutscene", "a narrative game cutscene is visibly playing", visibleTag},

	// Moments and outcomes. These need evidence beyond an isolated title claim.
	{"ace", "Ace", "one player eliminates the opposing team", strongTag},
	{"clutch", "Clutch", "a high-pressure play reverses likely defeat", strongTag},
	{"victory", "Victory", "a win or victory is explicitly evidenced", strongTag},
	{"defeat", "Defeat", "a loss or defeat is explicitly evidenced", strongTag},
	{"comeback", "Comeback", "a recovery from a losing position", strongTag},
	{"close-call", "Close Call", "a narrowly avoided failure or defeat", strongTag},
	{"elimination", "Elimination", "a player or opponent is eliminated", strongTag},
	{"team-wipe", "Team Wipe", "an entire opposing team is eliminated", strongTag},
	{"speedrun", "Speedrun", "timed speedrun attempt or route", contextualTag},
	{"record", "Record", "a record result is explicitly established", strongTag},
	{"personal-best", "Personal Best", "a personal-best result is established", strongTag},
	{"boss-kill", "Boss Kill", "a boss is visibly or explicitly defeated", strongTag},
	{"jump-scare", "Jump Scare", "a sudden scare and reaction are evidenced", strongTag},
	{"trick-shot", "Trick Shot", "an intentionally difficult stylized shot succeeds", strongTag},
	{"achievement", "Achievement", "an achievement or milestone is reached", strongTag},
	{"discovery", "Discovery", "a new location, item, or fact is discovered", strongTag},
	{"challenge", "Challenge", "a named challenge attempt", contextualTag},
	{"reveal", "Reveal", "a result, item, or announcement is revealed", contextualTag},
	{"celebration", "Celebration", "visible celebration of an evidenced event", visibleTag},
	{"fail", "Fail", "an attempted action clearly fails", strongTag},
	{"bug", "Bug", "a software or game defect is evidenced", strongTag},
	{"lucky", "Lucky", "an outcome depends on clearly unusual luck", strongTag},
	{"highlight", "Highlight", "a single standout moment", strongTag},
	{"highlights", "Highlights", "a compilation or collection of standout moments", contextualTag},

	// Tone and response. Prefer literal, visible descriptions over mind-reading.
	{"funny", "Funny", "clearly comedic content, not merely a playful title", strongTag},
	{"wholesome", "Wholesome", "clearly warm or uplifting interaction", strongTag},
	{"emotional", "Emotional", "strong visible emotion with supporting context", contextualTag},
	{"surprise", "Surprise", "clear visible surprise or astonishment", visibleTag},
	{"rage", "Rage", "clear intense anger, not neutral concentration", visibleTag},
	{"scary", "Scary", "frightening content or a fear response", contextualTag},

	// Retained canonical legacy labels. They remain searchable, but the model is
	// told to use them only when evidence supports their subjective judgment.
	{"insane", "Insane", "an exceptionally extreme moment", strongTag},
	{"toxic", "Toxic", "explicitly abusive or hostile conduct", strongTag},
	{"epic", "Epic", "an unusually grand or dramatic moment", strongTag},
	{"noob", "Beginner", "clearly beginner-oriented play or instruction", strongTag},
	{"pro", "Professional", "verified professional-level context", strongTag},
}

var contentTagSlugs = func() []string {
	slugs := make([]string, 0, len(contentTagCatalog))
	for _, tag := range contentTagCatalog {
		slugs = append(slugs, tag.Slug)
	}
	return slugs
}()

func contentTagPromptCatalog() string {
	groups := []struct {
		evidence contentTagEvidence
		heading  string
	}{
		{visibleTag, "VISIBLE: may be selected from direct thumbnail evidence"},
		{contextualTag, "CONTEXTUAL: requires explicit metadata or transcript plus compatible visual evidence"},
		{strongTag, "STRONG: requires explicit outcome evidence; title wording or one ambiguous frame is not enough"},
	}
	var builder strings.Builder
	for groupIndex, group := range groups {
		if groupIndex > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(group.heading)
		builder.WriteString(":\n")
		for _, tag := range contentTagCatalog {
			if tag.Evidence != group.evidence {
				continue
			}
			fmt.Fprintf(&builder, "- %s: %s\n", tag.Slug, tag.Description)
		}
	}
	return strings.TrimSpace(builder.String())
}
