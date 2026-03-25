package template

// SoundCloudFlipTemplate returns the task lists for a SoundCloud Flip campaign.
// Same social/PR campaign as Single minus distribution (no DSPs, just SoundCloud upload).
// Default schedule: 4 weeks (shorter turnaround).
func SoundCloudFlipTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Audio Production",
					Tasks: []TemplateTask{
						{Name: "Final mixdown + master (WAV + MP3)", DaysOffset: days(-28)},
						{Name: "Extended mix", DaysOffset: days(-28)},
						{Name: "DJ-friendly intro/outro edit", DaysOffset: days(-21)},
						{Name: "Stems pack for remixers", DaysOffset: days(-21)},
						{Name: "Lock metadata sheet (BPM, key, mood tags, subgenre, credits)", DaysOffset: days(-26)},
					},
				},
				{
					Name: "Visual Assets",
					Tasks: []TemplateTask{
						{Name: "Create asset folder", DaysOffset: days(-28)},
						{Name: "Final artwork (3000x3000)", DaysOffset: days(-21)},
						{Name: "Alternate artwork crops (square, banner, story-sized)", DaysOffset: days(-14)},
						{Name: "Visualizer video for YouTube", DaysOffset: days(-7)},
						{Name: "Artist press photos", DaysOffset: days(-21)},
						{Name: "Promo content folder (clips, BTS, etc.)", DaysOffset: days(-14)},
					},
				},
				{
					Name: "Content Production",
					Tasks: []TemplateTask{
						{Name: "Studio session / production walkthrough clips", DaysOffset: days(-14)},
						{Name: "\"How I made this sound\" breakdown", DaysOffset: days(-10)},
						{Name: "DJ booth footage playing the track", DaysOffset: days(-7)},
						{Name: "Audiogram snippets (waveform + clip for stories)", DaysOffset: days(-10)},
						{Name: "Countdown assets (3 days, 1 day, tonight, now)", DaysOffset: days(-5)},
					},
				},
			},
		},
		{
			Name:  "Campaign Tasklist",
			Color: "#22C55E",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release (4 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Campaign setup + timeline", DaysOffset: days(-28)},
						{Name: "Create content plan", DaysOffset: days(-28)},
						{Name: "Commission artwork / visual assets", DaysOffset: days(-28)},
						{Name: "Create smart link (Soundplate / Linkfire / Toneden)", DaysOffset: days(-14)},
						{Name: "Create or update Linktree", DaysOffset: days(-14)},
						{Name: "Write out hashtags", DaysOffset: days(-10)},
						{Name: "Update website - coming soon", DaysOffset: days(-10)},
						{Name: "Update social media bios - coming soon", DaysOffset: days(-10)},
					},
				},
				{
					Name: "Pre-Release (2 weeks out)",
					Tasks: []TemplateTask{
						{Name: "Send track to DJ promo pools", Subtasks: []string{"Reaktion", "ZIPDJ", "PromoPush", "Music Worx"}, DaysOffset: days(-14)},
						{Name: "Send private links to target DJs (20-50 in your subgenre)", DaysOffset: days(-14)},
						{Name: "Collect and log DJ feedback / support quotes", DaysOffset: days(-7)},
					},
				},
				{
					Name: "Release Day",
					Tasks: []TemplateTask{
						{Name: "Coordinated post across all platforms", DaysOffset: days(0)},
						{Name: "Update all bios and link-in-bios to smart link", DaysOffset: days(0)},
						{Name: "Upload to SoundCloud", DaysOffset: days(0)},
						{Name: "Upload visualizer to YouTube", DaysOffset: days(0)},
						{Name: "Send \"out now\" email to mailing list", DaysOffset: days(0)},
					},
				},
				{
					Name: "Post-Release (weeks 1-4)",
					Tasks: []TemplateTask{
						{Name: "Track SoundCloud plays, reposts, comments", DaysOffset: days(1)},
						{Name: "Share press coverage on socials", DaysOffset: days(3)},
						{Name: "Repost DJ support and fan reactions", DaysOffset: days(3)},
						{Name: "Respond to all fan comments and messages", DaysOffset: days(1)},
						{Name: "Compile campaign performance report", DaysOffset: days(14)},
						{Name: "Campaign retrospective: what worked, what to improve", DaysOffset: days(21)},
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
						{Name: "Teaser clip (15-30s preview)", DaysOffset: days(-10)},
						{Name: "Artwork reveal post", DaysOffset: days(-7)},
						{Name: "Banner - release date + link", DaysOffset: days(-7)},
						{Name: "Story-sized artwork (with release date)", DaysOffset: days(-7)},
						{Name: "Studio session / production breakdown clip", DaysOffset: days(-5)},
						{Name: "Behind the track post (inspiration, story)", DaysOffset: days(-3)},
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
						{Name: "Credits / collab shoutout post", DaysOffset: days(4)},
						{Name: "Fan reaction / UGC repost", DaysOffset: days(7)},
					},
				},
				{
					Name: "TikTok / Reels / Shorts",
					Tasks: []TemplateTask{
						{Name: "TikTok with track (10-15s, hook-first)", DaysOffset: days(-3)},
						{Name: "Production breakdown reel (DAW screen recording)", DaysOffset: days(-2)},
						{Name: "\"How I made this sound\" short", DaysOffset: days(-1)},
						{Name: "DJ booth POV playing the track", DaysOffset: days(1)},
						{Name: "Before/after sound design clip", DaysOffset: days(3)},
						{Name: "Crowd reaction clip (if available)", DaysOffset: days(7)},
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
						{Name: "Write press release", DaysOffset: days(-21)},
						{Name: "Build media outreach list (30-50 outlets)", DaysOffset: days(-21)},
						{Name: "Pitch exclusive premiere to top-choice outlet", DaysOffset: days(-14)},
						{Name: "Pitch to EDM blogs (EDM.com, Your EDM, Dancing Astronaut, etc.)", DaysOffset: days(-12)},
						{Name: "Pitch to YouTube channels (MrSuicideSheep, Trap Nation, etc.)", DaysOffset: days(-12)},
						{Name: "Press follow ups (1 week after initial pitch)", DaysOffset: days(-7)},
						{Name: "Share and amplify all press coverage on socials", DaysOffset: days(0)},
					},
				},
				{
					Name: "Community & DJ Network",
					Tasks: []TemplateTask{
						{Name: "Send to DJ friends for support", DaysOffset: days(-7)},
						{Name: "Post in producer Discord servers / Facebook groups", DaysOffset: days(0)},
						{Name: "Post on Reddit (r/EDM, r/electronicmusic, genre subs)", DaysOffset: days(0)},
						{Name: "Engage in comments on related posts", DaysOffset: days(1)},
						{Name: "Coordinate SoundCloud repost swaps with other artists/labels", DaysOffset: days(0)},
						{Name: "Cross-promote with other label artists", DaysOffset: days(0)},
					},
				},
			},
		},
	}
}
