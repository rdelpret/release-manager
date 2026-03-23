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
}

// DefaultTemplate returns the 5 pre-loaded task lists for a new campaign
func DefaultTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Audio Production",
					Tasks: []TemplateTask{
						{Name: "Final mixdown + master (WAV + MP3)"},
						{Name: "Extended mix"},
						{Name: "DJ-friendly intro/outro edit"},
						{Name: "Stems pack for remixers"},
						{Name: "Sync-ready edits (30s, 60s, instrumental)"},
						{Name: "Assign ISRC code to each version"},
						{Name: "Obtain UPC/EAN barcode"},
						{Name: "Lock metadata sheet (BPM, key, mood tags, subgenre, credits)"},
					},
				},
				{
					Name: "Visual Assets",
					Tasks: []TemplateTask{
						{Name: "Create asset folder"},
						{Name: "Final artwork (3000x3000)"},
						{Name: "Alternate artwork crops (square, banner, story-sized)"},
						{Name: "Spotify Canvas (3-8s looping video)"},
						{Name: "Visualizer video for YouTube"},
						{Name: "Music video (if applicable)"},
						{Name: "Artist press photos"},
						{Name: "Update EPK (bio, top tracks, press quotes, notable shows)"},
						{Name: "Promo content folder (clips, BTS, etc.)"},
					},
				},
				{
					Name: "Content Production",
					Tasks: []TemplateTask{
						{Name: "Studio session / production walkthrough clips"},
						{Name: "\"How I made this sound\" breakdown"},
						{Name: "DJ booth footage playing the track"},
						{Name: "Audiogram snippets (waveform + clip for stories)"},
						{Name: "Countdown assets (7 days, 3 days, 1 day, tonight, now)"},
						{Name: "Mood/aesthetic video matching the track's vibe"},
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
						{Name: "Campaign setup + timeline"},
						{Name: "Contracts / splits / publishing"},
						{Name: "Create content plan"},
						{Name: "Commission artwork / visual assets"},
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits & metadata", "Set release date", "Upload artwork", "Select stores & territories", "Tag BPM, key, subgenre accurately"}},
						{Name: "Set Beatport Exclusive window (2, 4, or 8 weeks)"},
						{Name: "Enable Beatport Pre-Order"},
						{Name: "Submit to Beatport editorial with track description"},
					},
				},
				{
					Name: "Pre-Release (4 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Submit to Spotify editorial via Spotify for Artists"},
						{Name: "Create pre-save link"},
						{Name: "Create smart link (Soundplate / Linkfire / Toneden)"},
						{Name: "Pre-save campaign / contest"},
						{Name: "Create or update Linktree"},
						{Name: "Write out hashtags"},
						{Name: "Update website - coming soon"},
						{Name: "Update social media bios - coming soon"},
						{Name: "Update Spotify / Apple Music artist profiles"},
					},
				},
				{
					Name: "Pre-Release (2-3 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Send track to DJ promo pools", Subtasks: []string{"Reaktion", "ZIPDJ", "PromoPush", "Music Worx"}},
						{Name: "Send private links to target DJs (20-50 in your subgenre)"},
						{Name: "Send to radio promo contacts"},
						{Name: "Collect and log DJ feedback / support quotes"},
						{Name: "Send pre-save email to mailing list"},
					},
				},
				{
					Name: "Release Day",
					Tasks: []TemplateTask{
						{Name: "Coordinated post across all platforms"},
						{Name: "Update all bios and link-in-bios to smart link"},
						{Name: "Upload to SoundCloud"},
						{Name: "Add to all label-owned Spotify playlists"},
						{Name: "Push fans to Beatport purchase link (purchases > streams for charting)"},
						{Name: "Send \"out now\" email to mailing list"},
						{Name: "Post on Bandcamp (time around Bandcamp Friday if possible)"},
					},
				},
				{
					Name: "Post-Release (weeks 1-4)",
					Tasks: []TemplateTask{
						{Name: "Monitor Spotify for Artists daily (streams, saves, playlist adds)"},
						{Name: "Monitor Beatport chart position and sales"},
						{Name: "Track SoundCloud plays, reposts, comments"},
						{Name: "Share press coverage on socials"},
						{Name: "Repost DJ support and fan reactions"},
						{Name: "Respond to all fan comments and messages"},
						{Name: "Compile campaign performance report"},
						{Name: "Campaign retrospective: what worked, what to improve"},
						{Name: "Archive campaign assets for future reference"},
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
						{Name: "Teaser clip (15-30s preview)"},
						{Name: "Artwork reveal post"},
						{Name: "Banner - release date + pre-save"},
						{Name: "Story-sized artwork (with release date)"},
						{Name: "Pre-save post"},
						{Name: "Studio session / production breakdown clip"},
						{Name: "Behind the track post (inspiration, story)"},
						{Name: "\"Midnight\" / \"Tomorrow\" countdown post"},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now"},
						{Name: "Track clip - out now"},
						{Name: "Artwork edit - out now"},
						{Name: "Story-sized out now graphic"},
						{Name: "Visualizer clip for socials"},
						{Name: "DJ set clip featuring the track"},
						{Name: "Studio breakdown / production tips reel"},
						{Name: "Credits / collab shoutout post"},
						{Name: "Fan reaction / UGC repost"},
					},
				},
				{
					Name: "TikTok / Reels / Shorts",
					Tasks: []TemplateTask{
						{Name: "TikTok with track (10-15s, hook-first)"},
						{Name: "Production breakdown reel (DAW screen recording)"},
						{Name: "\"How I made this sound\" short"},
						{Name: "DJ booth POV playing the track"},
						{Name: "Before/after sound design clip"},
						{Name: "Crowd reaction clip (if available)"},
						{Name: "Boost top-performing TikTok via Spark Ads (if organic traction)"},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "DJ mix snippet featuring the track"},
						{Name: "Live set / festival clip with track"},
						{Name: "Remix teasers (if applicable)"},
						{Name: "Playlist placement screenshot posts"},
						{Name: "Milestone celebration posts (streams, charts)"},
						{Name: "VIP / remix / bootleg version tease"},
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
						{Name: "Write press release"},
						{Name: "Build media outreach list (30-50 outlets)"},
						{Name: "Pitch exclusive premiere to top-choice outlet"},
						{Name: "Pitch to EDM blogs (EDM.com, Your EDM, Dancing Astronaut, etc.)"},
						{Name: "Pitch to YouTube channels (MrSuicideSheep, Trap Nation, etc.)"},
						{Name: "Pitch to DJ Mag / Mixmag / Decoded Magazine"},
						{Name: "Submit to DJ mix shows / radio (BBC Radio 1, Rinse FM, etc.)"},
						{Name: "Press follow ups (1 week after initial pitch)"},
						{Name: "Share and amplify all press coverage on socials"},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research genre-specific Spotify playlists"},
						{Name: "Write playlist pitch"},
						{Name: "Pitch to independent curators"},
						{Name: "Submit via SubmitHub / PlaylistPush / Groover"},
						{Name: "Pitch to Apple Music / Amazon Music editorial"},
						{Name: "Update label-owned playlists with new release"},
					},
				},
				{
					Name: "Community & DJ Network",
					Tasks: []TemplateTask{
						{Name: "Send to DJ friends for support"},
						{Name: "Post in producer Discord servers / Facebook groups"},
						{Name: "Post on Reddit (r/EDM, r/electronicmusic, genre subs)"},
						{Name: "Engage in comments on related posts"},
						{Name: "Coordinate SoundCloud repost swaps with other artists/labels"},
						{Name: "Cross-promote with other label artists"},
					},
				},
				{
					Name: "Sync & Licensing",
					Tasks: []TemplateTask{
						{Name: "Register with sync platforms (Symphonic, Musicbed, Songtradr)"},
						{Name: "Prepare sync one-sheet (mood, tempo, comparable placements)"},
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
						{Name: "Upload Spotify Canvas"},
						{Name: "Update Spotify artist profile (bio, images, Artist Pick)"},
						{Name: "Update Apple Music for Artists profile"},
						{Name: "Update Beatport artist/label profile"},
						{Name: "Upload to SoundCloud (release day)"},
						{Name: "Upload to Bandcamp", Subtasks: []string{"Detailed liner notes & credits", "Offer FLAC + WAV + MP3", "Set pricing (name your price or fixed)", "Tag accurately for discovery"}},
						{Name: "Upload visualizer to YouTube"},
					},
				},
				{
					Name: "Rights & Registration",
					Tasks: []TemplateTask{
						{Name: "Register with PROs (ASCAP/BMI/SESAC/PRS)"},
						{Name: "Register with SoundExchange"},
						{Name: "Register with MLC (mechanical royalties)"},
						{Name: "Obtain sample clearance (if applicable)"},
					},
				},
				{
					Name: "Beatport Strategy",
					Tasks: []TemplateTask{
						{Name: "Enroll in Beatport Hype program (if eligible)"},
						{Name: "Confirm Beatport Exclusive window is set"},
						{Name: "Coordinate social push directing fans to Beatport purchase"},
						{Name: "Ask supporting DJs to purchase on Beatport"},
						{Name: "Monitor Hype chart → main genre chart transition"},
						{Name: "Screenshot chart positions for press kit"},
					},
				},
			},
		},
	}
}
