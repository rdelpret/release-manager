package template

// LPEPTemplate returns the task lists for an LP/EP campaign.
// Longer timeline (10-12 weeks), more assets, per-track masters,
// album-specific marketing, and a Merch & Physical list.
func LPEPTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Audio Production",
					Tasks: []TemplateTask{
						{Name: "Final mixdown + master for each track (WAV + MP3)", DaysOffset: days(-84)},
						{Name: "Extended mixes for club tracks", DaysOffset: days(-84)},
						{Name: "DJ-friendly intro/outro edits", DaysOffset: days(-70)},
						{Name: "Stems pack for remixers", DaysOffset: days(-70)},
						{Name: "Sync-ready edits (30s, 60s, instrumental)", DaysOffset: days(-56)},
						{Name: "Assign ISRC code to each track/version", DaysOffset: days(-80)},
						{Name: "Obtain UPC/EAN barcode for the release", DaysOffset: days(-80)},
						{Name: "Lock metadata sheet per track (BPM, key, mood tags, subgenre, credits)", DaysOffset: days(-80)},
						{Name: "Finalize tracklist ordering", DaysOffset: days(-70)},
					},
				},
				{
					Name: "Visual Assets",
					Tasks: []TemplateTask{
						{Name: "Create asset folder", DaysOffset: days(-84)},
						{Name: "Final album artwork (3000x3000)", DaysOffset: days(-56)},
						{Name: "Artwork variants per single", DaysOffset: days(-49)},
						{Name: "Alternate artwork crops (square, banner, story-sized)", DaysOffset: days(-42)},
						{Name: "Spotify Canvas per track (3-8s looping video)", DaysOffset: days(-28)},
						{Name: "Album trailer / teaser video", DaysOffset: days(-28)},
						{Name: "Visualizer video for YouTube", DaysOffset: days(-14)},
						{Name: "Music video (lead single)", DaysOffset: days(-7)},
						{Name: "Artist press photos", DaysOffset: days(-56)},
						{Name: "Update EPK (bio, top tracks, press quotes, notable shows)", DaysOffset: days(-56)},
						{Name: "Promo content folder (clips, BTS, etc.)", DaysOffset: days(-42)},
					},
				},
				{
					Name: "Content Production",
					Tasks: []TemplateTask{
						{Name: "Studio session / production walkthrough clips", DaysOffset: days(-42)},
						{Name: "\"How I made this sound\" breakdown per track", DaysOffset: days(-35)},
						{Name: "Track-by-track preview series", DaysOffset: days(-28)},
						{Name: "DJ booth footage playing tracks from the EP/LP", DaysOffset: days(-21)},
						{Name: "Audiogram snippets (waveform + clip for stories)", DaysOffset: days(-28)},
						{Name: "Countdown assets (7 days, 3 days, 1 day, tonight, now)", DaysOffset: days(-10)},
						{Name: "Mood/aesthetic video matching the album's vibe", DaysOffset: days(-21)},
						{Name: "Listening party announcement content", DaysOffset: days(-14)},
					},
				},
			},
		},
		{
			Name:  "Campaign Tasklist",
			Color: "#22C55E",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release (10+ weeks out)",
					Tasks: []TemplateTask{
						{Name: "Campaign setup + timeline", DaysOffset: days(-84)},
						{Name: "Contracts / splits / publishing (all tracks)", DaysOffset: days(-84)},
						{Name: "Select lead single(s)", DaysOffset: days(-84)},
						{Name: "Plan single rollout timeline", DaysOffset: days(-80)},
						{Name: "Create content plan", DaysOffset: days(-80)},
						{Name: "Commission artwork / visual assets", DaysOffset: days(-80)},
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits & metadata per track", "Set release date", "Upload artwork", "Select stores & territories", "Tag BPM, key, subgenre accurately for each track", "Set track ordering", "Configure as album vs compilation"}, DaysOffset: days(-70)},
						{Name: "Set Beatport Exclusive window (2, 4, or 8 weeks)", DaysOffset: days(-70)},
						{Name: "Enable Beatport Pre-Order", DaysOffset: days(-70)},
						{Name: "Submit to Beatport editorial with album description", DaysOffset: days(-63)},
					},
				},
				{
					Name: "Pre-Release (6 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Submit to Spotify editorial via Spotify for Artists", DaysOffset: days(-42)},
						{Name: "Create pre-save link", DaysOffset: days(-42)},
						{Name: "Create smart link (Soundplate / Linkfire / Toneden)", DaysOffset: days(-42)},
						{Name: "Pre-save campaign / contest", DaysOffset: days(-42)},
						{Name: "Create or update Linktree", DaysOffset: days(-35)},
						{Name: "Write out hashtags", DaysOffset: days(-28)},
						{Name: "Update website - coming soon", DaysOffset: days(-28)},
						{Name: "Update social media bios - coming soon", DaysOffset: days(-28)},
						{Name: "Update Spotify / Apple Music artist profiles", DaysOffset: days(-28)},
					},
				},
				{
					Name: "Pre-Release (2-3 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Send album to DJ promo pools", Subtasks: []string{"Reaktion", "ZIPDJ", "PromoPush", "Music Worx"}, DaysOffset: days(-21)},
						{Name: "Send private links to target DJs (20-50 in your subgenre)", DaysOffset: days(-21)},
						{Name: "Send to radio promo contacts", DaysOffset: days(-21)},
						{Name: "Collect and log DJ feedback / support quotes", DaysOffset: days(-14)},
						{Name: "Send pre-save email to mailing list", DaysOffset: days(-14)},
						{Name: "Host listening party (Discord / IG Live / Twitch)", DaysOffset: days(-7)},
					},
				},
				{
					Name: "Release Day",
					Tasks: []TemplateTask{
						{Name: "Coordinated post across all platforms", DaysOffset: days(0)},
						{Name: "Update all bios and link-in-bios to smart link", DaysOffset: days(0)},
						{Name: "Upload all tracks to SoundCloud", DaysOffset: days(0)},
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
						{Name: "Compile campaign performance report", DaysOffset: days(28)},
						{Name: "Campaign retrospective: what worked, what to improve", DaysOffset: days(35)},
						{Name: "Archive campaign assets for future reference", DaysOffset: days(35)},
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
						{Name: "Teaser clip (15-30s preview of lead single)", DaysOffset: days(-28)},
						{Name: "Album artwork reveal post", DaysOffset: days(-21)},
						{Name: "Tracklist reveal post", DaysOffset: days(-18)},
						{Name: "Banner - release date + pre-save", DaysOffset: days(-14)},
						{Name: "Story-sized artwork (with release date)", DaysOffset: days(-14)},
						{Name: "Pre-save post", DaysOffset: days(-14)},
						{Name: "Track-by-track teaser series", DaysOffset: days(-10)},
						{Name: "Behind the album post (inspiration, story, concept)", DaysOffset: days(-7)},
						{Name: "\"Midnight\" / \"Tomorrow\" countdown post", DaysOffset: days(-1)},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now", DaysOffset: days(0)},
						{Name: "Track clip - out now (lead single)", DaysOffset: days(0)},
						{Name: "Artwork edit - out now", DaysOffset: days(1)},
						{Name: "Story-sized out now graphic", DaysOffset: days(1)},
						{Name: "Album trailer clip for socials", DaysOffset: days(2)},
						{Name: "DJ set clip featuring album tracks", DaysOffset: days(3)},
						{Name: "Studio breakdown / production tips reel", DaysOffset: days(5)},
						{Name: "Credits / collab shoutout post", DaysOffset: days(4)},
						{Name: "Fan reaction / UGC repost", DaysOffset: days(7)},
						{Name: "Listening party highlights post", DaysOffset: days(3)},
					},
				},
				{
					Name: "TikTok / Reels / Shorts",
					Tasks: []TemplateTask{
						{Name: "TikTok with lead single (10-15s, hook-first)", DaysOffset: days(-7)},
						{Name: "Production breakdown reel (DAW screen recording)", DaysOffset: days(-5)},
						{Name: "\"How I made this sound\" short", DaysOffset: days(-3)},
						{Name: "DJ booth POV playing album tracks", DaysOffset: days(1)},
						{Name: "Before/after sound design clip", DaysOffset: days(3)},
						{Name: "Crowd reaction clip (if available)", DaysOffset: days(7)},
						{Name: "Boost top-performing TikTok via Spark Ads (if organic traction)", DaysOffset: days(7)},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "DJ mix snippet featuring album tracks", DaysOffset: days(5)},
						{Name: "Live set / festival clip with album tracks", DaysOffset: days(10)},
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
						{Name: "Write press release (album angle)", DaysOffset: days(-56)},
						{Name: "Build media outreach list (30-50 outlets)", DaysOffset: days(-56)},
						{Name: "Pitch exclusive premiere to top-choice outlet", DaysOffset: days(-42)},
						{Name: "Pitch album review to EDM blogs (EDM.com, Your EDM, Dancing Astronaut, etc.)", DaysOffset: days(-35)},
						{Name: "Pitch to YouTube channels (MrSuicideSheep, Trap Nation, etc.)", DaysOffset: days(-35)},
						{Name: "Pitch to DJ Mag / Mixmag / Decoded Magazine for album review", DaysOffset: days(-35)},
						{Name: "Submit to DJ mix shows / radio (BBC Radio 1, Rinse FM, etc.)", DaysOffset: days(-28)},
						{Name: "Press follow ups (1 week after initial pitch)", DaysOffset: days(-28)},
						{Name: "Share and amplify all press coverage on socials", DaysOffset: days(0)},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research genre-specific Spotify playlists", DaysOffset: days(-49)},
						{Name: "Write playlist pitch (per single + album)", DaysOffset: days(-42)},
						{Name: "Pitch to independent curators", DaysOffset: days(-42)},
						{Name: "Submit via SubmitHub / PlaylistPush / Groover", DaysOffset: days(-42)},
						{Name: "Pitch to Apple Music / Amazon Music editorial", DaysOffset: days(-42)},
						{Name: "Update label-owned playlists with new release", DaysOffset: days(0)},
					},
				},
				{
					Name: "Community & DJ Network",
					Tasks: []TemplateTask{
						{Name: "Send to DJ friends for support", DaysOffset: days(-21)},
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
						{Name: "Register with sync platforms (Symphonic, Musicbed, Songtradr)", DaysOffset: days(-14)},
						{Name: "Prepare sync one-sheet per track (mood, tempo, comparable placements)", DaysOffset: days(-14)},
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
						{Name: "Upload Spotify Canvas for each track", DaysOffset: days(-7)},
						{Name: "Update Spotify artist profile (bio, images, Artist Pick)", DaysOffset: days(-7)},
						{Name: "Update Apple Music for Artists profile", DaysOffset: days(-7)},
						{Name: "Update Beatport artist/label profile", DaysOffset: days(-7)},
						{Name: "Upload all tracks to SoundCloud (release day)", DaysOffset: days(0)},
						{Name: "Upload to Bandcamp", Subtasks: []string{"Detailed liner notes & credits per track", "Offer FLAC + WAV + MP3", "Set pricing (name your price or fixed)", "Tag accurately for discovery", "Set track ordering"}, DaysOffset: days(0)},
						{Name: "Upload visualizer to YouTube", DaysOffset: days(0)},
					},
				},
				{
					Name: "Rights & Registration",
					Tasks: []TemplateTask{
						{Name: "Register all tracks with PROs (ASCAP/BMI/SESAC/PRS)", DaysOffset: days(-63)},
						{Name: "Register with SoundExchange", DaysOffset: days(-63)},
						{Name: "Register with MLC (mechanical royalties)", DaysOffset: days(-63)},
						{Name: "Obtain sample clearance for all tracks (if applicable)", DaysOffset: days(-84)},
					},
				},
				{
					Name: "Beatport Strategy",
					Tasks: []TemplateTask{
						{Name: "Enroll in Beatport Hype program (if eligible)", DaysOffset: days(-70)},
						{Name: "Confirm Beatport Exclusive window is set", DaysOffset: days(-63)},
						{Name: "Coordinate social push directing fans to Beatport purchase", DaysOffset: days(0)},
						{Name: "Ask supporting DJs to purchase on Beatport", DaysOffset: days(0)},
						{Name: "Monitor Hype chart → main genre chart transition", DaysOffset: days(3)},
						{Name: "Screenshot chart positions for press kit", DaysOffset: days(7)},
					},
				},
			},
		},
		{
			Name:  "Merch & Physical",
			Color: "#EC4899",
			Groups: []TemplateGroup{
				{
					Name: "Vinyl Pressing",
					Tasks: []TemplateTask{
						{Name: "Finalize vinyl tracklist + side splits", DaysOffset: days(-84)},
						{Name: "Vinyl mastering (separate from digital master)", DaysOffset: days(-80)},
						{Name: "Design vinyl label art + sleeve", DaysOffset: days(-70)},
						{Name: "Order test pressing", DaysOffset: days(-70)},
						{Name: "Approve test pressing", DaysOffset: days(-56)},
						{Name: "Place full vinyl order (16+ week lead time)", DaysOffset: days(-112)},
						{Name: "Coordinate vinyl release date with digital", DaysOffset: days(-56)},
					},
				},
				{
					Name: "Merch Design",
					Tasks: []TemplateTask{
						{Name: "Design album merch (t-shirts, hoodies)", DaysOffset: days(-56)},
						{Name: "Design limited edition / collector items", DaysOffset: days(-49)},
						{Name: "Order merch samples", DaysOffset: days(-42)},
						{Name: "Approve samples + place bulk order", DaysOffset: days(-35)},
						{Name: "Product photography for web store", DaysOffset: days(-21)},
					},
				},
				{
					Name: "Pre-Orders & Fulfillment",
					Tasks: []TemplateTask{
						{Name: "Set up pre-orders (Bandcamp + web store)", DaysOffset: days(-28)},
						{Name: "Bandcamp physical listing (vinyl + merch)", DaysOffset: days(-28)},
						{Name: "Promote pre-orders on socials", DaysOffset: days(-21)},
						{Name: "Coordinate fulfillment logistics", DaysOffset: days(-14)},
						{Name: "Ship pre-orders on release day", DaysOffset: days(0)},
						{Name: "Post unboxing / customer photos on socials", DaysOffset: days(7)},
					},
				},
			},
		},
	}
}
