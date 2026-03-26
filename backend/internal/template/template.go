package template

// TemplateList defines a task list with its groups and tasks
type TemplateList struct {
	Name   string
	Color  string
	Groups []TemplateGroup
}

type TemplateGroup struct {
	Name  string
	Tasks []TemplateTask
}

type TemplateTask struct {
	Name     string
	Subtasks []string
	// DaysOffset is the number of days relative to release date.
	// Negative = before release, 0 = release day, positive = after release.
	// nil means no auto-date.
	DaysOffset *int
}

func days(n int) *int { return &n }

// Valid template types
const (
	TypeSingle        = "single"
	TypeSoundCloudFlip = "soundcloud_flip"
	TypeLPEP          = "lp_ep"
)

// ValidTemplateType returns true if the given type is a known template.
func ValidTemplateType(t string) bool {
	switch t {
	case TypeSingle, TypeSoundCloudFlip, TypeLPEP:
		return true
	}
	return false
}

// GetTemplate returns the task lists for the given template type.
// Falls back to Single if type is unknown.
func GetTemplate(templateType string) []TemplateList {
	switch templateType {
	case TypeSoundCloudFlip:
		return SoundCloudFlipTemplate()
	case TypeLPEP:
		return LPEPTemplate()
	default:
		return DefaultTemplate()
	}
}

// DefaultScheduleWeeks returns the default schedule duration for the template type.
func DefaultScheduleWeeks(templateType string) int {
	switch templateType {
	case TypeSoundCloudFlip:
		return 4
	case TypeLPEP:
		return 8
	default:
		return 8
	}
}

