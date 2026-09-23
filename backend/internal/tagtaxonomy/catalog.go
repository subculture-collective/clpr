// Package tagtaxonomy describes clpr's tag namespaces: which process creates
// each kind of tag, how it should be labelled, and, for content tags chosen by
// the vision tagger, what evidence the tag definition requires.
package tagtaxonomy

// Evidence is the evidence a content tag definition requires before the
// vision tagger may assign it.
type Evidence string

const (
	// Visible describes a scene, activity, or object that a thumbnail can
	// directly support without guessing what happened before or after it.
	Visible Evidence = "visible"
	// Contextual requires explicit Twitch metadata or an authorized
	// transcript in addition to compatible visual evidence.
	Contextual Evidence = "contextual"
	// Strong describes an outcome or subjective interpretation. A catchy
	// source title or a single ambiguous frame is not sufficient evidence.
	Strong Evidence = "strong"
)

// ContentTag is one entry in the model-facing content vocabulary.
type ContentTag struct {
	Slug        string
	Name        string
	Description string
	Evidence    Evidence
}

// ContentCatalog is the canonical model-facing content vocabulary. Keep these
// slugs aligned with migration 000143. Structural game, language, and duration
// tags remain deterministic and intentionally do not appear here.
var ContentCatalog = []ContentTag{
	// Broad formats and activities that can usually be seen in one frame.
	{"gameplay", "Gameplay", "video game play is visibly on screen", Visible},
	{"facecam", "Facecam", "streamer camera is the primary visible subject", Visible},
	{"vtuber", "VTuber", "a virtual avatar is the primary visible host", Visible},
	{"screen-share", "Screen Share", "a desktop, browser, or application is being shared", Visible},
	{"conversation", "Conversation", "people are visibly talking together", Visible},
	{"interview", "Interview", "host and guest interview setup", Visible},
	{"podcast", "Podcast", "podcast or panel recording setup", Visible},
	{"commentary", "Commentary", "commentator or analysis presentation", Contextual},
	{"reaction", "Reaction", "a clearly visible reaction to content", Visible},
	{"review", "Review", "a product, game, or media review", Contextual},
	{"educational", "Educational", "teaching or explanatory content", Contextual},
	{"tutorial", "Tutorial", "step-by-step instructional content", Contextual},
	{"demonstration", "Demonstration", "a process or technique being demonstrated", Contextual},
	{"irl", "IRL", "real-world people or places rather than gameplay", Visible},
	{"event", "Event", "organized real-world live event or gathering", Visible},
	{"crowd", "Crowd", "a real-world audience or large group", Visible},
	{"stage", "Stage", "a real-world stage or live venue is visible", Visible},
	{"charity", "Charity", "fundraising or charity stream", Contextual},

	// Creative, performance, and real-world activities.
	{"music", "Music", "music is the primary subject", Contextual},
	{"music-performance", "Music Performance", "real-world live musical performance is visible", Visible},
	{"singing", "Singing", "a real person is visibly performing vocals", Contextual},
	{"instrument", "Instrument", "a real musical instrument is being played", Visible},
	{"dance", "Dance", "real-world dance performance or practice", Visible},
	{"creative", "Creative", "real-world creative work is the primary subject", Visible},
	{"art", "Art", "real-world visual art creation or presentation", Visible},
	{"drawing", "Drawing", "real-world drawing, painting, or illustration in progress", Visible},
	{"cosplay", "Cosplay", "a real person wears a character costume", Visible},
	{"cooking", "Cooking", "meal preparation or cooking", Visible},
	{"baking", "Baking", "baking or pastry preparation", Visible},
	{"crafting", "Crafting", "hands-on craft work", Visible},
	{"maker", "Maker", "building or repairing a physical project", Visible},
	{"coding", "Coding", "software development or code is visibly central", Visible},
	{"fitness", "Fitness", "real-world exercise or workout activity", Visible},
	{"sports", "Sports", "real-world athletic activity", Visible},
	{"travel", "Travel", "real-world travel destination or transit experience", Visible},
	{"outdoors", "Outdoors", "real-world outdoor activity or natural setting", Visible},
	{"animals", "Animals", "a real animal is a central visible subject", Visible},
	{"unboxing", "Unboxing", "a product package is being opened", Contextual},
	{"pack-opening", "Pack Opening", "a digital card or loot pack is being opened", Contextual},

	// Gaming format and genre. Game/category metadata is supporting context, not
	// permission to invent an on-screen event.
	{"esports", "Esports", "organized competitive gaming", Contextual},
	{"tournament", "Tournament", "bracketed or organized competition", Contextual},
	{"ranked", "Ranked", "ranked or ladder gameplay", Contextual},
	{"casual-play", "Casual Play", "non-competitive casual gameplay", Contextual},
	{"multiplayer", "Multiplayer", "multiple human players are involved", Contextual},
	{"cooperative", "Cooperative", "players are cooperating toward a goal", Contextual},
	{"pvp", "PvP", "player-versus-player action", Contextual},
	{"pve", "PvE", "players face computer-controlled enemies", Contextual},
	{"boss-fight", "Boss Fight", "a boss encounter is visibly underway", Contextual},
	{"puzzle", "Puzzle", "puzzle solving is central", Contextual},
	{"exploration", "Exploration", "world or level exploration is central", Contextual},
	{"building", "Building", "in-game construction is central", Contextual},
	{"roleplay", "Roleplay", "roleplay performance or roleplay server", Contextual},
	{"simulation", "Simulation", "simulation gameplay", Contextual},
	{"racing-game", "Racing Game", "vehicle racing gameplay", Visible},
	{"fighting-game", "Fighting Game", "one-on-one or arena fighting gameplay", Visible},
	{"sports-game", "Sports Game", "sports video game gameplay", Visible},
	{"strategy-game", "Strategy Game", "strategy game interface or play", Contextual},
	{"horror-game", "Horror Game", "horror game imagery or context", Contextual},
	{"retro-game", "Retro Game", "retro or classic game play", Contextual},
	{"shooter-game", "Shooter Game", "first-person or third-person shooter gameplay", Visible},
	{"moba", "MOBA", "multiplayer online battle arena gameplay", Contextual},
	{"battle-royale", "Battle Royale", "battle royale gameplay", Contextual},
	{"survival-game", "Survival Game", "survival-focused gameplay", Contextual},
	{"sandbox-game", "Sandbox Game", "open-ended sandbox gameplay", Contextual},
	{"card-game", "Card Game", "digital or physical card game play", Visible},
	{"cutscene", "Cutscene", "a narrative game cutscene is visibly playing", Visible},

	// Moments and outcomes. These need evidence beyond an isolated title claim.
	{"ace", "Ace", "one player eliminates the opposing team", Strong},
	{"clutch", "Clutch", "a high-pressure play reverses likely defeat", Strong},
	{"victory", "Victory", "a win or victory is explicitly evidenced", Strong},
	{"defeat", "Defeat", "a loss or defeat is explicitly evidenced", Strong},
	{"comeback", "Comeback", "a recovery from a losing position", Strong},
	{"close-call", "Close Call", "a narrowly avoided failure or defeat", Strong},
	{"elimination", "Elimination", "a player or opponent is eliminated", Strong},
	{"team-wipe", "Team Wipe", "an entire opposing team is eliminated", Strong},
	{"speedrun", "Speedrun", "timed speedrun attempt or route", Contextual},
	{"record", "Record", "a record result is explicitly established", Strong},
	{"personal-best", "Personal Best", "a personal-best result is established", Strong},
	{"boss-kill", "Boss Kill", "a boss is visibly or explicitly defeated", Strong},
	{"jump-scare", "Jump Scare", "a sudden scare and reaction are evidenced", Strong},
	{"trick-shot", "Trick Shot", "an intentionally difficult stylized shot succeeds", Strong},
	{"achievement", "Achievement", "an achievement or milestone is reached", Strong},
	{"discovery", "Discovery", "a new location, item, or fact is discovered", Strong},
	{"challenge", "Challenge", "a named challenge attempt", Contextual},
	{"reveal", "Reveal", "a result, item, or announcement is revealed", Contextual},
	{"celebration", "Celebration", "visible celebration of an evidenced event", Visible},
	{"fail", "Fail", "an attempted action clearly fails", Strong},
	{"bug", "Bug", "a software or game defect is evidenced", Strong},
	{"lucky", "Lucky", "an outcome depends on clearly unusual luck", Strong},
	{"highlight", "Highlight", "a single standout moment", Strong},
	{"highlights", "Highlights", "a compilation or collection of standout moments", Contextual},

	// Tone and response. Prefer literal, visible descriptions over mind-reading.
	{"funny", "Funny", "clearly comedic content, not merely a playful title", Strong},
	{"wholesome", "Wholesome", "clearly warm or uplifting interaction", Strong},
	{"emotional", "Emotional", "strong visible emotion with supporting context", Contextual},
	{"surprise", "Surprise", "clear visible surprise or astonishment", Visible},
	{"rage", "Rage", "clear intense anger, not neutral concentration", Visible},
	{"scary", "Scary", "frightening content or a fear response", Contextual},

	// Retained canonical legacy labels. They remain searchable, but the model is
	// told to use them only when evidence supports their subjective judgment.
	{"insane", "Insane", "an exceptionally extreme moment", Strong},
	{"toxic", "Toxic", "explicitly abusive or hostile conduct", Strong},
	{"epic", "Epic", "an unusually grand or dramatic moment", Strong},
	{"noob", "Beginner", "clearly beginner-oriented play or instruction", Strong},
	{"pro", "Professional", "verified professional-level context", Strong},
}

var contentBySlug = func() map[string]ContentTag {
	bySlug := make(map[string]ContentTag, len(ContentCatalog))
	for _, tag := range ContentCatalog {
		bySlug[tag.Slug] = tag
	}
	return bySlug
}()

// LookupContent returns the catalog entry for a bare content slug such as
// "boss-fight".
func LookupContent(slug string) (ContentTag, bool) {
	tag, ok := contentBySlug[slug]
	return tag, ok
}