// DefaultTemplate returns the 5 pre-loaded task lists for a Single release campaign.
// Tasks have DaysOffset relative to the release date (8-week campaign timeline).
func DefaultTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Audio Production",
					Tasks: []TemplateTask{
						{Name: "Final mixdown + master (WAV + MP3)", DaysOffset: days(-56)},
						{Name: "Extended mix", DaysOffset: days(-56)},
						{Name: "DJ-friendly intro/outro edit", DaysOffset: days(-49)},
						{Name: "Stems pack for remixers", DaysOffset: days(-49)},
						{Name: "Sync-ready edits (30s, 60s, instrumental)", DaysOffset: days(-42)},
						{Name: "Assign ISRC code to each version", DaysOffset: days(-54)},
						{Name: "Obtain UPC/EAN barcode", DaysOffset: days(-54)},
						{Name: "Lock metadata sheet (BPM, key, mood tags, subgenre, credits)", DaysOffset: days(-54)},
					},
				},
				{
					Name: "Visual Assets",
					Tasks: []TemplateTask{
						{Name: "Create asset folder", DaysOffset: days(-56)},
						{Name: "Final artwork (3000x3000)", DaysOffset: days(-42)},
						{Name: "Alternate artwork crops (square, banner, story-sized)", DaysOffset: days(-35)},
						{Name: "Spotify Canvas (3-8s looping video)", DaysOffset: days(-28)},
						{Name: "Visualizer video for YouTube", DaysOffset: days(-14)},
						{Name: "Music video (if applicable)", DaysOffset: days(-7)},
						{Name: "Artist press photos", DaysOffset: days(-42)},
						{Name: "Update EPK (bio, top tracks, press quotes, notable shows)", DaysOffset: days(-42)},
						{Name: "Promo content folder (clips, BTS, etc.)", DaysOffset: days(-35)},
					},
				},
				{
					Name: "Content Production",
					Tasks: []TemplateTask{
						{Name: "Studio session / production walkthrough clips", DaysOffset: days(-28)},
						{Name: "\"How I made this sound\" breakdown", DaysOffset: days(-21)},
						{Name: "DJ booth footage playing the track", DaysOffset: days(-14)},
						{Name: "Audiogram snippets (waveform + clip for stories)", DaysOffset: days(-21)},
						{Name: "Countdown assets (7 days, 3 days, 1 day, tonight, now)", DaysOffset: days(-10)},
						{Name: "Mood/aesthetic video matching the track's vibe", DaysOffset: days(-14)},
					},
				},
			},
		},
		{
			Name:  "Campaign Tasklist",
			Color: "#22C55E",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release (8+ weeks out)",
					Tasks: []TemplateTask{
						{Name: "Campaign setup + timeline", DaysOffset: days(-56)},
						{Name: "Contracts / splits / publishing", DaysOffset: days(-56)},
						{Name: "Create content plan", DaysOffset: days(-56)},
						{Name: "Commission artwork / visual assets", DaysOffset: days(-56)},
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits & metadata", "Set release date", "Upload artwork", "Select stores & territories", "Tag BPM, key, subgenre accurately"}, DaysOffset: days(-56)},
						{Name: "Set Beatport Exclusive window (2, 4, or 8 weeks)", DaysOffset: days(-56)},
						{Name: "Enable Beatport Pre-Order", DaysOffset: days(-56)},
						{Name: "Submit to Beatport editorial with track description", DaysOffset: days(-49)},
					},
				},
				{
					Name: "Pre-Release (4 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Submit to Spotify editorial via Spotify for Artists", DaysOffset: days(-28)},
						{Name: "Create pre-save link", DaysOffset: days(-28)},
						{Name: "Create smart link (Soundplate / Linkfire / Toneden)", DaysOffset: days(-28)},
						{Name: "Pre-save campaign / contest", DaysOffset: days(-28)},
						{Name: "Create or update Linktree", DaysOffset: days(-28)},
						{Name: "Write out hashtags", DaysOffset: days(-21)},
						{Name: "Update website - coming soon", DaysOffset: days(-21)},
						{Name: "Update social media bios - coming soon", DaysOffset: days(-21)},
						{Name: "Update Spotify / Apple Music artist profiles", DaysOffset: days(-21)},
					},
				},
				{
					Name: "Pre-Release (2-3 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Send track to DJ promo pools", Subtasks: []string{"Reaktion", "ZIPDJ", "PromoPush", "Music Worx"}, DaysOffset: days(-18)},
						{Name: "Send private links to target DJs (20-50 in your subgenre)", DaysOffset: days(-18)},
						{Name: "Send to radio promo contacts", DaysOffset: days(-18)},
						{Name: "Collect and log DJ feedback / support quotes", DaysOffset: days(-10)},
						{Name: "Send pre-save email to mailing list", DaysOffset: days(-14)},
					},
				},
				{
					Name: "Release Day",
					Tasks: []TemplateTask{
						{Name: "Coordinated post across all platforms", DaysOffset: days(0)},
						{Name: "Update all bios and link-in-bios to smart link", DaysOffset: days(0)},
						{Name: "Upload to SoundCloud", DaysOffset: days(0)},
						{Name: "Add to all label-owned Spotify playlists", DaysOffset: days(0)},
						{Name: "Push fans to Beatport purchase link", DaysOffset: days(0)},
						{Name: "Send \"out now\" email to mailing list", DaysOffset: days(0)},
						{Name: "Post on Bandcamp (time around Bandcamp Friday if possible)", DaysOffset: days(0)},
					},
				},
				{
					Name: "Post-Release (weeks 1-4)",
					Tasks: []TemplateTask{
						{Name: "Monitor Spotify for Artists daily", DaysOffset: days(1)},
						{Name: "Monitor Beatport chart position and sales", DaysOffset: days(1)},
						{Name: "Track SoundCloud plays, reposts, comments", DaysOffset: days(3)},
						{Name: "Share press coverage on socials", DaysOffset: days(3)},
						{Name: "Repost DJ support and fan reactions", DaysOffset: days(5)},
						{Name: "Respond to all fan comments and messages", DaysOffset: days(1)},
						{Name: "Compile campaign performance report", DaysOffset: days(21)},
						{Name: "Campaign retrospective: what worked, what to improve", DaysOffset: days(28)},
						{Name: "Archive campaign assets for future reference", DaysOffset: days(28)},
					},
				},
			},
		},
		{
			Name:  "Social Media Content",
			Color: "#A78BFA",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release Content",
					Tasks: []TemplateTask{
						{Name: "Teaser clip (15-30s preview)", DaysOffset: days(-21)},
						{Name: "Artwork reveal post", DaysOffset: days(-18)},
						{Name: "Banner - release date + pre-save", DaysOffset: days(-14)},
						{Name: "Story-sized artwork (with release date)", DaysOffset: days(-14)},
						{Name: "Pre-save post", DaysOffset: days(-14)},
						{Name: "Studio session / production breakdown clip", DaysOffset: days(-10)},
						{Name: "Behind the track post (inspiration, story)", DaysOffset: days(-7)},
						{Name: "\"Midnight\" / \"Tomorrow\" countdown post", DaysOffset: days(-1)},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now", DaysOffset: days(0)},
						{Name: "Track clip - out now", DaysOffset: days(0)},
						{Name: "Artwork edit - out now", DaysOffset: days(1)},
						{Name: "Story-sized out now graphic", DaysOffset: days(1)},
						{Name: "Visualizer clip for socials", DaysOffset: days(2)},
						{Name: "DJ set clip featuring the track", DaysOffset: days(3)},
						{Name: "Studio breakdown / production tips reel", DaysOffset: days(5)},
						{Name: "Credits / collab shoutout post", DaysOffset: days(4)},
						{Name: "Fan reaction / UGC repost", DaysOffset: days(7)},
					},
				},
				{
					Name: "TikTok / Reels / Shorts",
					Tasks: []TemplateTask{
						{Name: "TikTok with track (10-15s, hook-first)", DaysOffset: days(-7)},
						{Name: "Production breakdown reel (DAW screen recording)", DaysOffset: days(-5)},
						{Name: "\"How I made this sound\" short", DaysOffset: days(-3)},
						{Name: "DJ booth POV playing the track", DaysOffset: days(1)},
						{Name: "Before/after sound design clip", DaysOffset: days(3)},
						{Name: "Crowd reaction clip (if available)", DaysOffset: days(7)},
						{Name: "Boost top-performing TikTok via Spark Ads (if organic traction)", DaysOffset: days(7)},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "DJ mix snippet featuring the track", DaysOffset: days(5)},
						{Name: "Live set / festival clip with track", DaysOffset: days(10)},
						{Name: "Remix teasers (if applicable)", DaysOffset: days(14)},
						{Name: "Playlist placement screenshot posts", DaysOffset: days(7)},
						{Name: "Milestone celebration posts (streams, charts)", DaysOffset: days(14)},
						{Name: "VIP / remix / bootleg version tease", DaysOffset: days(21)},
						{Name: "Meme / relatable producer content"},
					},
				},
			},
		},
		{
			Name:  "PR & Outreach",
			Color: "#EAB308",
			Groups: []TemplateGroup{
				{
					Name: "Press & Blog Outreach",
					Tasks: []TemplateTask{
						{Name: "Write press release", DaysOffset: days(-42)},
						{Name: "Build media outreach list (30-50 outlets)", DaysOffset: days(-42)},
						{Name: "Pitch exclusive premiere to top-choice outlet", DaysOffset: days(-28)},
						{Name: "Pitch to EDM blogs (EDM.com, Your EDM, Dancing Astronaut, etc.)", DaysOffset: days(-25)},
						{Name: "Pitch to YouTube channels (MrSuicideSheep, Trap Nation, etc.)", DaysOffset: days(-25)},
						{Name: "Pitch to DJ Mag / Mixmag / Decoded Magazine", DaysOffset: days(-25)},
						{Name: "Submit to DJ mix shows / radio (BBC Radio 1, Rinse FM, etc.)", DaysOffset: days(-21)},
						{Name: "Press follow ups (1 week after initial pitch)", DaysOffset: days(-18)},
						{Name: "Share and amplify all press coverage on socials", DaysOffset: days(0)},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research genre-specific Spotify playlists", DaysOffset: days(-35)},
						{Name: "Write playlist pitch", DaysOffset: days(-30)},
						{Name: "Pitch to independent curators", DaysOffset: days(-28)},
						{Name: "Submit via SubmitHub / PlaylistPush / Groover", DaysOffset: days(-28)},
						{Name: "Pitch to Apple Music / Amazon Music editorial", DaysOffset: days(-28)},
						{Name: "Update label-owned playlists with new release", DaysOffset: days(0)},
					},
				},
				{
					Name: "Community & DJ Network",
					Tasks: []TemplateTask{
						{Name: "Send to DJ friends for support", DaysOffset: days(-14)},
						{Name: "Post in producer Discord servers / Facebook groups", DaysOffset: days(0)},
						{Name: "Post on Reddit (r/EDM, r/electronicmusic, genre subs)", DaysOffset: days(0)},
						{Name: "Engage in comments on related posts", DaysOffset: days(1)},
						{Name: "Coordinate SoundCloud repost swaps with other artists/labels", DaysOffset: days(0)},
						{Name: "Cross-promote with other label artists", DaysOffset: days(0)},
					},
				},
				{
					Name: "Sync & Licensing",
					Tasks: []TemplateTask{
						{Name: "Register with sync platforms (Symphonic, Musicbed, Songtradr)", DaysOffset: days(-7)},
						{Name: "Prepare sync one-sheet (mood, tempo, comparable placements)", DaysOffset: days(-7)},
						{Name: "Pitch to gaming studios (racing, action, indie games)"},
						{Name: "Monitor sync briefs and respond to music supervisor requests"},
					},
				},
			},
		},
		{
			Name:  "Distribution",
			Color: "#EF4444",
			Groups: []TemplateGroup{
				{
					Name: "Platform Setup",
					Tasks: []TemplateTask{
						{Name: "Upload Spotify Canvas", DaysOffset: days(-7)},
						{Name: "Update Spotify artist profile (bio, images, Artist Pick)", DaysOffset: days(-7)},
						{Name: "Update Apple Music for Artists profile", DaysOffset: days(-7)},
						{Name: "Update Beatport artist/label profile", DaysOffset: days(-7)},
						{Name: "Upload to SoundCloud (release day)", DaysOffset: days(0)},
						{Name: "Upload to Bandcamp", Subtasks: []string{"Detailed liner notes & credits", "Offer FLAC + WAV + MP3", "Set pricing (name your price or fixed)", "Tag accurately for discovery"}, DaysOffset: days(0)},
						{Name: "Upload visualizer to YouTube", DaysOffset: days(0)},
					},
				},
				{
					Name: "Rights & Registration",
					Tasks: []TemplateTask{
						{Name: "Register with PROs (ASCAP/BMI/SESAC/PRS)", DaysOffset: days(-49)},
						{Name: "Register with SoundExchange", DaysOffset: days(-49)},
						{Name: "Register with MLC (mechanical royalties)", DaysOffset: days(-49)},
						{Name: "Obtain sample clearance (if applicable)", DaysOffset: days(-56)},
					},
				},
				{
					Name: "Beatport Strategy",
					Tasks: []TemplateTask{
						{Name: "Enroll in Beatport Hype program (if eligible)", DaysOffset: days(-56)},
						{Name: "Confirm Beatport Exclusive window is set", DaysOffset: days(-49)},
						{Name: "Coordinate social push directing fans to Beatport purchase", DaysOffset: days(0)},
						{Name: "Ask supporting DJs to purchase on Beatport", DaysOffset: days(0)},
						{Name: "Monitor Hype chart → main genre chart transition", DaysOffset: days(3)},
						{Name: "Screenshot chart positions for press kit", DaysOffset: days(7)},
					},
				},
			},
		},
	}
}
